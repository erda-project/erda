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
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/envconf"

	"github.com/erda-project/erda/modules/ops/dbclient"
)

func (c *Clusters) AddClusters(req apistructs.CloudClusterRequest, userid string) (uint64, error) {
	var recordID uint64
	// load central cluster env
	err := envconf.Load(&req)

	logrus.Debugf("add cluster request:%+v", req)

	if err != nil {
		errstr := fmt.Sprint("failed to load DICE_CLUSTER_NAME")
		logrus.Errorf(errstr)
		return recordID, err
	}

	// check DICE_CLUSTER_NAME
	if isEmpty(req.CentralClusterName) {
		errstr := fmt.Sprint("DICE_CLUSTER_NAME is empty")
		logrus.Errorf(errstr)
		return recordID, errors.New(errstr)
	}

	if isEmpty(req.DisplayName) {
		req.DisplayName = req.ClusterName
	}

	clusterInfo, err := c.bdl.QueryClusterInfo(req.CentralClusterName)
	if err != nil {
		errstr := fmt.Sprintf("failed to query clusterinfo: %v, err: %v", clusterInfo, err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	if req.CollectorURL == "" {
		// protocol is http,https in gts dice-cluster-info (key: DICE_PROTOCOL), manual is http or https;
		if strings.Contains(req.CentralDiceProtocol, "https") {
			req.CentralDiceProtocol = "https"
		} else {
			req.CentralDiceProtocol = "http"
		}
		req.CollectorURL = fmt.Sprintf("%s://collector.%s", req.CentralDiceProtocol, req.CentralRootDomain)
	}
	if req.OpenAPI == "" {
		req.OpenAPI = fmt.Sprintf("%s://openapi.%s", req.CentralDiceProtocol, req.CentralRootDomain)
	}

	// check DICE_ROOT_DOMAIN
	if isEmpty(clusterInfo.Get(apistructs.DICE_ROOT_DOMAIN)) {
		errstr := fmt.Sprint("DICE_ROOT_DOMAIN is empty")
		logrus.Errorf(errstr)
		return recordID, errors.New(errstr)
	}

	if isEmpty(req.Terraform) {
		req.Terraform = "apply"
	}

	if !stringInSlice(string(req.CloudVendor), apistructs.CloudVendorSlice) {
		errstr := fmt.Sprintf("cloud vendor:%v is not valid", req.CloudVendor)
		logrus.Errorf(errstr)
		return recordID, errors.New(errstr)
	}

	strs := strings.Split(string(req.CloudVendor), "-")
	req.CloudVendorName, req.CloudBasicRsc = strs[0], strs[1]

	logrus.Debugf("cloud request: %v", req)

	var yml apistructs.PipelineYml
	var recordType dbclient.RecordType
	if req.CloudVendor == apistructs.CloudVendorAliEcs || req.InstallerIp != "" {
		// only alicloud-ecs need DockerCIDR
		if isEmpty(req.DockerCIDR) && req.CloudVendor == apistructs.CloudVendorAliEcs {
			errstr := "DockerCIDR is empty"
			logrus.Errorf(errstr)
			return recordID, err
		}
		if isEmpty(req.DockerBip) && req.CloudVendor == apistructs.CloudVendorAliEcs {
			strs := strings.Split(req.DockerCIDR, "/")
			ip := strs[0]
			mask := strs[1]
			bip := strings.Split(ip, ".")
			bip[3] = "1"
			dockerBip := strings.Join(bip, ".")
			req.DockerBip = strings.Join([]string{dockerBip, mask}, "/")
		}
		recordType = getRecordType(req.CloudVendor)
		yml = buildEcsPipeline(req)
	} else if req.CloudVendor == apistructs.CloudVendorAliAck { // TODO remove
		recordType = dbclient.RecordTypeAddAliACKECluster
		yml = buildCSPipeline(req)
	} else if req.CloudVendor == apistructs.CloudVendorAliCS {
		recordType = dbclient.RecordTypeAddAliCSECluster
		yml = buildCSPipeline(req)
	} else if req.CloudVendor == apistructs.CloudVendorAliCSManaged {
		recordType = dbclient.RecordTypeAddAliCSManagedCluster
		yml = buildCSPipeline(req)
	} else {
		errstr := fmt.Sprintf("cloud vendor:%v is not valid", req.CloudVendor)
		logrus.Errorf(errstr)
		return recordID, errors.New(errstr)
	}

	b, err := yaml.Marshal(yml)
	if err != nil {
		errstr := fmt.Sprintf("failed to marshal pipeline yml: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	dto, err := c.bdl.CreatePipeline(&apistructs.PipelineCreateRequestV2{
		PipelineYml:     string(b),
		PipelineYmlName: fmt.Sprintf("ops-add-cluster-%s.yml", req.ClusterName),
		ClusterName:     clusterInfo.MustGet(apistructs.DICE_CLUSTER_NAME),
		PipelineSource:  apistructs.PipelineSourceOps,
		AutoRunAtOnce:   true,
	})
	if err != nil {
		errstr := fmt.Sprintf("failed to create pipeline: %v", err)
		logrus.Errorf(errstr)
		return recordID, err
	}

	recordID, err = c.db.RecordsWriter().Create(&dbclient.Record{
		RecordType:  recordType,
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getRecordType(vendor apistructs.CloudVendor) dbclient.RecordType {
	switch vendor {
	case apistructs.CloudVendorAliAck:
		return dbclient.RecordTypeAddAliCSECluster
	case apistructs.CloudVendorAliCS:
		return dbclient.RecordTypeAddAliCSECluster
	case apistructs.CloudVendorAliCSManaged:
		return dbclient.RecordTypeAddAliCSManagedCluster
	default:
		return dbclient.RecordTypeAddAliECSECluster
	}
}

func buildEcsPipeline(req apistructs.CloudClusterRequest) apistructs.PipelineYml {
	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{{{
			Type:    "cluster-resource",
			Version: "1.0",
			Alias:   string(apistructs.NodePhasePlan),
			Params: map[string]interface{}{
				// base fields which create cluster
				"org_name":      req.OrgName,
				"dice_version":  req.DiceVersion,
				"cluster_name":  req.ClusterName,
				"root_domain":   req.RootDomain,
				"enable_https":  req.EnableHttps,
				"cluster_size":  req.ClusterSize,
				"nameservers":   req.Nameservers,
				"collector_url": req.CollectorURL,
				"open_api":      req.OpenAPI,

				// fields which create cluster on vpc
				"cloud_vendor":     req.CloudVendor,
				"region":           req.Region,
				"cluster_type":     req.ClusterType,
				"cluster_spec":     req.ClusterSpec,
				"charge_type":      req.ChargeType,
				"charge_period":    req.ChargePeriod,
				"ak":               req.AccessKey,
				"sk":               req.SecretKey,
				"vpc_id":           req.VpcID,
				"vpc_cidr":         req.VpcCIDR,
				"vswitch_id":       req.VSwitchID,
				"vswitch_cidr":     req.VSwitchCIDR,
				"nat_gateway_id":   req.NatGatewayID,
				"forward_table_id": req.ForwardTableID,
				"snat_table_id":    req.SnatTableID,

				// container service fields
				"service_cidr": req.ServiceCIDR,
				"pod_cidr":     req.PodCIDR,
				"docker_cidr":  req.DockerCIDR,
				"docker_bip":   req.DockerBip,
				"docker_root":  req.DockerRoot,
				"exec_root":    req.ExecRoot,

				// jump server field
				"installer_ip": req.InstallerIp,
				"user":         req.User,
				"password":     req.Password,
				"port":         req.Port,

				// host ip and disk info
				"host_ips": req.HostIps,
				"device":   req.Device,

				// nas/glusterfs fields
				"nas_domain":    req.NasDomain,
				"nas_path":      req.NasPath,
				"glusterfs_ips": req.GlusterfsIps,

				"terraform": "plan",
			},
		}}, {{
			Type:    "cluster-resource",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseBuyNode),
			Params: map[string]interface{}{
				"org_name":      req.OrgName,
				"dice_version":  req.DiceVersion,
				"cluster_name":  req.ClusterName,
				"root_domain":   req.RootDomain,
				"enable_https":  req.EnableHttps,
				"cluster_size":  req.ClusterSize,
				"nameservers":   req.Nameservers,
				"collector_url": req.CollectorURL,
				"open_api":      req.OpenAPI,

				"cloud_vendor":     req.CloudVendor,
				"region":           req.Region,
				"cluster_type":     req.ClusterType,
				"cluster_spec":     req.ClusterSpec,
				"charge_type":      req.ChargeType,
				"charge_period":    req.ChargePeriod,
				"ak":               req.AccessKey,
				"sk":               req.SecretKey,
				"vpc_id":           req.VpcID,
				"vpc_cidr":         req.VpcCIDR,
				"vswitch_id":       req.VSwitchID,
				"vswitch_cidr":     req.VSwitchCIDR,
				"nat_gateway_id":   req.NatGatewayID,
				"forward_table_id": req.ForwardTableID,
				"snat_table_id":    req.SnatTableID,

				"service_cidr": req.ServiceCIDR,
				"pod_cidr":     req.PodCIDR,
				"docker_cidr":  req.DockerCIDR,
				"docker_bip":   req.DockerBip,
				"docker_root":  req.DockerRoot,
				"exec_root":    req.ExecRoot,

				"installer_ip": req.InstallerIp,
				"user":         req.User,
				"password":     req.Password,
				"port":         req.Port,

				"host_ips": req.HostIps,
				"device":   req.Device,

				"nas_domain":    req.NasDomain,
				"nas_path":      req.NasPath,
				"glusterfs_ips": req.GlusterfsIps,

				"terraform": req.Terraform,
				// "terraform": "plan",
			},
		}}, {{
			Type:    "dice-install",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseInstall),
			Params: map[string]interface{}{
				"dice_version":  req.DiceVersion,
				"cluster_name":  req.ClusterName,
				"display_name":  req.DisplayName,
				"root_domain":   req.RootDomain,
				"org_id":        req.OrgID,
				"cloud_vendor":  req.CloudVendor,
				"region":        req.Region,
				"charge_type":   req.ChargeType,
				"charge_period": req.ChargePeriod,
				"ak":            req.AccessKey,
				"sk":            req.SecretKey,
			},
		}},
		},
	}
	return yml
}

func buildCSPipeline(req apistructs.CloudClusterRequest) apistructs.PipelineYml {
	managed := false
	worker_number := 0 // TODO: Get param from frontend
	if req.CloudVendor == apistructs.CloudVendorAliCSManaged {
		managed = true
		worker_number = 2
	}

	yml := apistructs.PipelineYml{
		Version: "1.1",
		Stages: [][]*apistructs.PipelineYmlAction{{{
			Type:    "cs-kubernetes",
			Version: "1.0",
			Alias:   string(apistructs.NodePhasePlan),
			Params: map[string]interface{}{
				"org_name":     req.OrgName,
				"root_domain":  req.RootDomain,
				"enable_https": req.EnableHttps,

				"cluster_name":      req.ClusterName,
				"access_key_id":     req.AccessKey,
				"access_key_secret": req.SecretKey,
				"region":            req.Region,
				"charge_type":       req.ChargeType,
				"charge_period":     req.ChargePeriod,
				"vpc_id":            req.VpcID,
				"vswitch_id":        req.VSwitchID,
				"vpc_subnet":        req.VpcCIDR,
				"vswitch_subnet":    req.VSwitchCIDR,
				"nat_gateway_id":    req.NatGatewayID,
				"forward_table_id":  req.ForwardTableID,
				"snat_table_id":     req.SnatTableID,
				"container_subnet":  req.PodCIDR,
				"vip_subnet":        req.ServiceCIDR,
				"worker_number":     worker_number,
				"ecs_instance_type": req.EcsInstType,
				"k8s_version":       req.K8sVersion,
				"managed":           managed,

				"collector_url":     req.CollectorURL,
				"openapi_url":       req.OpenAPI,
				"terraform_command": "plan",
			},
		}}, {{
			Type:    "cs-kubernetes",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseBuyNode),
			Params: map[string]interface{}{
				"org_name":     req.OrgName,
				"root_domain":  req.RootDomain,
				"enable_https": req.EnableHttps,

				"cluster_name":      req.ClusterName,
				"access_key_id":     req.AccessKey,
				"access_key_secret": req.SecretKey,
				"region":            req.Region,
				"charge_type":       req.ChargeType,
				"charge_period":     req.ChargePeriod,
				"vpc_subnet":        req.VpcCIDR,
				"vpc_id":            req.VpcID,
				"vswitch_id":        req.VSwitchID,
				"vswitch_subnet":    req.VSwitchCIDR,
				"nat_gateway_id":    req.NatGatewayID,
				"forward_table_id":  req.ForwardTableID,
				"snat_table_id":     req.SnatTableID,
				"container_subnet":  req.PodCIDR,
				"vip_subnet":        req.ServiceCIDR,
				"worker_number":     worker_number,
				"ecs_instance_type": req.EcsInstType,
				"k8s_version":       req.K8sVersion,
				"managed":           managed,

				"collector_url":     req.CollectorURL,
				"openapi_url":       req.OpenAPI,
				"terraform_command": req.Terraform,
				// "terraform_command": "plan",
			},
		}}, {{
			Type:    "dice-install",
			Version: "1.0",
			Alias:   string(apistructs.NodePhaseInstall),
			Params: map[string]interface{}{
				"dice_version":  req.DiceVersion,
				"cluster_name":  req.ClusterName,
				"display_name":  req.DisplayName,
				"root_domain":   req.RootDomain,
				"org_id":        req.OrgID,
				"cloud_vendor":  req.CloudVendor,
				"region":        req.Region,
				"charge_type":   req.ChargeType,
				"charge_period": req.ChargePeriod,
				"ak":            req.AccessKey,
				"sk":            req.SecretKey,
			},
		}},
		},
	}
	return yml
}
