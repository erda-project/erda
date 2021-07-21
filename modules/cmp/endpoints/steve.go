// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
		err := e.SteveAggregator.Add(&req.Content)
		if err != nil {
			logrus.Errorf("failed to start steve server for cluster %s, %v", req.Content.Name, err)
			return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, nil
		}
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
