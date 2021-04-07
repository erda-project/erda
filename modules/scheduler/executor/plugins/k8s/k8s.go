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

// Package k8s implements managing the servicegroup by k8s cluster
package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/mohae/deepcopy"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	eventboxapi "github.com/erda-project/erda/modules/scheduler/events"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon/daemonset"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon/elasticsearch"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon/mysql"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon/redis"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/clusterinfo"
	ds "github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/daemonset"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/deployment"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/event"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/ingress"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/instanceinfosync"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8sservice"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/nodelabel"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/persistentvolume"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/pod"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/resourceinfo"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/serviceaccount"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/statefulset"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/cpupolicy"
	"github.com/erda-project/erda/pkg/dbengine"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind = "K8S"

	// SUBSCRIBE_RATIO_SUFFIX 超分比的key后缀
	SUBSCRIBE_RATIO_SUFFIX = "_SUBSCRIBE_RATIO"

	// CPU_NUM_QUOTA cpu limit key
	CPU_NUM_QUOTA = "CPU_NUM_QUOTA"
	// CPU_CFS_PERIOD_US 100000 是 /sys/fs/cgroup/cpu/cpu.cfs_period_us 默认值
	CPU_CFS_PERIOD_US int = 100000
	// MIN_CPU_SIZE 最小申请cpu值
	MIN_CPU_SIZE = 0.1

	// ProjectNamespace Env
	LabelServiceGroupID                = "servicegroup-id"
	KeyServiceGroupID                  = "SERVICE_GROUP_ID"
	KeyOriginServiceName               = "ORIGIN_SERVICE_NAME"
	ProjectNamespaceServiceNameNameKey = "PROJECT_NAMESPACE_SERVICE_NAME"
	ProjectNamespace                   = "PROJECT_NAMESPACE"
)

// K8S plugin's configure
//
// EXECUTOR_K8S_K8SFORSERVICE_ADDR=http://127.0.0.1:8080
// EXECUTOR_K8S_K8SFORSERVICE_BASICAUTH=
// EXECUTOR_K8S_K8SFORSERVICE_CA_CRT=
// EXECUTOR_K8S_K8SFORSERVICE_CLIENT_CRT=
// EXECUTOR_K8S_K8SFORSERVICE_CLIENT_KEY=
// EXECUTOR_K8S_K8SFORSERVICE_BEARER_TOKEN=
// EXECUTOR_K8S_K8SFORSERVICE_CPU_SUBSCRIBE_RATIO = 1.0
// EXECUTOR_K8S_K8SFORSERVICE_CPU_NUM_QUOTA = 0
func init() {
	_ = executortypes.Register(kind, func(name executortypes.Name, clustername string, options map[string]string, optionsPlus interface{}) (
		executortypes.Executor, error) {
		k, err := New(name, clustername, options)
		if err != nil {
			return k, err
		}

		// 用来存储该 executor 的事件结构
		localStore := &sync.Map{}
		stopCh := make(chan struct{}, 1)
		notifier, err := eventboxapi.New("", nil)
		if err != nil {
			logrus.Errorf("failed to new eventbox api, executor: %s, (%v)", name, err)
			return nil, err
		}
		if err := k.registerEvent(localStore, stopCh, notifier); err != nil {
			logrus.Errorf("failed to register event sync fn, executor: %s, (%v)", name, err)
		}
		// 同步实例状态
		dbclient := instanceinfo.New(dbengine.MustOpen())
		bdl := bundle.New(bundle.WithCMDB())
		syncer := instanceinfosync.NewSyncer(clustername, k.addr, dbclient, bdl, k.pod, k.sts, k.deploy, k.event)
		if options["IS_EDAS"] == "true" {
			return k, nil
		}

		parentctx, cancelSyncInstanceinfo := context.WithCancel(context.Background())
		k.instanceinfoSyncCancelFunc = cancelSyncInstanceinfo
		go func() {
			for {
				select {
				case <-parentctx.Done():
					return
				default:
				}
				ctx, cancel := context.WithCancel(parentctx)
				lock, err := dlock.New(strutil.Concat("/instanceinfosync/", clustername), func() { cancel() })
				if err := lock.Lock(context.Background()); err != nil {
					logrus.Errorf("failed to lock: %v", err)
					continue
				}

				if err != nil {
					logrus.Errorf("failed to get dlock: %v", err)
					// try again
					continue
				}
				syncer.Sync(ctx)
				if err := lock.Unlock(); err != nil {
					logrus.Errorf("failed to unlock: %v", err)
				}
			}
		}()
		return k, nil
	})
}

