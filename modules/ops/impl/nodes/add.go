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

package nodes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/crypto/encrypt"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/strutil"

	"github.com/erda-project/erda/modules/ops/dbclient"
)

type Nodes struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

func New(db *dbclient.DBClient, bdl *bundle.Bundle) *Nodes {
	return &Nodes{db: db, bdl: bdl}
}

func (n *Nodes) AddNodes(req apistructs.AddNodesRequest, userid string) (uint64, error) {
	var recordID uint64
	clusterInfo, err := n.bdl.QueryClusterInfo(req.ClusterName)
	if err != nil {
		errstr := fmt.Sprintf("failed to queryclusterinfo: %v, clusterinfo: %v", err, clusterInfo)
		logrus.Errorf(errstr)
		return recordID, err
	}
	if !clusterInfo.IsK8S() {
		errstr := fmt.Sprintf("unsupported cluster type, cluster name: %s, cluster type: %s",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), clusterInfo.MustGet(apistructs.DICE_CLUSTER_TYPE))
		logrus.Errorf(errstr)
		err := errors.New(errstr)
		return recordID, err
	}

	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{{{
			Type:    "add-nodes",
			Version: "1.0",
			Params: map[string]interface{}{
				"hosts":             strutil.Join(req.Hosts, ",", true),
				"labels":            strutil.Join(req.Labels, ",", true),
				"port":              req.Port,
				"user":              req.User,
				"password":          req.Password,
				"sudo_has_password": req.SudoHasPassword,
				"data_disk_device":  req.DataDiskDevice,
			},
		}}},
	}
	b, err := yaml.Marshal(yml)
	if err != nil {
		errstr := fmt.Sprintf("failed to marshal pipelineyml: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	dto, err := n.bdl.CreatePipeline(&apistructs.PipelineCreateRequestV2{
		PipelineYml: string(b),
		PipelineYmlName: fmt.Sprintf("ops-add-nodes-%s-%s.yml",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), uuid.UUID()[:12]),
		ClusterName:    clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME),
		PipelineSource: apistructs.PipelineSourceOps,
		AutoRunAtOnce:  true,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to createpipeline: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	recordType := dbclient.RecordTypeAddNodes
	detail := ""
	if req.Source == apistructs.AddNodesEssSource {
		recordType = dbclient.RecordTypeAddEssNodes
		detail = req.Detail
	}

	recordID, err = n.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  recordType,
		UserID:      userid,
		OrgID:       strconv.FormatUint(req.OrgID, 10),
		ClusterName: req.ClusterName,
		Status:      dbclient.StatusTypeProcessing,
		Detail:      detail,
		PipelineID:  dto.ID,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to create record: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}
	return recordID, nil
}

