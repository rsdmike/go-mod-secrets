/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package vault

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

func (c *Client) HealthCheck() (int, error) {
	code, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodGet,
		Path:                 HealthAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "health check",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// Heath check returns 5xx codes when unhealthy;
	// return error object only if we don't get numeric code back
	if code == 0 {
		return 0, err
	}

	c.lc.Infof("vault health check HTTP status: StatusCode: %d", code)
	return code, nil
}

func (c *Client) Init(secretThreshold int, secretShares int) (types.InitResponse, error) {
	c.lc.Infof("vault init strategy (SSS parameters): shares=%d threshold=%d",
		secretShares,
		secretThreshold)

	request := InitRequest{
		SecretShares:    secretShares,
		SecretThreshold: secretThreshold,
	}

	response := types.InitResponse{}
	_, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodPost,
		Path:                 InitAPI,
		JSONObject:           &request,
		BodyReader:           nil,
		OperationDescription: "initialize secret store",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	return response, err
}

func (c *Client) Unseal(keys []string, keysBase64 []string) error {
	c.lc.Infof("Vault unsealing Process. Applying key shares.")

	secretShares := len(keys)

	keyCounter := 1
	for _, key := range keysBase64 {
		request := UnsealRequest{Key: key}
		response := UnsealResponse{}

		_, err := c.doRequest(RequestArgs{
			AuthToken:            "",
			Method:               http.MethodPost,
			Path:                 UnsealAPI,
			JSONObject:           &request,
			BodyReader:           nil,
			OperationDescription: "unseal secret store",
			ExpectedStatusCode:   http.StatusOK,
			ResponseObject:       &response,
		})

		if err != nil {
			c.lc.Error(fmt.Sprintf("Error applying key share %d/%d: %s", keyCounter, secretShares, err.Error()))
			return err
		}

		c.lc.Info(fmt.Sprintf("Vault key share %d/%d successfully applied.", keyCounter, secretShares))
		if !response.Sealed {
			c.lc.Info("Vault key share threshold reached. Unsealing complete.")
			return nil
		}
		keyCounter++
	}

	return fmt.Errorf("%d", 1)
}

func (c *Client) InstallPolicy(token string, policyName string, policyDocument string) error {
	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPut,
		Path:                 fmt.Sprintf(CreatePolicyPath, url.PathEscape(policyName)),
		JSONObject:           UpdateACLPolicyRequest{Policy: policyDocument},
		BodyReader:           nil,
		OperationDescription: "install policy",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) EnableKVSecretEngine(token string, mountPoint string, kvVersion string) error {
	urlPath := path.Join(MountsAPI, mountPoint)
	parameters := EnableSecretsEngineRequest{
		Type:        KeyValue,
		Description: "key/value secret storage",
		Options: &SecretsEngineOptions{
			Version: kvVersion,
		},
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "update mounts",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) EnableConsulSecretEngine(token string, mountPoint string, defaultLeaseTTL string) error {
	urlPath := path.Join(MountsAPI, mountPoint)
	parameters := EnableSecretsEngineRequest{
		Type:        Consul,
		Description: "consul secret storage",
		Config: &SecretsEngineConfig{
			DefaultLeaseTTLDuration: defaultLeaseTTL,
		},
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "update mounts for Consul",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) CheckSecretEngineInstalled(token string, mountPoint string, engine string) (bool, error) {
	var response ListSecretEnginesResponse

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 MountsAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "query mounts for Consul",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if mountData := response.Data[mountPoint]; mountData.Type == engine {
		return true, nil
	}

	return false, err
}