// Kubernetes is the Executor struct for k8s cluster
type Kubernetes struct {
	name         executortypes.Name
	clusterName  string
	options      map[string]string
	addr         string
	client       *httpclient.HTTPClient
	evCh         chan *eventtypes.StatusEvent
	deploy       *deployment.Deployment
	ds           *ds.Daemonset
	ingress      *ingress.Ingress
	namespace    *namespace.Namespace
	service      *k8sservice.Service
	pvc          *persistentvolumeclaim.PersistentVolumeClaim
	pv           *persistentvolume.PersistentVolume
	sts          *statefulset.StatefulSet
	pod          *pod.Pod
	secret       *secret.Secret
	sa           *serviceaccount.ServiceAccount
	nodeLabel    *nodelabel.NodeLabel
	ClusterInfo  *clusterinfo.ClusterInfo
	resourceInfo *resourceinfo.ResourceInfo
	event        *event.Event
	// 将上层实际设置的 cpu 除以一个比例, 传递给集群调度, 默认为1
	cpuSubscribeRatio        float64
	memSubscribeRatio        float64
	devCpuSubscribeRatio     float64
	devMemSubscribeRatio     float64
	testCpuSubscribeRatio    float64
	testMemSubscribeRatio    float64
	stagingCpuSubscribeRatio float64
	stagingMemSubscribeRatio float64
	// 将 cpu quota 值设置为 cpuNumQuota 个 cpu 的 quota, 默认为0，即cpu quota不限制
	// 当取值为 -1 时，表示使用真实的 cpu 个数来设置 cpu quota (quota 可能还会被其他参数修改，如向上弹的 cpu 个数)
	cpuNumQuota float64

	// operators
	elasticsearchoperator addon.AddonOperator
	redisoperator         addon.AddonOperator
	mysqloperator         addon.AddonOperator
	daemonsetoperator     addon.AddonOperator

	// instanceinfoSyncCancelFunc
	instanceinfoSyncCancelFunc context.CancelFunc

	dbclient *instanceinfo.Client

	istioEngine istioctl.IstioEngine
}

func (k *Kubernetes) GetK8SAddr() string {
	return k.addr
}

// Kind implements executortypes.Executor interface
func (k *Kubernetes) Kind() executortypes.Kind {
	return kind
}

// Name implements executortypes.Executor interface
func (k *Kubernetes) Name() executortypes.Name {
	return k.name
}

func (k *Kubernetes) CleanUpBeforeDelete() {
	k.instanceinfoSyncCancelFunc()
}

// Addr 返回 kubernetes addr
func (k *Kubernetes) Addr() string {
	return k.addr
}

func getWorkspaceRatio(options map[string]string, workspace string, t string, value *float64) {
	var subscribeRatioKey string
	if workspace == "PROD" {
		subscribeRatioKey = t + SUBSCRIBE_RATIO_SUFFIX
	} else {
		subscribeRatioKey = workspace + "_" + t + SUBSCRIBE_RATIO_SUFFIX
	}
	*value = 1.0
	if ratioValue, ok := options[subscribeRatioKey]; ok && len(ratioValue) > 0 {
		if ratio, err := strconv.ParseFloat(ratioValue, 64); err == nil && ratio >= 1.0 {
			*value = ratio
		}
	}
}

func getIstioEngine(info apistructs.ClusterInfoData) (istioctl.IstioEngine, error) {
	istioInfo := info.GetIstioInfo()
	if !istioInfo.Installed {
		return istioctl.EmptyEngine, nil
	}
	// TODO: 获取asm的kubeconfig 创建对应的engine
	if istioInfo.IsAliyunASM {
		return istioctl.EmptyEngine, nil
	}
	apiServerUrl := info.GetApiServerUrl()
	if apiServerUrl == "" {
		return istioctl.EmptyEngine, errors.Errorf("empty api server url, cluster:%s", info.Get(apistructs.DICE_CLUSTER_NAME))
	}
	// TODO: 结合版本进行选择
	localEngine, err := engines.NewLocalEngine(apiServerUrl)
	if err != nil {
		return istioctl.EmptyEngine, errors.Errorf("create local istio engine failed, cluster:%s, err:%v", info.Get(apistructs.DICE_CLUSTER_NAME), err)
	}
	return localEngine, nil
}

