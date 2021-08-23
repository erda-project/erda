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

package clusters

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type precheckResp struct {
	Success bool `json:"success"`
	Data    bool `json:"data"`
}
type cmdbDeleteClusterResp struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
}

func (c *Clusters) OfflineEdgeCluster(req apistructs.OfflineEdgeClusterRequest, userid string, orgid string) (uint64, error) {
	var recordID uint64
	var fakecluster bool
	clusterInfo, err := c.bdl.QueryClusterInfo(req.ClusterName)
	if err != nil {
		fakecluster = true
	} else {
		isEdgeCluster := clusterInfo.Get(apistructs.DICE_IS_EDGE)
		if isEdgeCluster != "true" {
			errstr := fmt.Sprintf("unsupport to offline non-edge cluster, clustername: %s", clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME))
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
	}
	status := dbclient.StatusTypeSuccess
	detail := ""
	if !fakecluster {
		// Check project whether to use cluster
		projectRefer := precheckResp{}
		resp, err := httpclient.New().Get(discover.CoreServices()).
			Header("Internal-Client", "cmp").
			Path("/api/projects/actions/refer-cluster").
			Param("cluster", req.ClusterName).Do().JSON(&projectRefer)
		if err != nil {
			errstr := fmt.Sprintf("failed to call core-services /api/projects/actions/refer-cluster: %v", err)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
		if !resp.IsOK() || !projectRefer.Success {
			errstr := fmt.Sprintf("call core-services /api/projects/actions/refer-cluster, statuscode: %d, resp: %+v", resp.StatusCode(), projectRefer)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
		if projectRefer.Data {
			status = dbclient.StatusTypeFailed
			detail = "An existing project is using the cluster and cannot offline this cluster."
		}

		if status == dbclient.StatusTypeSuccess {
			runtimeRefer := precheckResp{}
			resp, err := httpclient.New().Get(discover.Orchestrator()).
				Header("Internal-Client", "cmp").
				Path("/api/runtimes/actions/refer-cluster").
				Param("cluster", req.ClusterName).Do().JSON(&runtimeRefer)
			if err != nil {
				errstr := fmt.Sprintf("failed to call orch /api/runtimes/actions/refer-cluster: %v", err)
				logrus.Errorf(errstr)
				err := errors.New(errstr)
				return recordID, err
			}
			if !resp.IsOK() || !runtimeRefer.Success {
				errstr := fmt.Sprintf("call orch /api/runtimes/actions/refer-cluster, statuscode: %d, resp: %+v",
					resp.StatusCode(), runtimeRefer)
				logrus.Errorf(errstr)
				err := errors.New(errstr)
				return recordID, err
			}
			if runtimeRefer.Data {
				status = dbclient.StatusTypeFailed
				detail = "There are the Runtime (Addon) in the cluster, cannot offline this cluster"
			}
		}
	}
	// Offline cluster by call cmd /api/clusters/<clusterName>
	if status == dbclient.StatusTypeSuccess {
		if _, err = c.bdl.DereferenceCluster(req.OrgID, req.ClusterName, userid); err != nil {
			return recordID, err
		}

		deletecluster := cmdbDeleteClusterResp{}
		resp, err := httpclient.New().Delete(discover.ClusterManager()).
			Header("Internal-Client", "cmp").
			Path(fmt.Sprintf("/api/clusters/%s", req.ClusterName)).
			Do().JSON(&deletecluster)
		if err != nil {
			errstr := fmt.Sprintf("failed to call cluster-manager /api/clusters/%s : %v", req.ClusterName, err)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
		if !resp.IsOK() || !deletecluster.Success {
			errstr := fmt.Sprintf("call cluster-manager /api/clusters/%s, statuscode: %d, resp: %+v",
				req.ClusterName, resp.StatusCode(), deletecluster)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
	}

	recordID, err = c.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeOfflineEdgeCluster,
		UserID:      userid,
		OrgID:       orgid,
		ClusterName: req.ClusterName,
		Status:      status,
		Detail:      string(detail),
	})
	return recordID, nil
}