func (n *Nodes) AddCSNodes(req apistructs.CloudNodesRequest, userid string) (uint64, error) {
	var recordID uint64
	clusterInfo, err := n.bdl.QueryClusterInfo(req.ClusterName)
	if err != nil {
		err = fmt.Errorf("failed to queryclusterinfo: %v, clusterinfo: %v", err, clusterInfo)
		logrus.Errorln(err)
		return recordID, err
	}
	if !clusterInfo.IsK8S() {
		err = fmt.Errorf("unsupported cluster type, cluster name: %s, cluster type: %s",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), clusterInfo.MustGet(apistructs.DICE_CLUSTER_TYPE))
		logrus.Errorln(err)
		return recordID, err
	}

	clusterConf, err := n.bdl.GetCluster(req.ClusterName)
	if err != nil {
		err = fmt.Errorf("failed to query cluster cofig form cmdb: %v, cluster config: %v", err, clusterConf)
		logrus.Errorln(err)
		return recordID, err
	}
	opsConfig := clusterConf.OpsConfig
	if opsConfig == nil {
		err = fmt.Errorf("empty ops_config, cluster config: %v", clusterConf)
		logrus.Errorln(err)
		return recordID, err
	}

	clusterID := opsConfig.Extra["cluster_id"]
	masterNumber := opsConfig.Extra["master_number"]

	if opsConfig.Region == "" || opsConfig.AccessKey == "" ||
		opsConfig.SecretKey == "" || opsConfig.EcsPassword == "" ||
		clusterID == "" || masterNumber == "" {
		err = fmt.Errorf("invalide cluster ops config, err: %v, config: %v", err, opsConfig)
		logrus.Errorln(err)
		return recordID, err
	}

	req.Region = opsConfig.Region
	req.AccessKey = encrypt.AesDecrypt(opsConfig.AccessKey, apistructs.TerraformEcyKey)
	req.SecretKey = encrypt.AesDecrypt(opsConfig.SecretKey, apistructs.TerraformEcyKey)
	req.InstancePassword = encrypt.AesDecrypt(opsConfig.EcsPassword, apistructs.TerraformEcyKey)

	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{{{
			Type:    "cs-kubernetes-scale",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseBuyNode),
			Params: map[string]interface{}{
				"cluster_name":      req.ClusterName,
				"access_key_id":     req.AccessKey,
				"access_key_secret": req.SecretKey,
				"region":            req.Region,
				"ssh_password":      req.InstancePassword,
				"worker_number":     req.InstanceNum,
				"cluster_id":        clusterID,
				"master_number":     masterNumber,
			},
		}}, {{
			Type:    "add-nodes",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseAddNode),
			Params: map[string]interface{}{
				"labels":        strutil.Join(req.Labels, ",", true),
				"port":          22,
				"user":          "root",
				"password":      req.InstancePassword,
				"worker_number": req.InstanceNum,
				"cluster_id":    clusterID,
				"master_number": masterNumber,
				"cs":            true,
			},
		}},
		},
	}

	logrus.Debugf("cloud-nodes pipeline yml: %v", yml)
	b, err := yaml.Marshal(yml)
	if err != nil {
		err = fmt.Errorf("failed to marshal pipelineyml: %v", err)
		logrus.Errorln(err)
		return recordID, err
	}

	dto, err := n.bdl.CreatePipeline(&apistructs.PipelineCreateRequestV2{
		PipelineYml: string(b),
		PipelineYmlName: fmt.Sprintf("ops-cloud-nodes-%s-%s.yml",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), uuid.UUID()[:12]),
		ClusterName:    clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME),
		PipelineSource: apistructs.PipelineSourceOps,
		AutoRunAtOnce:  true,
	})
	if err != nil {
		err = fmt.Errorf("failed to createpipeline: %v", err)
		logrus.Errorln(err)
		return recordID, err
	}

	recordID, err = n.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeAddAliNodes,
		UserID:      userid,
		OrgID:       strconv.FormatUint(req.OrgID, 10),
		ClusterName: req.ClusterName,
		Status:      dbclient.StatusTypeProcessing,
		Detail:      "",
		PipelineID:  dto.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create record: %v", err)
		logrus.Errorln(err)
		return recordID, err
	}
	return recordID, nil
}