// New new kubernetes executor struct
func New(name executortypes.Name, clusterName string, options map[string]string) (*Kubernetes, error) {
	addr, ok := options["ADDR"]
	if !ok {
		return nil, errors.Errorf("not found k8s address in env variables")
	}

	if !strings.HasPrefix(addr, "inet://") {
		if !strings.HasPrefix(addr, "http") && !strings.HasPrefix(addr, "https") {
			addr = strutil.Concat("http://", addr)
		}
	}

	//获取不同环境的超分比的值
	var memSubscribeRatio,
		cpuSubscribeRatio,
		devMemSubscribeRatio,
		devCpuSubscribeRatio,
		testMemSubscribeRatio,
		testCpuSubscribeRatio,
		stagingMemSubscribeRatio,
		stagingCpuSubscribeRatio float64

	getWorkspaceRatio(options, "PROD", "MEM", &memSubscribeRatio)
	getWorkspaceRatio(options, "PROD", "CPU", &cpuSubscribeRatio)
	getWorkspaceRatio(options, "DEV", "MEM", &devMemSubscribeRatio)
	getWorkspaceRatio(options, "DEV", "CPU", &devCpuSubscribeRatio)
	getWorkspaceRatio(options, "TEST", "MEM", &testMemSubscribeRatio)
	getWorkspaceRatio(options, "TEST", "CPU", &testCpuSubscribeRatio)
	getWorkspaceRatio(options, "STAGING", "MEM", &stagingMemSubscribeRatio)
	getWorkspaceRatio(options, "STAGING", "CPU", &stagingCpuSubscribeRatio)

	cpuNumQuota := float64(0)
	if cpuNumQuotaValue, ok := options[CPU_NUM_QUOTA]; ok && len(cpuNumQuotaValue) > 0 {
		if num, err := strconv.ParseFloat(cpuNumQuotaValue, 64); err == nil && (num >= 0 || num == -1.0) {
			cpuNumQuota = num
			logrus.Debugf("executor(%s) cpuNumQuota set to %v", name, cpuNumQuota)
		}
	}

	client := httpclient.New()
	if _, ok := options["CA_CRT"]; ok {
		logrus.Infof("k8s executor(%s) addr for https: %v", name, addr)
		client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
			[]byte(options["CLIENT_KEY"]),
			[]byte(options["CA_CRT"])))

		token, ok := options["BEARER_TOKEN"]
		if !ok {
			return nil, errors.Errorf("not found k8s bearer token")
		}
		// 默认开启了 RBAC, 需要通过 token 进行用户鉴权
		client.BearerTokenAuth(token)
	}

	basicAuth, ok := options["BASICAUTH"]
	if ok {
		namePassword := strings.Split(basicAuth, ":")
		if len(namePassword) == 2 {
			client.BasicAuth(namePassword[0], namePassword[1])
		}
	}

	deploy := deployment.New(deployment.WithCompleteParams(addr, client))
	ds := ds.New(ds.WithCompleteParams(addr, client))
	ing := ingress.New(ingress.WithCompleteParams(addr, client))
	ns := namespace.New(namespace.WithCompleteParams(addr, client))
	svc := k8sservice.New(k8sservice.WithCompleteParams(addr, client))
	pvc := persistentvolumeclaim.New(persistentvolumeclaim.WithCompleteParams(addr, client))
	pv := persistentvolume.New(persistentvolume.WithCompleteParams(addr, client))
	sts := statefulset.New(statefulset.WithCompleteParams(addr, client))
	k8spod := pod.New(pod.WithCompleteParams(addr, client))
	k8ssecret := secret.New(secret.WithCompleteParams(addr, client))
	sa := serviceaccount.New(serviceaccount.WithCompleteParams(addr, client))
	nodeLabel := nodelabel.New(addr, client)
	event := event.New(event.WithCompleteParams(addr, client))
	dbclient := instanceinfo.New(dbengine.MustOpen())

	clusterInfo, err := clusterinfo.New(clusterName, clusterinfo.WithCompleteParams(addr, client))
	if err != nil {
		return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
			name, clusterName, err)
	}
	resourceInfo := resourceinfo.New(addr, client)

	// 同步 cluster info 到 ETCD（10m 一次）
	go clusterInfo.LoopLoadAndSync(context.Background(), true)

	rawData, err := clusterInfo.Get()
	if err != nil {
		return nil, errors.Errorf("failed to get cluster info, executorName:%s, clusterName: %s, err:%v",
			name, clusterName, err)
	}

	clusterInfoData := apistructs.ClusterInfoData{}
	for key, value := range rawData {
		clusterInfoData[apistructs.ClusterInfoMapKey(key)] = value
	}
	istioEngine, err := getIstioEngine(clusterInfoData)
	if err != nil {
		return nil, errors.Errorf("failed to get istio engine, executorName:%s, clusterName:%s, err:%v",
			name, clusterName, err)
	}
	evCh := make(chan *eventtypes.StatusEvent, 10)

	k := &Kubernetes{
		name:                     name,
		clusterName:              clusterName,
		options:                  options,
		addr:                     addr,
		client:                   client,
		evCh:                     evCh,
		deploy:                   deploy,
		ds:                       ds,
		ingress:                  ing,
		namespace:                ns,
		service:                  svc,
		pvc:                      pvc,
		pv:                       pv,
		sts:                      sts,
		pod:                      k8spod,
		secret:                   k8ssecret,
		sa:                       sa,
		nodeLabel:                nodeLabel,
		ClusterInfo:              clusterInfo,
		resourceInfo:             resourceInfo,
		event:                    event,
		cpuSubscribeRatio:        cpuSubscribeRatio,
		memSubscribeRatio:        memSubscribeRatio,
		devCpuSubscribeRatio:     devCpuSubscribeRatio,
		devMemSubscribeRatio:     devMemSubscribeRatio,
		testCpuSubscribeRatio:    testCpuSubscribeRatio,
		testMemSubscribeRatio:    testMemSubscribeRatio,
		stagingCpuSubscribeRatio: stagingCpuSubscribeRatio,
		stagingMemSubscribeRatio: stagingMemSubscribeRatio,
		cpuNumQuota:              cpuNumQuota,
		dbclient:                 dbclient,
		istioEngine:              istioEngine,
	}

	elasticsearchoperator := elasticsearch.New(k, sts, ns, svc, k, k8ssecret, k, client)
	k.elasticsearchoperator = elasticsearchoperator
	redisoperator := redis.New(k, deploy, sts, svc, ns, k, k8ssecret, client)
	k.redisoperator = redisoperator
	mysqloperator := mysql.New(k, ns, k8ssecret, pvc, client)
	k.mysqloperator = mysqloperator
	daemonsetoperator := daemonset.New(k, ns, k, k, ds, k)
	k.daemonsetoperator = daemonsetoperator
	return k, nil
}

