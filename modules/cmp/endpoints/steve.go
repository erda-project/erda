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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// clusterTypeK8S Identify the k8s cluster type in the colony event
	clusterTypeK8S = "k8s"
)

// SteveClusterHook starts steve server when create cluster and stop steve server when delete cluster
func (e *Endpoints) SteveClusterHook(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	if r.Body == nil {
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: "nil body"}, nil
	}
	var req apistructs.ClusterEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to decode clusterhook request: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: errstr}, nil
	}

	if !strutil.Equal(req.Content.Type, clusterTypeK8S, true) {
		return httpserver.HTTPResponse{Status: http.StatusOK}, nil
	}

	if strutil.Equal(req.Action, bundle.CreateAction, true) {
		e.SteveAggregator.Add(&req.Content)
	}

	if strutil.Equal(req.Action, bundle.DeleteAction, true) {
		err := e.SteveAggregator.Delete(req.Content.Name)
		if err != nil {
			logrus.Errorf("failed to stop steve server for cluster %s, %v", req.Content.Name, err)
			return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, nil
		}
	}

	return httpserver.HTTPResponse{Status: http.StatusOK}, nil
}
