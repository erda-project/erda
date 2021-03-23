package apistructs

import (
	"fmt"

	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ETCD_ENDPOINTS          ClusterInfoMapKey = "ETCD_ENDPOINTS"          // k8s etcd的rs ip地址，逗号分割
	DICE_INSIDE             ClusterInfoMapKey = "DICE_INSIDE"             // bool值, true表示当前集群是离线部署的
	DICE_CLUSTER_TYPE       ClusterInfoMapKey = "DICE_CLUSTER_TYPE"       // 集群类型
	DICE_CLUSTER_NAME       ClusterInfoMapKey = "DICE_CLUSTER_NAME"       // 集群名
	DICE_VERSION            ClusterInfoMapKey = "DICE_VERSION"            // dice 版本
	DICE_ROOT_DOMAIN        ClusterInfoMapKey = "DICE_ROOT_DOMAIN"        // 集群泛域名
	DICE_IS_EDGE            ClusterInfoMapKey = "DICE_IS_EDGE"            // true表示当前集群是边缘集群
	DICE_STORAGE_MOUNTPOINT ClusterInfoMapKey = "DICE_STORAGE_MOUNTPOINT" // 网盘的挂载路径
	DICE_PROTOCOL           ClusterInfoMapKey = "DICE_PROTOCOL"           // 集群的入口协议，未开启https时值为http，开启https时，值为https
	DICE_HTTP_PORT          ClusterInfoMapKey = "DICE_HTTP_PORT"          // http协议使用的端口
	DICE_HTTPS_PORT         ClusterInfoMapKey = "DICE_HTTPS_PORT"         // https协议使用的端口
	MASTER_ADDR             ClusterInfoMapKey = "MASTER_ADDR"             // master的地址
	MASTER_VIP_ADDR         ClusterInfoMapKey = "MASTER_VIP_ADDR"         // master vip的地址
	MASTER_URLS             ClusterInfoMapKey = "MASTER_URLS"             // 带协议头的master地址
	LB_ADDR                 ClusterInfoMapKey = "LB_ADDR"                 // lb的地址
	LB_URLS                 ClusterInfoMapKey = "LB_URLS"                 // 带协议头的lb地址
	NEXUS_ADDR              ClusterInfoMapKey = "NEXUS_ADDR"              // nexus的地址
	NEXUS_USERNAME          ClusterInfoMapKey = "NEXUS_USERNAME"          // nexus 用户名
	NEXUS_PASSWORD          ClusterInfoMapKey = "NEXUS_PASSWORD"          // nexus的密码
	REGISTRY_ADDR           ClusterInfoMapKey = "REGISTRY_ADDR"           // registry的地址
	SOLDIER_ADDR            ClusterInfoMapKey = "SOLDIER_ADDR"            // soldier的地址
	NETPORTAL_URL           ClusterInfoMapKey = "NETPORTAL_URL"           // netportal的集群入口url
	EDASJOB_CLUSTER_NAME    ClusterInfoMapKey = "EDASJOB_CLUSTER_NAME"    // edas 集群可能会使用别的集群运行 JOB，若该字段为空，则说明使用本集群运行 JOB
	CLUSTER_DNS             ClusterInfoMapKey = "CLUSTER_DNS"             // k8s 或 dcos 内部域名服务器，逗号分隔
	ISTIO_ALIYUN            ClusterInfoMapKey = "ISTIO_ALIYUN"            // 是否用aliyn asm，true or false
	ISTIO_INSTALLED         ClusterInfoMapKey = "ISTIO_INSTALLED"         // 是否启用了 istio
	ISTIO_VERSION           ClusterInfoMapKey = "ISTIO_VERSION"           // istio 的版本
)

type ClusterInfoResponse struct {
	Header
	Data ClusterInfoData `json:"data"`
}

type ClusterInfoListResponse struct {
	Header
	Data ClusterInfoDataList `json:"data"`
}

type IstioInfo struct {
	Installed   bool
	Version     string
	IsAliyunASM bool
}

type ClusterInfoData map[ClusterInfoMapKey]string

type ClusterInfoDataList []ClusterInfoData

type ClusterInfoMapKey string

func (info ClusterInfoData) MustGet(key ClusterInfoMapKey) string {
	v := info.Get(key)
	if v == "" {
		panic(fmt.Sprintf("clusterInfo missing %s", key))
	}
	return v
}