// Create implements creating servicegroup based on k8s api
func (k *Kubernetes) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, err := ValidateRuntime(specObj, "Create")
	if err != nil {
		return nil, err
	}

	if runtime.ProjectNamespace != "" {
		k.setProjectNamespaceEnvs(runtime)
	}

	logrus.Debugf("start to create runtime, namespace: %s, name: %s", runtime.Type, runtime.ID)

	operator, ok := runtime.Labels["USE_OPERATOR"]
	if ok {
		op, err := k.whichOperator(operator)
		if err != nil {
			return nil, fmt.Errorf("not found addonoperator: %v", operator)
		}
		return nil, addon.Create(op, runtime)
	}

	if err = k.createRuntime(runtime); err != nil {
		logrus.Errorf("failed to create runtime, namespace: %s, name: %s, (%v)",
			runtime.Type, runtime.ID, err)
		return nil, err
	}
	return nil, nil
}

// Destroy implements deleting servicegroup based on k8s api
func (k *Kubernetes) Destroy(ctx context.Context, specObj interface{}) error {
	runtime, err := ValidateRuntime(specObj, "Destroy")
	if err != nil {
		return err
	}

	operator, ok := runtime.Labels["USE_OPERATOR"]
	if ok {
		op, err := k.whichOperator(operator)
		if err != nil {
			return fmt.Errorf("not found addonoperator: %v", operator)
		}
		// 如果失败,尝试当作普通runtime删除
		if err := addon.Remove(op, runtime); err == nil {
			return nil
		}
	}
	var ns = MakeNamespace(runtime)
	if !IsGroupStateful(runtime) && runtime.ProjectNamespace != "" {
		ns = runtime.ProjectNamespace
		k.setProjectNamespaceEnvs(runtime)
	}
	if runtime.ProjectNamespace == "" {
		if err := k.destroyRuntime(ns); err != nil {
			if k8serror.NotFound(err) {
				logrus.Debugf("k8s namespace not found or already deleted, namespace: %s", ns)
				return nil
			}
			return err
		}
		// 删除有状态服务的 local pv
		if err := k.DeletePV(runtime); err != nil {
			logrus.Errorf("failed to delete pv, namespace: %s, name: %s, (%v)", runtime.Type, runtime.ID, err)
			return err
		}
		logrus.Debugf("succeed to destroy runtime, namespace: %s, name: %s", runtime.Type, runtime.ID)
		return nil
	} else {
		err = k.destroyRuntimeByProjectNamespace(ns, runtime)
		if err != nil {
			logrus.Errorf("failed to delete runtime resource %v", err)
			return err
		}
	}
	return nil
}

// Status implements getting servicegroup status based on k8s api
func (k *Kubernetes) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var status apistructs.StatusDesc
	runtime, err := ValidateRuntime(specObj, "Status")
	if err != nil {
		return status, err
	}

	return k.getGroupStatus(ctx, runtime)
}

// Remove implements removing servicegroup based on k8s api
func (k *Kubernetes) Remove(ctx context.Context, specObj interface{}) error {
	// TODO: currently as same as Destroy
	return k.Destroy(ctx, specObj)
}

