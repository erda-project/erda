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

package discover

import (
	"os"

	"github.com/sirupsen/logrus"
)

// 定义各个服务地址的环境变量配置名字.
const (
	EnvEventBox       = "EVENTBOX_ADDR"
	EnvCMDB           = "CMDB_ADDR"
	EnvScheduler      = "SCHEDULER_ADDR"
	EnvDiceHub        = "DICEHUB_ADDR"
	EnvSoldier        = "SOLDIER_ADDR"
	EnvOrchestrator   = "ORCHESTRATOR_ADDR"
	EnvAddOnPlatform  = "ADDON_PLATFORM_ADDR"
	EnvGittar         = "GITTAR_ADDR"
	EnvGittarAdaptor  = "GITTAR_ADAPTOR_ADDR"
	EnvCollector      = "COLLECTOR_ADDR"
	EnvMonitor        = "MONITOR_ADDR"
	EnvPipeline       = "PIPELINE_ADDR"
	EnvHepa           = "HEPA_ADDR"
	EnvCMP            = "CMP_ADDR"
	EnvOpenapi        = "OPENAPI_ADDR"
	EnvKMS            = "KMS_ADDR"
	EnvQA             = "QA_ADDR"
	EnvAPIM           = "APIM_ADDR"
	EnvTMC            = "TMC_ADDR" // TODO REMOVE
	EnvMSP            = "MSP_ADDR"
	EnvUC             = "UC_ADDR"
	EnvClusterDialer  = "CLUSTER_DIALER_ADDR"
	EnvDOP            = "DOP_ADDR"
	EnvECP            = "ECP_ADDR"
	EnvClusterManager = "CLUSTER_MANAGER_ADDR"
	EnvCoreServices   = "CORE_SERVICES_ADDR"
)

// 定义各个服务的 k8s svc 名称
const (
	SvcEventBox       = "eventbox"
	SvcCMDB           = "cmdb"
	SvcScheduler      = "scheduler"
	SvcDiceHub        = "dicehub"
	SvcSoldier        = "soldier"
	SvcOrchestrator   = "orchestrator"
	SvcAddOnPlatform  = "addon-platform"
	SvcGittar         = "gittar"
	SvcGittarAdaptor  = "gittar-adaptor"
	SvcCollector      = "collector"
	SvcMonitor        = "monitor"
	SvcPipeline       = "pipeline"
	SvcHepa           = "hepa"
	SvcCMP            = "cmp"
	SvcOpenapi        = "openapi"
	SvcKMS            = "addon-kms"
	SvcQA             = "qa"
	SvcAPIM           = "apim"
	SvcTMC            = "tmc"
	SvcMSP            = "msp"
	SvcUC             = "uc"
	SvcClusterDialer  = "cluster-dialer"
	SvcDOP            = "dop"
	SvcECP            = "ecp"
	SvcClusterManager = "cluster-manager"
	SvcCoreServices   = "core-services"
)

func EventBox() string {
	return getURL(EnvEventBox, SvcEventBox)
}

func CMDB() string {
	return getURL(EnvCMDB, SvcCMDB)
}

func Scheduler() string {
	return getURL(EnvScheduler, SvcScheduler)
}

func DiceHub() string {
	return getURL(EnvDiceHub, SvcDiceHub)
}

func Soldier() string {
	return getURL(EnvSoldier, SvcSoldier)
}

func Orchestrator() string {
	return getURL(EnvOrchestrator, SvcOrchestrator)
}

func AddOnPlatform() string {
	return getURL(EnvAddOnPlatform, SvcAddOnPlatform)
}

func Gittar() string {
	return getURL(EnvGittar, SvcGittar)
}

func GittarAdaptor() string {
	return getURL(EnvGittarAdaptor, SvcGittarAdaptor)
}

func Collector() string {
	return getURL(EnvCollector, SvcCollector)
}

func Monitor() string {
	return getURL(EnvMonitor, SvcMonitor)
}

func Pipeline() string {
	return getURL(EnvPipeline, SvcPipeline)
}

func Hepa() string {
	return getURL(EnvHepa, SvcHepa)
}

func TMC() string {
	return getURL(EnvTMC, SvcTMC)
}

func MSP() string {
	return getURL(EnvMSP, SvcMSP)
}

func CMP() string {
	return getURL(EnvCMP, SvcCMP)
}

func Openapi() string {
	return getURL(EnvOpenapi, SvcOpenapi)
}

func KMS() string {
	return getURL(EnvKMS, SvcKMS)
}

func QA() string {
	return getURL(EnvQA, SvcQA)
}

func APIM() string {
	return getURL(EnvAPIM, SvcAPIM)
}

func UC() string {
	return getURL(EnvUC, SvcUC)
}

func ClusterDialer() string {
	return getURL(EnvClusterDialer, SvcClusterDialer)
}

func DOP() string {
	return getURL(EnvDOP, SvcDOP)
}

func CoreServices() string {
	return getURL(EnvCoreServices, SvcCoreServices)
}

func ECP() string {
	return getURL(EnvECP, SvcECP)
}

func ClusterManager() string {
	return getURL(EnvClusterManager, SvcClusterManager)
}

func getURL(envKey, srvName string) string {
	v := os.Getenv(envKey)
	if v != "" {
		return v
	}
	url, err := GetEndpoint(srvName)
	if err != nil {
		logrus.Infof("get endpoint failed, service name: %s, error: %s",
			srvName, err)
	}
	return url
}
