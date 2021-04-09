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

// Package cluster impl cluster API
package cluster

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/jsonstore"
)

const (
	// clusterPrefix Is the path prefix of the cluster configuration information in etcd
	clusterPrefix = "/dice/scheduler/configs/cluster/"
	// edasExecutorPrefix It is the path prefix of edas cluster executor configuration information in etcd
	edasExecutorPrefix = "/dice/clustertoexecutor/"
	// clusterMarathonSuffix Suffix to identify marathon executor
	// e.g. /dice/scheduler/configs/cluster/terminus-y-service-marathon
	clusterMarathonSuffix = "-service-marathon"
	// clusterMetronomeSuffix Suffix to identify metronome executor
	// e.g. /dice/scheduler/configs/cluster/terminus-y-job-metronome
	clusterMetronomeSuffix = "-job-metronome"
	// clusterEdasSuffix Suffix to identify edas executor
	// e.g. /dice/scheduler/configs/cluster/bsd-edas-service-edas
	clusterEdasSuffix = "-service-edas"
	// clusterK8SSuffix Suffix that identifies the k8s service executor
	clusterK8SSuffix = "-service-k8s"
	// clusterK8SJobSuffix Suffix to identify k8s job executo
	clusterK8SJobSuffix = "-job-k8s"
	// clusterK8SFlinkSuffix Is the address suffix of k8s flink
	clusterK8SFlinkSuffix = "-flink-k8s"
	// clusterK8SSparkSuffix Is the address suffix of k8s spark
	clusterK8SSparkSuffix = "-spark-k8s"
	// jobKindFlink Identifies the flink type job
	jobKindFlink = "FLINK"
	// jobKindSpark Identifies spark type job
	jobKindSpark = "SPARK"
	// clusterTypeDcos Identify the dcos cluster type in the colony event
	clusterTypeDcos = "dcos"
	// clusterTypeK8S Identify the k8s cluster type in the colony event
	clusterTypeK8S = "k8s"
	// clusterTypeK8S Identify the edas cluster type in the colony event
	clusterTypeEdas = "edas"
	// clusterActionCreate Identifies the creation of an action in a colony event
	clusterActionCreate = "create"
	// clusterActionUpdate Identifies the update action in a colony event
	clusterActionUpdate = "update"
	// marathonAddrSuffix Is the marathon address suffix
	marathonAddrSuffix = "/service/marathon"
	// metronomeAddrSuffix Is the metronome address suffix
	metronomeAddrSuffix = "/service/metronome"
)

// ClusterInfo Cluster information, different from apistructs.ClusterInfo, this structure is used internally by cluster pkg
type ClusterInfo struct {
	// cluster name, e.g. "terminus-y"
	ClusterName string `json:"clusterName,omitempty"`
	// executor name, e.g. MARATHONFORTERMINUS
	ExecutorName string `json:"name,omitempty"`
	// executor type，e.g. MARATHON, METRONOME, K8S, EDAS
	Kind string `json:"kind,omitempty"`

	// options can include the following
	//"ADDR": "master.mesos/service/marathon",
	//"PREFIX": "/runtimes/v1",
	//"VERSION":"1.6.0",
	//"PUBLICIPGROUP":"external",
	//"ENABLETAG":"true",
	//"PRESERVEPROJECTS":"58"
	//"CA_CRT"
	//"CLIENT_CRT"
	//"CLIENT_KEY"
	Options     map[string]string `json:"options,omitempty"`
	OptionsPlus *conf.OptPlus     `json:"optionsPlus,omitempty"`
}

// Cluster clusterimpl interface
type Cluster interface {
	Hook(event *apistructs.ClusterEvent) error
}

// ClusterImpl Cluster interface implement
type ClusterImpl struct {
	js jsonstore.JsonStore
}

// NewClusterImpl create ClusterImpl
func NewClusterImpl(js jsonstore.JsonStore) Cluster {
	return &ClusterImpl{js}
}
