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

package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// GetAllNamespaces get all namespaces in target project workspace
func (e *Endpoints) GetAllNamespaces(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	projectID := vars["projectID"]
	workspace := vars["workspace"]

	podsInfo, err := e.db.GetPodsByWorkspace(projectID, workspace)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err).ToResp(), nil
	}

	var namespaces []string
	for _, pod := range podsInfo {
		namespaces = append(namespaces, pod.K8sNamespace)
	}
	return httpserver.OkResp(namespaces)
}
