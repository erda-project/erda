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
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/dbclient"
	"github.com/erda-project/erda/pkg/http/httputil"
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
	if !fakecluster && !req.Force {
		// project cluster refer check
		referred, err := c.bdl.ProjectClusterReferred(userid, orgid, req.ClusterName)
		if err != nil {
			status = dbclient.StatusTypeFailed
			logrus.Errorf("check project cluster refer info failed, orgid: %s, cluster_name: %s", orgid, req.ClusterName)
			return recordID, err
		}

		if referred {
			status = dbclient.StatusTypeFailed
			detail = "An existing project is using the cluster and cannot offline this cluster."
		}

		// runtime cluster refer check
		if status == dbclient.StatusTypeSuccess {
			referred, err := c.bdl.RuntimesClusterReferred(userid, orgid, req.ClusterName)
			if err != nil {
				status = dbclient.StatusTypeFailed
				logrus.Errorf("check runtime cluster refer info failed, orgid: %s, cluster_name: %s", orgid, req.ClusterName)
				return recordID, err
			}

			if referred {
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

		relations, err := c.bdl.ListOrgClusterRelation(userid, req.ClusterName)
		if err != nil {
			logrus.Errorf("list org cluster relation failed, cluster: %s, error: %v", req.ClusterName, err)
			return recordID, err
		}

		if len(relations) == 0 || req.Force {
			err = c.bdl.DeleteCluster(req.ClusterName, map[string][]string{httputil.InternalHeader: {"cmp"}})
			if err != nil {
				errstr := fmt.Sprintf("failed to delete cluster %s : %v", req.ClusterName, err)
				logrus.Errorf(errstr)
				err = errors.New(errstr)
				return recordID, err
			}

			// Delete accessKey
			if err = c.DeleteAccessKey(req.ClusterName); err != nil {
				errStr := fmt.Sprintf("failed to delete cluster access key, cluster: %v, err: %v", req.ClusterName, err)
				logrus.Error(errStr)
				return recordID, err
			}
		}
	}

	if req.OrgID <= 0 {
		return recordID, nil
	}

	recordID, err = createRecord(c.db, dbclient.Record{
		RecordType:  dbclient.RecordTypeOfflineEdgeCluster,
		UserID:      userid,
		OrgID:       orgid,
		ClusterName: req.ClusterName,
		Status:      status,
		Detail:      detail,
	})

	if err != nil {
		// ignore record update error
		logrus.Errorf("update cluster delete record failed, cluster: %s, error: %v", req.ClusterName, err)
	}

	logrus.Infof("detail: %v", detail)
	return recordID, nil
}

func createRecord(db *dbclient.DBClient, record dbclient.Record) (recordID uint64, err error) {
	return db.RecordsWriter().Create(&record)
}

func (c *Clusters) BatchOfflineEdgeCluster(req apistructs.BatchOfflineEdgeClusterRequest, userid string) error {
	for _, cluster := range req.Clusters {
		req := apistructs.OfflineEdgeClusterRequest{
			OrgID:       0,
			ClusterName: cluster,
			Force:       true,
		}
		_, err := c.OfflineEdgeCluster(req, "onlyYou", strconv.Itoa(int(req.OrgID)))
		if err != nil {
			err := fmt.Errorf("cluster offline failed, cluster: %s, error: %v", cluster, err)
			logrus.Errorf(err.Error())
			return err
		}
	}
	return nil
}