func (info ClusterInfoData) Get(key ClusterInfoMapKey) string {
	return info[key]
}

// {DICE_PROTOCOL}://{组件名}.{DICE_ROOT_DOMAIN}:{DICE_HTTP_PORT or DICE_HTTPS_PORT}
func (info ClusterInfoData) MustGetPublicURL(component string) string {
	// default is `http`
	diceProtocol := "http"
	port := info.Get(DICE_HTTP_PORT)
	// if is `https`
	if info.DiceProtocolIsHTTPS() {
		diceProtocol = "https"
		port = info.MustGet(DICE_HTTPS_PORT)
	}
	return fmt.Sprintf("%s://%s.%s:%s",
		diceProtocol,
		component,
		info.MustGet(DICE_ROOT_DOMAIN),
		port,
	)
}

// DiceProtocolIsHTTPS 判断 DiceProtocol 是否支持 HTTPS
func (info ClusterInfoData) DiceProtocolIsHTTPS() bool {
	protocols := strutil.Split(info.MustGet(DICE_PROTOCOL), ",", true)
	for _, protocol := range protocols {
		if strutil.Equal(protocol, "https", true) {
			return true
		}
	}
	return false
}

func (info ClusterInfoData) IsK8S() bool {
	clusterType := info.MustGet(DICE_CLUSTER_TYPE)
	switch strutil.ToLower(clusterType) {
	case "kubernetes", "k8s":
		return true
	default:
		return false
	}
}

func (info ClusterInfoData) IsDCOS() bool {
	clusterType := info.MustGet(DICE_CLUSTER_TYPE)
	switch strutil.ToLower(clusterType) {
	case "dcos", "dc/os":
		return true
	default:
		return false
	}
}

func (info ClusterInfoData) IsEDAS() bool {
	clusterType := info.MustGet(DICE_CLUSTER_TYPE)
	return strutil.ToLower(clusterType) == "edas"
}

func (info ClusterInfoData) GetIstioInfo() IstioInfo {
	istioInfo := IstioInfo{}
	installed := info.Get(ISTIO_INSTALLED)
	if installed == "true" {
		istioInfo.Installed = true
	}
	isASM := info.Get(ISTIO_ALIYUN)
	if isASM == "true" {
		istioInfo.IsAliyunASM = true
	}
	istioInfo.Version = info.Get(ISTIO_VERSION)
	return istioInfo
}

func (info ClusterInfoData) GetApiServerUrl() string {
	netportalUrl := info.Get(NETPORTAL_URL)
	masterAddr := info.Get(MASTER_VIP_ADDR)
	if netportalUrl == "" {
		return masterAddr
	} else {
		return netportalUrl + "/" + masterAddr
	}
}

type ClusterResourceInfoResponse struct {
	Header
	Data ClusterResourceInfoData `json:"data"`
}

type ClusterResourceInfoData struct {
	CPUOverCommit        float64 `json:"cpuOverCommit"`
	ProdCPUOverCommit    float64 `json:"prodCpuOverCommit"`
	DevCPUOverCommit     float64 `json:"devCpuOverCommit"`
	TestCPUOverCommit    float64 `json:"testCpuOverCommit"`
	StagingCPUOverCommit float64 `json:"stagingCpuOverCommit"`
	ProdMEMOverCommit    float64 `json:"prodMemOverCommit"`
	DevMEMOverCommit     float64 `json:"devMemOverCommit"`
	TestMEMOverCommit    float64 `json:"testMemOverCommit"`
	StagingMEMOverCommit float64 `json:"stagingMemOverCommit"`

	Nodes map[string]*NodeResourceInfo `json:"nodes"`
}

type NodeResourceInfo struct {
	// only 'dice-' prefixed labels
	Labels []string `json:"labels"`
	// dcos, edas 缺少一些 label 或无法获取 label, 所以告诉上层忽略 labels
	IgnoreLabels bool `json:"ignoreLabels"`

	Ready bool `json:"ready"`

	CPUAllocatable float64 `json:"cpuAllocatable"`
	MemAllocatable int64   `json:"memAllocatable"`

	CPUReqsUsage float64 `json:"cpuRequestUsage"`
	MemReqsUsage int64   `json:"memRequestUsage"`

	CPULimitUsage float64 `json:"cpuLimitUsage"`
	MemLimitUsage int64   `json:"memLimitUsage"`
}
