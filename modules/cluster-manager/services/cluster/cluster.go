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

package cluster

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cluster-manager/dbclient"
	"github.com/erda-project/erda/modules/cluster-manager/model"
	"github.com/erda-project/erda/modules/cluster-manager/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

type Option func(*Cluster)

// Cluster cluster
type Cluster struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

// New new cluster
func New(options ...Option) *Cluster {
	cluster := &Cluster{}
	for _, op := range options {
		op(cluster)
	}
	return cluster
}

func WithDBClient(db *dbclient.DBClient) Option {
	return func(c *Cluster) {
		c.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(c *Cluster) {
		c.bdl = bdl
	}
}

// CreateWithEvent create cluster with event request
func (c *Cluster) CreateWithEvent(req *apistructs.ClusterCreateRequest) error {
	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return err
	}
	if cluster != nil {
		return nil
	}

	if err = c.Create(req); err != nil {
		return err
	}

	clusterInfo, err := c.GetCluster(req.Name)
	if err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.CreateAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderClusterManager,
		Content: clusterInfo,
	}

	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send cluster create event, (%v)", err)
		return nil
	}

	return nil
}

// Create create cluster
func (c *Cluster) Create(req *apistructs.ClusterCreateRequest) error {
	// validate param
	if err := c.checkCreateParam(req); err != nil {
		return err
	}

	if req.SysConfig != nil {
		switch req.Type {
		case apistructs.K8S:
			if req.SysConfig.MainPlatform != nil {
				req.SchedulerConfig.MasterURL = strutil.Concat("inet://", req.WildcardDomain,
					"/insecure-kubernetes.default.svc.cluster.local")
			} else {
				req.SchedulerConfig.MasterURL = "http://insecure-kubernetes.default.svc.cluster.local"
			}
		case apistructs.DCOS:
			if req.SysConfig.MainPlatform != nil {
				req.SchedulerConfig.MasterURL = strutil.Concat("inet://", req.WildcardDomain, "/master.mesos")
			} else {
				req.SchedulerConfig.MasterURL = "http://master.mesos"
			}
		}
	}

	// parse json store
	sysConfig, err := json.MarshalIndent(req.SysConfig, "", "\t")
	if err != nil {
		return err
	}

	schedulerConfig, err := json.MarshalIndent(req.SchedulerConfig, "", "\t")
	if err != nil {
		return err
	}

	opsConfig, err := json.MarshalIndent(req.OpsConfig, "", "\t")
	if err != nil {
		return err
	}

	manageConfig, err := json.MarshalIndent(req.ManageConfig, "", "\t")
	if err != nil {
		return err
	}

	cluster := &model.Cluster{
		OrgID:           req.OrgID,
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Type:            req.Type,
		CloudVendor:     req.CloudVendor,
		Logo:            req.Logo,
		WildcardDomain:  req.WildcardDomain,
		SysConfig:       string(sysConfig),
		SchedulerConfig: string(schedulerConfig),
		OpsConfig:       string(opsConfig),
		ManageConfig:    string(manageConfig),
	}

	if err = c.db.CreateCluster(cluster); err != nil {
		return err
	}

	return nil
}

// UpdateWithEvent update cluster & sender cluster update event
func (c *Cluster) UpdateWithEvent(req *apistructs.ClusterUpdateRequest) error {
	if err := c.Update(req); err != nil {
		return err
	}

	cluster, err := c.GetCluster(req.Name)
	if err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.UpdateAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderClusterManager,
		Content: cluster,
	}

	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send cluster update event, (%v)", err)
		return nil
	}

	return nil
}