func (n *Nodes) AddCloudNodes(req apistructs.CloudNodesRequest, userid string) (uint64, error) {
	var recordID uint64
	clusterInfo, err := n.bdl.QueryClusterInfo(req.ClusterName)
	if err != nil {
		errstr := fmt.Sprintf("failed to queryclusterinfo: %v, clusterinfo: %v", err, clusterInfo)
		logrus.Errorf(errstr)
		return recordID, err
	}
	if !clusterInfo.IsK8S() {
		errstr := fmt.Sprintf("unsupported cluster type, cluster name: %s, cluster type: %s",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), clusterInfo.MustGet(apistructs.DICE_CLUSTER_TYPE))
		logrus.Errorf(errstr)
		err := errors.New(errstr)
		return recordID, err
	}

	clusterConf, err := n.bdl.GetCluster(req.ClusterName)
	if err != nil {
		errstr := fmt.Sprintf("failed to query cluster cofig form cmdb: %v, cluster config: %v", err, clusterConf)
		logrus.Errorf(errstr)
		return recordID, err
	}
	opsConfig := clusterConf.OpsConfig
	if opsConfig == nil {
		err := fmt.Errorf("empty ops_config, cluster name: %v", req.ClusterName)
		logrus.Errorf(err.Error())
		return recordID, err
	}

	if isEmpty(opsConfig.VSwitchIDs) || isEmpty(opsConfig.AvailabilityZones) || isEmpty(opsConfig.AccessKey) ||
		isEmpty(opsConfig.SecretKey) || isEmpty(opsConfig.EcsPassword) || isEmpty(opsConfig.SgIDs) {
		err := fmt.Errorf("invalide ops_config, cluster name: %v", req.ClusterName)
		logrus.Errorf(err.Error())
		return recordID, err
	}

	req.AvailabilityZone = strings.Split(opsConfig.AvailabilityZones, ",")[0]
	req.VSwitchId = strings.Split(opsConfig.VSwitchIDs, ",")[0]
	req.SecurityGroupIds = strings.Split(opsConfig.SgIDs, ",")
	req.ChargeType = opsConfig.ChargeType
	if req.ChargeType == apistructs.PrePaidChargeType {
		req.ChargePeriod = opsConfig.ChargePeriod
	}
	req.AccessKey = encrypt.AesDecrypt(opsConfig.AccessKey, apistructs.TerraformEcyKey)
	req.SecretKey = encrypt.AesDecrypt(opsConfig.SecretKey, apistructs.TerraformEcyKey)
	req.InstancePassword = encrypt.AesDecrypt(opsConfig.EcsPassword, apistructs.TerraformEcyKey)

	if len(req.Region) == 0 {
		x := strings.Split(req.AvailabilityZone, "-")
		req.Region = strings.Join(x[:len(x)-1], "-")
	}

	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{{{
			Type:    "cloud-resource",
			Version: "1.0",
			Alias:   string(apistructs.NodePhasePlan),
			Params: map[string]interface{}{
				"cloud_vendor":      req.CloudVendor,
				"region":            req.Region,
				"availability_zone": req.AvailabilityZone,
				"charge_type":       req.ChargeType,
				"charge_period":     req.ChargePeriod,
				"ak":                req.AccessKey,
				"sk":                req.SecretKey,
				"cluster_name":      req.ClusterName,
				"cloud_resource":    req.CloudResource,
				"ecs_num":           req.InstanceNum,
				"ecs_password":      req.InstancePassword,
				"ecs_type":          req.InstanceType,
				"vswitch_id":        req.VSwitchId,
				"sg_ids":            strutil.Join(req.SecurityGroupIds, ",", true),
				"disk_type":         req.DiskType,
				"disk_size":         req.DiskSize,
				"terraform":         "plan",
			},
		}}, {{
			Type:    "cloud-resource",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseBuyNode),
			Params: map[string]interface{}{
				"cloud_vendor":      req.CloudVendor,
				"region":            req.Region,
				"availability_zone": req.AvailabilityZone,
				"charge_type":       req.ChargeType,
				"charge_period":     req.ChargePeriod,
				"ak":                req.AccessKey,
				"sk":                req.SecretKey,
				"cluster_name":      req.ClusterName,
				"cloud_resource":    req.CloudResource,
				"ecs_num":           req.InstanceNum,
				"ecs_password":      req.InstancePassword,
				"ecs_type":          req.InstanceType,
				"vswitch_id":        req.VSwitchId,
				"sg_ids":            strutil.Join(req.SecurityGroupIds, ",", true),
				"disk_type":         req.DiskType,
				"disk_size":         req.DiskSize,
				"terraform":         "apply",
			},
		}}, {{
			Type:    "add-nodes",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseAddNode),
			Params: map[string]interface{}{
				"labels":         strutil.Join(req.Labels, ",", true),
				"port":           22,
				"user":           "root",
				"password":       req.InstancePassword,
				"cloud_resource": "${" + string(apistructs.NodePhaseBuyNode) + "}" + "/cloud_resource",
			},
		}},
		},
	}

	logrus.Debugf("cloud-nodes pipeline yml: %v", yml)
	b, err := yaml.Marshal(yml)
	if err != nil {
		errstr := fmt.Sprintf("failed to marshal pipelineyml: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	dto, err := n.bdl.CreatePipeline(&apistructs.PipelineCreateRequestV2{
		PipelineYml: string(b),
		PipelineYmlName: fmt.Sprintf("ops-cloud-nodes-%s-%s.yml",
			clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME), uuid.UUID()[:12]),
		ClusterName:    clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME),
		PipelineSource: apistructs.PipelineSourceOps,
		AutoRunAtOnce:  true,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to createpipeline: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	recordID, err = n.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  dbclient.RecordTypeAddAliNodes,
		UserID:      userid,
		OrgID:       strconv.FormatUint(req.OrgID, 10),
		ClusterName: req.ClusterName,
		Status:      dbclient.StatusTypeProcessing,
		Detail:      "",
		PipelineID:  dto.ID,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to create record: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}
	return recordID, nil
}

func isEmpty(str string) bool {
	return strings.Replace(str, " ", "", -1) == ""
}
