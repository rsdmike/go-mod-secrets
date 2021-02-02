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

package types

import (
	"testing"
)

func TestBuildUrl(t *testing.T) {
	cfgNoPath := SecretConfig{Host: "localhost", Port: 8080, Protocol: "http"}
	cfgWithPath := SecretConfig{Host: "localhost", Port: 8080, Protocol: "http", Path: "/ping"}
	cfgWithTrailingSlash := SecretConfig{Host: "localhost", Port: 8080, Protocol: "http", Path: "/api/v1/ping/"}

	tests := []struct {
		name string
		cfg  SecretConfig
		path string
	}{
		{"No Path", cfgNoPath, "http://localhost:8080"},
		{"With Path", cfgWithPath, "http://localhost:8080/ping"},
		{"With Trailing Slash", cfgWithTrailingSlash, "http://localhost:8080/api/v1/ping"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.cfg.BuildURL(tt.cfg.Path)
			if val != tt.path {
				t.Errorf("%s unexpected path %s", tt.name, val)
			}
		})
	}
}