// Update update cluster
func (c *Cluster) Update(req *apistructs.ClusterUpdateRequest) error {
	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return err
	}
	if cluster == nil {
		return errors.Errorf("not found")
	}
	logrus.Infof("before updated cluster info: %+v", cluster)

	cluster.DisplayName = req.DisplayName
	if req.Type != "" {
		cluster.Type = req.Type
	}
	cluster.Logo = req.Logo
	cluster.Description = req.Description
	if req.WildcardDomain != "" {
		cluster.WildcardDomain = req.WildcardDomain
	}

	if req.SchedulerConfig != nil {
		schedulerConfig, err := json.MarshalIndent(req.SchedulerConfig, "", "\t")
		if err != nil {
			return err
		}
		cluster.SchedulerConfig = string(schedulerConfig)
	}
	if req.OpsConfig != nil {
		opsConfig, err := json.MarshalIndent(req.OpsConfig, "", "\t")
		if err != nil {
			return err
		}
		cluster.OpsConfig = string(opsConfig)
	}
	if req.SysConfig != nil {
		var newSysConfig *apistructs.Sysconf
		if err := json.Unmarshal([]byte(cluster.SysConfig), &newSysConfig); err != nil {
			return err
		}
		// Check field which change disabled
		if err := c.diffSysConfig(req.SysConfig, newSysConfig); err != nil {
			return err
		}
		sysConfig, err := json.MarshalIndent(req.SysConfig, "", "\t")
		if err != nil {
			return err
		}
		cluster.SysConfig = string(sysConfig)
	}
	if req.ManageConfig != nil {
		manageConfig, err := json.MarshalIndent(req.ManageConfig, "", "\t")
		if err != nil {
			return err
		}
		cluster.ManageConfig = string(manageConfig)
	}
	if req.CloudVendor != "" {
		cluster.CloudVendor = req.CloudVendor
	}

	if err = c.db.UpdateCluster(cluster); err != nil {
		return err
	}

	return nil
}

// PatchWithEvent patch with event
func (c *Cluster) PatchWithEvent(req *apistructs.ClusterPatchRequest) error {
	cluster, err := c.db.GetClusterByName(req.Name)
	if err != nil {
		return err
	}
	if cluster == nil {
		return errors.Errorf("not found")
	}

	if req.ManageConfig == nil {
		return nil
	}

	cCluster := c.convert(cluster)

	if req.ManageConfig.CredentialSource == "" {
		req.ManageConfig.CredentialSource = cCluster.ManageConfig.CredentialSource
	}

	manageConfig, err := json.MarshalIndent(req.ManageConfig, "", "\t")
	if err != nil {
		return err
	}
	cluster.ManageConfig = string(manageConfig)

	if err = c.db.UpdateCluster(cluster); err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.UpdateAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderClusterManager,
		Content: c.convert(cluster),
	}

	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send cluster update event, (%v)", err)
		return nil
	}

	return nil
}

// DeleteWithEvent delete cluster with delete event
func (c *Cluster) DeleteWithEvent(clusterName string) error {
	cluster, err := c.GetCluster(clusterName)
	if err != nil {
		return apierrors.ErrDeleteCluster.InternalError(err)
	}
	if cluster == nil {
		return errors.Errorf("not found")
	}

	logrus.Infof("deleting cluster: %+v", *cluster)

	if err := c.DeleteByName(clusterName); err != nil {
		return apierrors.ErrDeleteCluster.InternalError(err)
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:     bundle.ClusterEvent,
			Action:    bundle.DeleteAction,
			OrgID:     "-1",
			ProjectID: "-1",
			TimeStamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderClusterManager,
		Content: cluster,
	}

	if err = c.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("[alert]failed to send cluster delete event, (%v)", err)
	}

	return nil
}

// DeleteByName delete cluster by name
func (c *Cluster) DeleteByName(clusterName string) error {
	return c.db.DeleteCluster(clusterName)
}

// GetCluster get cluster by name
func (c *Cluster) GetCluster(idOrName string) (*apistructs.ClusterInfo, error) {
	var cluster *model.Cluster
	clusterID, err := strutil.Atoi64(idOrName)
	if err == nil {
		cluster, err = c.db.GetClusterByID(clusterID)
	} else {
		cluster, err = c.db.GetClusterByName(idOrName)
	}
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.Errorf("not found")
	}

	return c.convert(cluster), nil
}

// ListCluster list all cluster
func (c *Cluster) ListCluster() (*[]apistructs.ClusterInfo, error) {
	clusters, err := c.db.ListCluster()
	if err != nil {
		return nil, err
	}
	clusterInfos := make([]apistructs.ClusterInfo, 0, len(*clusters))
	for i := range *clusters {
		item := c.convert(&(*clusters)[i])
		// TODO: Deprecated at future version.
		if item.SchedConfig != nil {
			item.SchedConfig.AuthPassword = ""
			item.SchedConfig.ClientCrt = ""
			item.SchedConfig.CACrt = ""
			item.SchedConfig.ClientKey = ""
			if item.SchedConfig.CPUSubscribeRatio == "" {
				item.SchedConfig.CPUSubscribeRatio = "1"
			}
		}

		clusterInfos = append(clusterInfos, *item)
	}

	return &clusterInfos, nil
}

