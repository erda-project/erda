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

package clusters

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
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
			Header("Internal-Client", "ops").
			Path("/api/projects/actions/refer-cluster").
			Param("cluster", req.ClusterName).Do().JSON(&projectRefer)
		if err != nil {
			errstr := fmt.Sprintf("failed to call cmdb /api/projects/actions/refer-cluster: %v", err)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
		if !resp.IsOK() || !projectRefer.Success {
			errstr := fmt.Sprintf("call cmdb /api/projects/actions/refer-cluster, statuscode: %d, resp: %+v", resp.StatusCode(), projectRefer)
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
				Header("Internal-Client", "ops").
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
		deletecluster := cmdbDeleteClusterResp{}
		resp, err := httpclient.New().Delete(discover.CMDB()).
			Header("Internal-Client", "ops").
			Path(fmt.Sprintf("/api/clusters/%s", req.ClusterName)).
			Do().JSON(&deletecluster)
		if err != nil {
			errstr := fmt.Sprintf("failed to call cmdb /api/clusters/%s : %v", req.ClusterName, err)
			logrus.Errorf(errstr)
			err := errors.New(errstr)
			return recordID, err
		}
		if !resp.IsOK() || !deletecluster.Success {
			errstr := fmt.Sprintf("call cmdb /api/clusters/%s, statuscode: %d, resp: %+v",
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
