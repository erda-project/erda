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
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cluster-manager/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetCluster get cluster meta info
func (e *Endpoints) GetCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	cluster, err := e.cluster.GetCluster(vars["idOrName"])
	if err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return apierrors.ErrGetCluster.NotFound().ToResp(), nil
		}
		return apierrors.ErrGetCluster.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(*cluster)
}

// ListCluster list all cluster
func (e *Endpoints) ListCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusters *[]apistructs.ClusterInfo
		err      error
	)

	clusterType := r.URL.Query().Get("clusterType")

	if clusterType != "" {
		clusters, err = e.cluster.ListClusterByType(clusterType)
	} else {
		clusters, err = e.cluster.ListCluster()
	}

	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(clusters)
}

// CreateCluster create cluster
func (e *Endpoints) CreateCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.ClusterCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateCluster.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	if err := e.cluster.CreateWithEvent(&req); err != nil {
		return apierrors.ErrCreateCluster.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

// UpdateCluster update cluster
func (e *Endpoints) UpdateCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.ClusterUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateCluster.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	if err := e.cluster.UpdateWithEvent(&req); err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return apierrors.ErrGetCluster.NotFound().ToResp(), nil
		}
		return apierrors.ErrUpdateCluster.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

// PatchCluster patch cluster
func (e *Endpoints) PatchCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.ClusterPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrPatchCluster.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	if err := e.cluster.PatchWithEvent(&req); err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return apierrors.ErrGetCluster.NotFound().ToResp(), nil
		}
		return apierrors.ErrPatchCluster.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}

// DeleteCluster delete cluster
func (e *Endpoints) DeleteCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	if err := e.cluster.DeleteWithEvent(vars["clusterName"]); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}