func (c *Cluster) ListClusterByType(clusterType string) (*[]apistructs.ClusterInfo, error) {
	var clusterList []apistructs.ClusterInfo

	clusters, err := c.db.ListClusterByType(clusterType)
	if err != nil {
		return nil, err
	}

	clusterList = make([]apistructs.ClusterInfo, 0, len(*clusters))
	// TODO: Deprecated at future version.
	for i := range *clusters {
		item := c.convert(&(*clusters)[i])
		// TODO: Deprecated at future version.
		if item.SchedConfig != nil {
			item.SchedConfig.AuthPassword = ""
			item.SchedConfig.ClientCrt = ""
			item.SchedConfig.CACrt = ""
			item.SchedConfig.ClientKey = ""
			if item.SchedConfig.CPUSubscribeRatio == "" {
				item.SchedConfig.CPUSubscribeRatio = "1"
			}
		}

		clusterList = append(clusterList, *item)
	}

	return &clusterList, nil
}

// checkCreateParam check create param
func (c *Cluster) checkCreateParam(req *apistructs.ClusterCreateRequest) error {
	if req.Name == "" {
		return errors.Errorf("name is empty")
	}

	switch req.Type {
	case apistructs.DCOS, apistructs.EDAS, apistructs.K8S:
	default:
		return errors.Errorf("type is invalid")
	}
	return nil
}

func (c *Cluster) diffSysConfig(new, old *apistructs.Sysconf) error {
	if new.Cluster.Name != old.Cluster.Name {
		return errors.Errorf("cluster name mismatch")
	}
	if new.FPS.Host != old.FPS.Host {
		return errors.Errorf("fps host mismatch")
	}
	if new.Platform.WildcardDomain != old.Platform.WildcardDomain {
		return errors.Errorf("wildcard domain mismatch")
	}
	if new.Storage.MountPoint != old.Storage.MountPoint {
		return errors.New("nas mount point mismatch")
	}
	if new.Docker.DataRoot != old.Docker.DataRoot {
		return errors.New("docker data root mismatch")
	}
	newNodes := make(map[string]string, len(new.Nodes))
	for _, node := range new.Nodes {
		newNodes[node.IP] = node.Type
	}
	for _, node := range old.Nodes {
		if _, ok := newNodes[node.IP]; !ok {
			return errors.New("nodes info mismatch")
		}
	}
	return nil
}

// convert convert cluster model to ClusterInfo
func (c *Cluster) convert(cluster *model.Cluster) *apistructs.ClusterInfo {
	var (
		schedulerConfig *apistructs.ClusterSchedConfig
		opsConfig       *apistructs.OpsConfig
		manageConfig    *apistructs.ManageConfig
		sysConfig       *apistructs.Sysconf
		// Deprecated at 1.2
		urls = make(map[string]string)
	)

	if cluster.SysConfig != "" {
		if err := json.Unmarshal([]byte(cluster.SysConfig), &sysConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}

	if cluster.ManageConfig != "" {
		if err := json.Unmarshal([]byte(cluster.ManageConfig), &manageConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}

	if cluster.SchedulerConfig != "" {
		if err := json.Unmarshal([]byte(cluster.SchedulerConfig), &schedulerConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}
	if cluster.OpsConfig != "" {
		if err := json.Unmarshal([]byte(cluster.OpsConfig), &opsConfig); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}
	// TODO: Deprecated at 1.2, use for edas soldier 1.1 version
	if cluster.URLs != "" {
		if err := json.Unmarshal([]byte(cluster.URLs), &urls); err != nil {
			logrus.Warnf("failed to unmarshal, (%v)", err)
		}
	}

	return &apistructs.ClusterInfo{
		ID:             int(cluster.ID),
		Name:           cluster.Name,
		DisplayName:    cluster.DisplayName,
		Description:    cluster.Description,
		Type:           cluster.Type,
		CloudVendor:    cluster.CloudVendor,
		Logo:           cluster.Logo,
		WildcardDomain: cluster.WildcardDomain,
		SchedConfig:    schedulerConfig,
		System:         sysConfig,
		OpsConfig:      opsConfig,
		ManageConfig:   manageConfig,
		URLs:           urls,
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
	}
}