// Update implements updating servicegroup based on k8s api
// 暂不支持更新云盘(pvc)
func (k *Kubernetes) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, err := ValidateRuntime(specObj, "Update")
	if err != nil {
		return nil, err
	}

	operator, ok := runtime.Labels["USE_OPERATOR"]
	if ok {
		op, err := k.whichOperator(operator)
		if err != nil {
			return nil, fmt.Errorf("not found addonoperator: %v", operator)
		}
		return nil, addon.Update(op, runtime)
	}

	// Update需要区分出普通的更新和中途创建失败后系统删除了runtime(namespace)后"重新分析"的更新，
	// 这种更新其实就是创建runtime了, 通过namespace是否存在去判断
	var ns = MakeNamespace(runtime)
	if !IsGroupStateful(runtime) && runtime.ProjectNamespace != "" {
		ns = runtime.ProjectNamespace
		k.setProjectNamespaceEnvs(runtime)
	}

	notFound, err := k.NotfoundNamespace(ns)
	if err != nil {
		logrus.Errorf("failed to get whether namespace existed, ns: %s, (%v)", ns, err)
		return nil, err
	}
	// namespace不存在，此更新相当于创建
	if notFound {
		if err = k.createRuntime(runtime); err != nil {
			return nil, err
		}

		logrus.Debugf("succeed to create runtime, namespace: %s, name: %s", runtime.Type, runtime.ID)
		return nil, nil
	}
	// namespace存在，走普通更新的流程
	// Update提供两种实现
	// 1, forceUpdate, 全量删除再创建
	// 2, updateOneByOne, 分类出要创建的服务，要更新的服务，要删除的服务等三类并逐一处理

	// 目前无状态采用updateOneByOne
	return nil, k.updateRuntime(runtime)
}

// Inspect implements getting servicegroup info
func (k *Kubernetes) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, err := ValidateRuntime(specObj, "Inspect")
	if err != nil {
		return nil, err
	}

	operator, ok := runtime.Labels["USE_OPERATOR"]
	if ok {
		op, err := k.whichOperator(operator)
		if err != nil {
			return nil, fmt.Errorf("not found addonoperator: %v", operator)
		}
		return addon.Inspect(op, runtime)
	}

	if IsGroupStateful(runtime) {
		return k.InspectStateful(runtime)
	}

	// 元数据信息从上层传入，这里只需要拿到runtime的状态并拼装到runtime里返回
	status, err := k.Status(ctx, specObj)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("inspect runtime status, runtime: %s, status: %+v", runtime.ID, status)
	runtime.Status = status.Status
	runtime.LastMessage = status.LastMessage

	return k.inspectStateless(runtime)
}

// Cancel implements canceling manipulating servicegroup
func (k *Kubernetes) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, nil
}

func (k *Kubernetes) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	sg, err := ValidateRuntime(specObj, "Precheck")
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}
	resourceinfo, err := k.resourceInfo.Get(true)
	if err != nil {
		return apistructs.ServiceGroupPrecheckData{}, err
	}
	return precheck(sg, resourceinfo)
}

func (k *Kubernetes) CapacityInfo() apistructs.CapacityInfoData {
	r := apistructs.CapacityInfoData{}
	r.ElasticsearchOperator = k.elasticsearchoperator.IsSupported()
	r.RedisOperator = k.redisoperator.IsSupported()
	r.MysqlOperator = k.mysqloperator.IsSupported()
	r.DaemonsetOperator = k.daemonsetoperator.IsSupported()
	return r
}

func (k *Kubernetes) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	r, err := k.resourceInfo.Get(brief)
	if err != nil {
		return r, err
	}
	r.ProdCPUOverCommit = k.cpuSubscribeRatio
	r.DevCPUOverCommit = k.devCpuSubscribeRatio
	r.TestCPUOverCommit = k.testCpuSubscribeRatio
	r.StagingCPUOverCommit = k.stagingCpuSubscribeRatio
	r.ProdMEMOverCommit = k.memSubscribeRatio
	r.DevMEMOverCommit = k.devMemSubscribeRatio
	r.TestMEMOverCommit = k.testMemSubscribeRatio
	r.StagingMEMOverCommit = k.stagingMemSubscribeRatio

	return r, nil
}

func (*Kubernetes) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for kubernetes")
}

func (k *Kubernetes) KillPod(podname string) error {
	pods, err := k.dbclient.PodReader().ByCluster(k.clusterName).ByPodName(podname).Do()
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("cluster(%s),pod:(%s), not found", k.clusterName, podname)
	}
	if len(pods) > 1 {
		return fmt.Errorf("cluster(%s),pod:(%s), find multiple pods", k.clusterName, podname)
	}
	pod := pods[0]
	return k.killPod(pod.K8sNamespace, podname)
}

