// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import "net/http"

func Authorizer(req *http.Request) (string, bool, error) {
	// inner proxy not need auth
	if req.URL.Path == "/clusterdialer" {
		return "proxy", true, nil
	}
	clusterKey := req.Header.Get("X-Erda-Cluster-Key")
	// TODO: support openapi auth
	auth := req.Header.Get("Authorization")
	return clusterKey, auth != "", nil
}
