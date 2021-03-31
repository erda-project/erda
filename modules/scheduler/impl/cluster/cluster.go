// Package cluster impl cluster API
package cluster

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/pkg/jsonstore"
)

const (
	// clusterPrefix 是集群配置信息在 etcd 中的路径前缀
	clusterPrefix = "/dice/scheduler/configs/cluster/"
	// edasExecutorPrefix 是edas集群executor配置信息在 etcd 中的路径前缀
	edasExecutorPrefix = "/dice/clustertoexecutor/"
	// clusterMarathonSuffix 标识 marathon executor 的后缀
	// e.g. /dice/scheduler/configs/cluster/terminus-y-service-marathon
	clusterMarathonSuffix = "-service-marathon"
	// clusterMetronomeSuffix 标识 metronome executor 的后缀
	// e.g. /dice/scheduler/configs/cluster/terminus-y-job-metronome
	clusterMetronomeSuffix = "-job-metronome"
	// clusterEdasSuffix 标识 edas executor 的后缀
	// e.g. /dice/scheduler/configs/cluster/bsd-edas-service-edas
	clusterEdasSuffix = "-service-edas"
	// clusterK8SSuffix 标识 k8s service executor 的后缀
	clusterK8SSuffix = "-service-k8s"
	// clusterK8SJobSuffix 标识 k8s job executor 的后缀
	clusterK8SJobSuffix = "-job-k8s"
	// clusterK8SFlinkSuffix 是 k8s flink 的地址后缀
	clusterK8SFlinkSuffix = "-flink-k8s"
	// clusterK8SSparkSuffix 是 k8s spark 的地址后缀
	clusterK8SSparkSuffix = "-spark-k8s"
	// jobKindFlink 标识 flink 类型 job
	jobKindFlink = "FLINK"
	// jobKindSpark 标识 spark 类型 job
	jobKindSpark = "SPARK"
	// clusterTypeDcos 标识 colony 事件中 dcos 集群类型
	clusterTypeDcos = "dcos"
	// clusterTypeK8S 标识 colony 事件中 k8s 集群类型
	clusterTypeK8S = "k8s"
	// clusterTypeK8S 标识 colony 事件中 edas 集群类型
	clusterTypeEdas = "edas"
	// clusterTypeLocaldocker 标识用本地 docker 的单机集群类型
	clusterTypeLocaldocker = "localdocker"
	// clusterActionCreate 标识 colony 事件中创建动作
	clusterActionCreate = "create"
	// clusterActionUpdate 标识 colony 事件中更新动作
	clusterActionUpdate = "update"
	// marathonAddrSuffix 是 marathon 地址后缀
	marathonAddrSuffix = "/service/marathon"
	// metronomeAddrSuffix 是 metronome 地址后缀
	metronomeAddrSuffix = "/service/metronome"
)

// ClusterInfo 集群信息，不同于 apistructs.ClusterInfo, 这个结构 cluster pkg 内部使用
type ClusterInfo struct {
	// 集群名称, e.g. "terminus-y"
	ClusterName string `json:"clusterName,omitempty"`
	// executor名称, e.g. MARATHONFORTERMINUS
	ExecutorName string `json:"name,omitempty"`
	// executor类型，对应plugins，e.g. MARATHON, METRONOME, K8S, EDAS
	Kind string `json:"kind,omitempty"`

	// options可包含如下
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

// Cluster clusterimpl 的接口
type Cluster interface {
	Hook(event *apistructs.ClusterEvent) error
}

// ClusterImpl Cluster interface 的实现
type ClusterImpl struct {
	js jsonstore.JsonStore
}

// NewClusterImpl 创建 ClusterImpl
func NewClusterImpl(js jsonstore.JsonStore) Cluster {
	return &ClusterImpl{js}
}