// 两个接口可能会调用此函数
// 1, Create
// 2, Update
func (k *Kubernetes) createOne(service *apistructs.Service, sg *apistructs.ServiceGroup) error {

	if service == nil {
		return errors.Errorf("service empty")
	}

	if sg.ProjectNamespace != "" && len(service.Ports) > 0 {
		var runtimeService, _ = deepcopy.Copy(service).(*apistructs.Service)

		runtimeService.Env[KeyOriginServiceName] = service.Name
		runtimeService.Name = service.Env[ProjectNamespaceServiceNameNameKey]
		if err := k.updateService(runtimeService); err != nil {
			return err
		}

		if k.istioEngine != istioctl.EmptyEngine {
			if err := k.istioEngine.OnServiceOperator(istioctl.ServiceCreate, runtimeService); err != nil {
				return err
			}
		}
	}

	// Step 1. Firstly create service
	// 只对填写了暴露端口的服务创建k8s service
	if len(service.Ports) > 0 {
		if err := k.updateService(service); err != nil {
			return err
		}
	}
	var err error
	switch service.WorkLoad {
	case ServicePerNode:
		err = k.createDaemonSet(service, sg)
	default:
		// Step 2. Create related deployment
		err = k.createDeployment(service, sg)
	}
	if err != nil {
		return err
	}

	// TODO: Wait for the deployment running well ?
	// status, err := m.getDeployment(service)

	if k.istioEngine != istioctl.EmptyEngine {
		if err := k.istioEngine.OnServiceOperator(istioctl.ServiceCreate, service); err != nil {
			return err
		}
	}

	return nil
}

// not sure with whether this service exists
// bool 变量表示是否真正删除
// 在以下情况发生,
// 1, Update接口分析出的删除，此时无法确保k8s的资源是否存在
func (k *Kubernetes) tryDelete(namespace, name string) error {
	var (
		wg         sync.WaitGroup
		err1, err2 error
	)
	if k.istioEngine != istioctl.EmptyEngine {
		svc := &apistructs.Service{Namespace: namespace, Name: name}
		if err := k.istioEngine.OnServiceOperator(istioctl.ServiceDelete, svc); err != nil {
			return err
		}
	}
	wg.Add(2)
	go func() {
		err1 = k.deleteDeployment(namespace, name)
		wg.Done()
	}()
	go func() {
		err2 = k.deleteDaemonSet(namespace, name)
		wg.Done()
	}()
	wg.Wait()

	if err1 != nil || err2 != nil {
		return errors.Errorf("failed to delete deployment or  daemonset, namespace: %s, name: %s, (%v, %v)",
			namespace, name, err1, err2)
	}

	return nil
}

func (k *Kubernetes) getClusterIP(namespace, name string) (string, error) {
	svc, err := k.GetService(namespace, name)
	if err != nil {
		return "", err
	}
	return svc.Spec.ClusterIP, nil
}

// 创建操作需要先于更新操作完成，因为新创建的service有可能是要更新的服务的依赖
// TODO: 后面要抽象出updateOne函数
func (k *Kubernetes) updateOneByOne(sg *apistructs.ServiceGroup) error {
	labelSelector := make(map[string]string)
	var ns = sg.ProjectNamespace
	if ns == "" {
		ns = MakeNamespace(sg)
	} else {
		labelSelector[LabelServiceGroupID] = sg.ID
	}

	visited := make([]string, 0)
	oldSpecServices, err := k.listServiceName(ns, labelSelector)
	if err != nil {
		//TODO:
		return err
	}
	// runtime.Services are services with desired states
	for _, svc := range sg.Services {
		svc.Namespace = ns
		// find it in oldServices
		found := false
		deployName := getDeployName(&svc)
		for _, s := range oldSpecServices {
			if deployName != s {
				continue
			}
			found = true
			break
		}
		if found {
			// 已存在老的service集合中，做put操作
			// visited记录已更新过的service
			visited = append(visited, deployName)
			continue
		}

		// 不存在于老的service集合中，做create操作
		// TODO: what to do if errors in Create ? before k8s create deployment ?
		// logrus.Debugf("in Update interface, going to create service(%s/%s)", ns, svc.Name)
		if err := k.createOne(&svc, sg); err != nil {
			logrus.Errorf("failed to create service in update interface, name: %s, (%v)", svc.Name, err)
			return err
		}
	}

	// update 操作
	for _, svc := range sg.Services {
		deployName := getDeployName(&svc)
		for _, s := range visited {
			if s != deployName {
				continue
			}

			// firstly update the service
			// service同deployment不一样，只对暴露端口的服务才会创建service
			if err := k.updateService(&svc); err != nil {
				return err
			}
			switch svc.WorkLoad {
			case ServicePerNode:
				desireDaemonSet, err := k.newDaemonSet(&svc, sg)
				if err != nil {
					return err
				}
				if err = k.updateDaemonSet(desireDaemonSet); err != nil {
					logrus.Debugf("failed to update daemonset in update interface, name: %s, (%v)", svc.Name, err)
					return err
				}
			default:
				// then update the deployment
				desiredDeployment, err := k.newDeployment(&svc, sg)
				if err != nil {
					return err
				}
				if err = k.putDeployment(desiredDeployment); err != nil {
					logrus.Debugf("failed to update deployment in update interface, name: %s, (%v)", svc.Name, err)
					return err
				}
			}
			if k.istioEngine != istioctl.EmptyEngine {
				if err := k.istioEngine.OnServiceOperator(istioctl.ServiceUpdate, &svc); err != nil {
					return err
				}
			}
			break
		}
	}

	if len(visited) == len(oldSpecServices) {
		return nil
	}

	// 老的service中除去已被更新的service即需要删除的service
	toBeDeleted := make([]string, 0)
	for _, s := range oldSpecServices {
		existed := false
		for _, v := range visited {
			if s == v {
				existed = true
				break
			}
		}
		if !existed {
			toBeDeleted = append(toBeDeleted, s)
		}
	}

	k8sServices := []string{}

	for _, svc := range sg.Services {
		deployName := getDeployName(&svc)
		for _, svcName := range toBeDeleted {
			if deployName == svcName {
				k8sServices = append(k8sServices, svc.Name)
				break
			}
		}
	}

	for _, svcName := range toBeDeleted {
		// logrus.Debugf("in Update interface, going to delete service(%s/%s)", ns, svcName)
		// TODO: what to do if errors in DELETE ?
		if err := k.tryDelete(ns, svcName); err != nil {
			logrus.Errorf("failed to delete service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
			return err
		}
	}

	for _, svcName := range toBeDeleted {
		if err = k.service.Delete(ns, svcName); err != nil {
			logrus.Errorf("failed to delete k8s service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
			return err
		}
	}

	for _, svc := range k8sServices {
		deploys, err := k.deploy.List(ns, map[string]string{"app": svc})
		if err != nil {
			logrus.Errorf("failed to get deploys in ns %s", ns)
			return err
		}

		remainCount := 0
		for _, deploy := range deploys.Items {
			if deploy.DeletionTimestamp == nil {
				remainCount++
			}
		}
		if remainCount < 1 {
			err = k.service.Delete(ns, svc)
			if err != nil {
				logrus.Errorf("failed to delete global service %s in ns %s", svc, ns)
				return err
			}
		}
	}

	return nil
}

func (k *Kubernetes) getStatelessStatus(ctx context.Context, sg *apistructs.ServiceGroup) (apistructs.StatusDesc, error) {
	var resultStatus apistructs.StatusDesc
	// init "unknown" status for each service
	for i := range sg.Services {
		sg.Services[i].Status = apistructs.StatusUnknown
		sg.Services[i].LastMessage = ""
	}

	var ns = MakeNamespace(sg)
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		k.setProjectNamespaceEnvs(sg)
	}
	isReady := true
	for i := range sg.Services {
		var (
			status apistructs.StatusDesc
			err    error
		)
		switch sg.Services[i].WorkLoad {
		case ServicePerNode:
			status, err = k.getDaemonSetStatus(&sg.Services[i])
		default:
			// 要区分出下面的异常：
			// 1, 创建过程中发生错误，整个runtime被删掉了后再来查询
			// 2, 其他
			status, err = k.getDeploymentStatus(&sg.Services[i])
		}
		if err != nil {
			// TODO: 状态可改成"Error"..
			status.Status = apistructs.StatusUnknown

			if !k8serror.NotFound(err) {
				return status, err
			}
			notfound, err := k.NotfoundNamespace(ns)
			if err != nil {
				errMsg := fmt.Sprintf("failed to get namespace existed info, namespace:%s, (%v)", ns, err)
				logrus.Errorf(errMsg)
				status.LastMessage = errMsg
				return status, err
			}

			// namespace不存在说明创建途中有错误，runtime已被调度器删除
			if notfound {
				status.Status = apistructs.StatusErrorAndDeleted
				status.LastMessage = fmt.Sprintf("namespace not found, probably deleted, namespace: %s", ns)
			} else {
				// 理论上只会出现在删除namespace过程中，某个deployment已被删除而namespace处于terminating状态并即将被删除
				status.Status = apistructs.StatusUnknown
				status.LastMessage = fmt.Sprintf("found namespace exists but deployment not found,"+
					" namespace: %s, deployment: %s", sg.Services[i].Namespace, getDeployName(&sg.Services[i]))
			}

			return status, err
		}
		if status.Status != apistructs.StatusReady {
			isReady = false
			resultStatus.Status = apistructs.StatusProgressing
			sg.Services[i].Status = apistructs.StatusProgressing
			podstatuses, err := k.pod.GetNamespacedPodsStatus(sg.Services[i].Namespace)
			if err != nil {
				logrus.Errorf("failed to get pod unready reasons, namespace: %v, name: %s, %v",
					sg.Services[i].Namespace,
					getDeployName(&sg.Services[i]), err)
			}
			if len(podstatuses) != 0 {
				sg.Services[i].LastMessage = podstatuses[0].Message
				sg.Services[i].Reason = string(podstatuses[0].Reason)
			}
			continue
		}

		sg.Services[i].Status = apistructs.StatusHealthy
	}

	if isReady {
		resultStatus.Status = apistructs.StatusHealthy
	}
	return resultStatus, nil
}

func (k *Kubernetes) SetOverCommitMem(container *apiv1.Container, memSubscribeRatio float64) error {
	format := container.Resources.Requests.Memory().Format
	origin := container.Resources.Requests.Memory().Value()
	r := resource.NewQuantity(int64(float64(origin)/memSubscribeRatio), format)
	container.Resources.Requests[apiv1.ResourceMemory] = *r
	return nil
}

// SetFineGrainedCPU 设置合适的 cpu ratio & quota
func (k *Kubernetes) SetFineGrainedCPU(container *apiv1.Container, extra map[string]string, cpuSubscribeRatio float64) error {
	// 1, 处理申请 cpu 值
	requestCPU := float64(container.Resources.Requests.Cpu().MilliValue()) / 1000

	if requestCPU < MIN_CPU_SIZE {
		return errors.Errorf("invalid request cpu, value: %v, (which is lower than min cpu(%v))",
			requestCPU, MIN_CPU_SIZE)
	}

	// 2, 处理 cpu 超卖
	ratio := cpupolicy.CalcCPUSubscribeRatio(cpuSubscribeRatio, extra)
	actualCPU := requestCPU / ratio
	container.Resources.Requests[apiv1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%v", actualCPU))

	// 3, 处理最大 cpu, 即对应的 cpu quota，默认不限制 cpu quota, 即对应到 cgroup 下的 cpu.cfs_quota_us 值为 -1
	quota := k.cpuNumQuota

	// 按申请 cpu 来设置最大 cpu
	if k.cpuNumQuota == -1.0 {
		quota = cpupolicy.AdjustCPUSize(requestCPU)
	}

	if quota >= requestCPU {
		container.Resources.Limits = apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%v", requestCPU)),
			apiv1.ResourceMemory: container.Resources.Requests[apiv1.ResourceMemory],
		}
	}

	logrus.Debugf("set container cpu: name: %s, request cpu: %v, actual cpu: %v, subscribe ratio: %v, cpu quota: %v",
		container.Name, requestCPU, actualCPU, ratio, quota)
	return nil
}

func (k *Kubernetes) whichOperator(operator string) (addon.AddonOperator, error) {
	switch operator {
	case "elasticsearch":
		return k.elasticsearchoperator, nil
	case "redis":
		return k.redisoperator, nil
	case "mysql":
		return k.mysqloperator, nil
	case "daemonset":
		return k.daemonsetoperator, nil
	}
	return nil, fmt.Errorf("not found")
}

func (k *Kubernetes) CPUOvercommit(limit float64) float64 {
	return limit / k.cpuSubscribeRatio
}

func (k *Kubernetes) MemoryOvercommit(limit int) int {
	return int(float64(limit) / k.memSubscribeRatio)
}

func GenTolerations() []apiv1.Toleration {
	return []apiv1.Toleration{
		{
			Key:      "node-role.kubernetes.io/lb",
			Operator: "Exists",
			Effect:   "NoSchedule",
		},
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: "Exists",
			Effect:   "NoSchedule",
		},
	}
}

func (k *Kubernetes) setProjectNamespaceEnvs(sg *apistructs.ServiceGroup) {
	for index, service := range sg.Services {
		service.Env[ProjectNamespaceServiceNameNameKey] = k.composeNewKey([]string{service.Name, "-", sg.ID})
		service.Env[KeyServiceGroupID] = sg.ID
		service.Env[ProjectNamespace] = "true"
		sg.Services[index] = service
	}
}

func (k *Kubernetes) composeNewKey(keys []string) string {
	var newKey = strings.Builder{}
	for _, key := range keys {
		newKey.WriteString(key)
	}
	return newKey.String()
}
