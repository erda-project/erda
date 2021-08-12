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

	appsv1 "k8s.io/api/apps/v1"

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
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/cpupolicy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind = "K8S"

	// SUBSCRIBE_RATIO_SUFFIX the key suffix of the super ratio
	SUBSCRIBE_RATIO_SUFFIX = "_SUBSCRIBE_RATIO"

	// CPU_NUM_QUOTA cpu limit key
	CPU_NUM_QUOTA = "CPU_NUM_QUOTA"
	// CPU_CFS_PERIOD_US 100000  /sys/fs/cgroup/cpu/cpu.cfs_period_us default value
	CPU_CFS_PERIOD_US int = 100000
	// MIN_CPU_SIZE Minimum application cpu value
	MIN_CPU_SIZE = 0.1

	// ProjectNamespace Env
	LabelServiceGroupID = "servicegroup-id"
)

type RuntimeServiceOperator string

const (
	RuntimeServiceRetain RuntimeServiceOperator = "Retain"
	RuntimeServiceDelete RuntimeServiceOperator = "Delete"
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
		// Synchronize instance status
		dbclient := instanceinfo.New(dbengine.MustOpen())
		bdl := bundle.New(bundle.WithCoreServices())
		syncer := instanceinfosync.NewSyncer(clustername, k.addr, dbclient, bdl, k.pod, k.sts, k.deploy, k.event)

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
	cluster      *apistructs.ClusterInfo
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
	// Divide the CPU actually set by the upper layer by a ratio and pass it to the cluster scheduling, the default is 1
	cpuSubscribeRatio        float64
	memSubscribeRatio        float64
	devCpuSubscribeRatio     float64
	devMemSubscribeRatio     float64
	testCpuSubscribeRatio    float64
	testMemSubscribeRatio    float64
	stagingCpuSubscribeRatio float64
	stagingMemSubscribeRatio float64
	// Set the cpu quota value to cpuNumQuota cpu quota, the default is 0, that is, the cpu quota is not limited
	// When the value is -1, it means that the actual number of cpus is used to set the cpu quota (quota may also be modified by other parameters, such as the number of cpus that pop up)
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
	if k.instanceinfoSyncCancelFunc != nil {
		k.instanceinfoSyncCancelFunc()
	}
}

// Addr return kubernetes addr
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
	// TODO: Take asm's kubeconfig to create the corresponding engine
	if istioInfo.IsAliyunASM {
		return istioctl.EmptyEngine, nil
	}
	apiServerUrl := info.GetApiServerUrl()
	if apiServerUrl == "" {
		return istioctl.EmptyEngine, errors.Errorf("empty api server url, cluster:%s", info.Get(apistructs.DICE_CLUSTER_NAME))
	}
	// TODO: Combine version to choose
	localEngine, err := engines.NewLocalEngine(apiServerUrl)
	if err != nil {
		return istioctl.EmptyEngine, errors.Errorf("create local istio engine failed, cluster:%s, err:%v", info.Get(apistructs.DICE_CLUSTER_NAME), err)
	}
	return localEngine, nil
}

// New new kubernetes executor struct
func New(name executortypes.Name, clusterName string, options map[string]string) (*Kubernetes, error) {
	// get cluster from cluster manager
	b := bundle.New(bundle.WithClusterManager())
	cluster, err := b.GetCluster(clusterName)
	if err != nil {
		logrus.Errorf("get cluster error: %v", cluster)
		return nil, err
	}

	addr, client, err := util.GetClient(clusterName, cluster.ManageConfig)
	if err != nil {
		logrus.Errorf("cluster %s get http client and addr error: %v", clusterName, err)
		return nil, err
	}

	//Get the value of the super-scoring ratio for different environments
	var (
		memSubscribeRatio,
		cpuSubscribeRatio,
		devMemSubscribeRatio,
		devCpuSubscribeRatio,
		testMemSubscribeRatio,
		testCpuSubscribeRatio,
		stagingMemSubscribeRatio,
		stagingCpuSubscribeRatio float64
	)

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

	// Synchronize cluster info to ETCD (every 10m)
	go clusterInfo.LoopLoadAndSync(context.Background(), true)

	var istioEngine istioctl.IstioEngine

	rawData, err := clusterInfo.Get()
	if err != nil {
		logrus.Errorf("failed to get cluster info, executorName:%s, clusterName: %s, err:%v",
			name, clusterName, err)
	} else {
		clusterInfoData := apistructs.ClusterInfoData{}
		for key, value := range rawData {
			clusterInfoData[apistructs.ClusterInfoMapKey(key)] = value
		}
		istioEngine, err = getIstioEngine(clusterInfoData)
		if err != nil {
			return nil, errors.Errorf("failed to get istio engine, executorName:%s, clusterName:%s, err:%v",
				name, clusterName, err)
		}
	}

	evCh := make(chan *eventtypes.StatusEvent, 10)

	k := &Kubernetes{
		name:                     name,
		clusterName:              clusterName,
		cluster:                  cluster,
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
	}

	if istioEngine != nil {
		k.istioEngine = istioEngine
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
		k.setProjectServiceName(runtime)
	}

	logrus.Infof("start to create runtime, namespace: %s, name: %s", runtime.Type, runtime.ID)

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
		// If it fails, try to delete it as a normal runtime
		if err := addon.Remove(op, runtime); err == nil {
			return nil
		}
	}
	var ns = MakeNamespace(runtime)
	if !IsGroupStateful(runtime) && runtime.ProjectNamespace != "" {
		ns = runtime.ProjectNamespace
		k.setProjectServiceName(runtime)
	}
	if runtime.ProjectNamespace == "" {
		logrus.Infof("delete runtime %s on namespace %s", runtime.ID, runtime.Type)
		if err := k.destroyRuntime(ns); err != nil {
			if k8serror.NotFound(err) {
				logrus.Debugf("k8s namespace not found or already deleted, namespace: %s", ns)
				return nil
			}
			return err
		}
		// Delete the local pv of the stateful service
		if err := k.DeletePV(runtime); err != nil {
			logrus.Errorf("failed to delete pv, namespace: %s, name: %s, (%v)", runtime.Type, runtime.ID, err)
			return err
		}
		return nil
	} else {
		logrus.Infof("delete runtime %s on namespace %s", runtime.ID, runtime.ProjectNamespace)
		err = k.destroyRuntimeByProjectNamespace(ns, runtime)
		if err != nil {
			logrus.Errorf("failed to delete runtime resource %v", err)
			return err
		}
	}
	logrus.Infof("delete runtime %s finished", runtime.ID)
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
// Does not support updating cloud disk (pvc)
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

	// Update needs to distinguish between ordinary updates and updates that are "re-analyzed" after the system deletes the runtime (namespace) after the creation fails.
	// This kind of update is actually to create a runtime, which is judged by whether the namespace exists
	var ns = MakeNamespace(runtime)
	if !IsGroupStateful(runtime) && runtime.ProjectNamespace != "" {
		ns = runtime.ProjectNamespace
		k.setProjectServiceName(runtime)
	}

	notFound, err := k.NotfoundNamespace(ns)
	if err != nil {
		logrus.Errorf("failed to get whether namespace existed, ns: %s, (%v)", ns, err)
		return nil, err
	}
	// namespace does not exist, this update is equivalent to creating
	if notFound {
		if err = k.createRuntime(runtime); err != nil {
			return nil, err
		}

		logrus.Debugf("succeed to create runtime, namespace: %s, name: %s", runtime.Type, runtime.ID)
		return nil, nil
	}
	// namespace exists, follow the normal update process
	// Update provides two implementations
	// 1, forceUpdate, Delete all and create again
	// 2, updateOneByOne, Categorize the three types of services to be created, services to be updated, and services to be deleted, and deal with them one by one

	// Stateless service using updateOneByOne currently
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

	// Metadata information is passed in from the upper layer, here you only need to get the state of the runtime and assemble it into the runtime to return
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

// Two interfaces may call this function
// 1, Create
// 2, Update
func (k *Kubernetes) createOne(service *apistructs.Service, sg *apistructs.ServiceGroup) error {

	if service == nil {
		return errors.Errorf("service empty")
	}
	// Step 1. Firstly create service
	// Only create k8s service for services with exposed ports
	if len(service.Ports) > 0 {
		if err := k.updateService(service, nil); err != nil {
			return err
		}
	}
	if service.ProjectServiceName != "" && len(service.Ports) > 0 {
		err := k.createProjectService(service, sg.ID)
		if err != nil {
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
		if service.ProjectServiceName != "" {
			if projectService, ok := deepcopy.Copy(service).(*apistructs.Service); ok {
				projectService.Name = service.ProjectServiceName
				if err := k.istioEngine.OnServiceOperator(istioctl.ServiceCreate, projectService); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// not sure with whether this service exists
// bool The variable indicates whether it is really deleted
// Occurs in the following situations,
// 1, If the update interface analyzes the deletion, it is impossible to ensure the existence of k8s resources at this time
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

	if err1 != nil && !util.IsNotFound(err1) {
		return errors.Errorf("failed to delete deployment, namespace: %s, name: %s, (%v)",
			namespace, name, err2)
	}

	if err2 != nil && !util.IsNotFound(err2) {
		return errors.Errorf("failed to delete daemonset, namespace: %s, name: %s, (%v)",
			namespace, name, err2)
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

// The creation operation needs to be completed before the update operation, because the newly created service may be a dependency of the service to be updated
// TODO: The updateOne function will be abstracted later
func (k *Kubernetes) updateOneByOne(sg *apistructs.ServiceGroup) error {
	labelSelector := make(map[string]string)
	var ns = sg.ProjectNamespace
	if ns == "" {
		ns = MakeNamespace(sg)
	} else {
		labelSelector[LabelServiceGroupID] = sg.ID
	}

	runtimeServiceMap, err := k.getRuntimeServiceMap(ns, labelSelector)
	if err != nil {
		//TODO:
		errMsg := fmt.Sprintf("list runtime service name err %v", err)
		logrus.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	for _, svc := range sg.Services {
		svc.Namespace = ns
		runtimeServiceName := getDeployName(&svc)
		// Existing in the old service collection, do the put operation
		// The visited record has been updated service
		if _, ok := runtimeServiceMap[runtimeServiceName]; ok {
			// firstly update the service
			// service is not the same as deployment, service is only created for services with exposed ports
			if err := k.updateService(&svc, nil); err != nil {
				return err
			}
			if err := k.updateProjectService(&svc, sg.ID); err != nil {
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
			runtimeServiceMap[runtimeServiceName] = RuntimeServiceRetain
		} else {
			// Does not exist in the old service collection, do the create operation
			// TODO: what to do if errors in Create ? before k8s create deployment ?
			// logrus.Debugf("in Update interface, going to create service(%s/%s)", ns, svc.Name)
			if err := k.createOne(&svc, sg); err != nil {
				logrus.Errorf("failed to create service in update interface, name: %s, (%v)", svc.Name, err)
				return err
			}
			runtimeServiceMap[runtimeServiceName] = RuntimeServiceRetain
		}
	}

	for svcName, operator := range runtimeServiceMap {
		if operator == RuntimeServiceDelete {
			if err := k.tryDelete(ns, svcName); err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to delete service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			svc, err := k.service.Get(ns, svcName)
			if err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to get k8s service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			if err = k.service.Delete(ns, svcName); err != nil {
				if !util.IsNotFound(err) {
					logrus.Errorf("failed to delete k8s service in update interface, namespace: %s, name: %s, (%v)", ns, svcName, err)
					return err
				}
			}
			appLabelValue := svc.Spec.Selector["app"]
			deploys, err := k.deploy.List(ns, map[string]string{"app": appLabelValue})
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
				err = k.service.Delete(ns, appLabelValue)
				if err != nil {
					if !util.IsNotFound(err) {
						logrus.Errorf("failed to delete global service %s in ns %s", svc.Name, ns)
						return err
					}
				}
			}
		}
	}
	return nil
}

func (k *Kubernetes) getStatelessStatus(ctx context.Context, sg *apistructs.ServiceGroup) (apistructs.StatusDesc, error) {
	var (
		resultStatus apistructs.StatusDesc
		deploys      []appsv1.Deployment
		dsMap        map[string]appsv1.DaemonSet
		err          error
	)
	// init "unknown" status for each service
	for i := range sg.Services {
		sg.Services[i].Status = apistructs.StatusUnknown
		sg.Services[i].LastMessage = ""
	}

	var ns = MakeNamespace(sg)
	if sg.ProjectNamespace != "" {
		ns = sg.ProjectNamespace
		k.setProjectServiceName(sg)
	}
	isReady := true

	if sg.ProjectNamespace != "" {
		deployList, err := k.deploy.List(ns, map[string]string{
			LabelServiceGroupID: sg.ID,
		})
		if err != nil {
			return apistructs.StatusDesc{}, fmt.Errorf("list deploy in ns %s err: %v", ns, err)
		}
		deploys = deployList.Items

		daemonsetExist := false

		for _, svc := range sg.Services {
			if svc.WorkLoad == ServicePerNode {
				daemonsetExist = true
				break
			}
		}
		if daemonsetExist {
			daemonsets, err := k.ds.List(ns, map[string]string{
				LabelServiceGroupID: sg.ID,
			})
			if err != nil {
				return apistructs.StatusDesc{}, fmt.Errorf("list daemonset in ns %s err: %v", ns, err)
			}

			dsMap = make(map[string]appsv1.DaemonSet, len(daemonsets.Items))
			for _, ds := range daemonsets.Items {
				dsMap[ds.Name] = ds
			}
		}

	} else {
		for _, svc := range sg.Services {
			deployList, err := k.deploy.List(ns, map[string]string{
				"app": svc.Name,
			})
			if err != nil {
				return apistructs.StatusDesc{}, fmt.Errorf("list deploy in ns %s err: %v", ns, err)
			}
			deploys = append(deploys, deployList.Items...)
		}
	}

	deployMap := make(map[string]appsv1.Deployment, len(deploys))
	for _, deploy := range deploys {
		deployMap[deploy.Name] = deploy
	}

	pods, err := k.pod.ListNamespacePods(ns)
	if err != nil {
		return apistructs.StatusDesc{}, err
	}

	for i := range sg.Services {
		var (
			status apistructs.StatusDesc
			err    error
		)
		switch sg.Services[i].WorkLoad {
		case ServicePerNode:
			status, err = k.getDaemonSetStatusFromMap(&sg.Services[i], dsMap)
		default:
			// To distinguish the following exceptions：
			// 1, An error occurred during the creation process, and the entire runtime is deleted and then come back to query
			// 2, Others
			status, err = k.getDeploymentStatusFromMap(&sg.Services[i], deployMap)
		}
		if err != nil {
			// TODO: the state can be chanded to "Error"..
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

			// The namespace does not exist, indicating that there was an error during creation, and the runtime has been deleted by the scheduler
			if notfound {
				status.Status = apistructs.StatusErrorAndDeleted
				status.LastMessage = fmt.Sprintf("namespace not found, probably deleted, namespace: %s", ns)
			} else {
				// In theory, it will only appear in the process of deleting the namespace. A deployment has been deleted and the namespace is in terminating state and is about to be deleted.
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
			podstatuses, err := k.pod.GetNamespacedPodsStatus(pods.Items)
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

// SetFineGrainedCPU Set proper cpu ratio & quota
func (k *Kubernetes) SetFineGrainedCPU(container *apiv1.Container, extra map[string]string, cpuSubscribeRatio float64) error {
	// 1, Processing request cpu value
	requestCPU := float64(container.Resources.Requests.Cpu().MilliValue()) / 1000

	if requestCPU < MIN_CPU_SIZE {
		return errors.Errorf("invalid request cpu, value: %v, (which is lower than min cpu(%v))",
			requestCPU, MIN_CPU_SIZE)
	}

	// 2, Dealing with cpu oversold
	ratio := cpupolicy.CalcCPUSubscribeRatio(cpuSubscribeRatio, extra)
	actualCPU := requestCPU / ratio
	container.Resources.Requests[apiv1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%v", actualCPU))

	// 3, Processing the maximum cpu, that is, the corresponding cpu quota, the default is not limited cpu quota, that is, the value corresponding to cpu.cfs_quota_us under the cgroup is -1
	quota := k.cpuNumQuota

	// Set the maximum cpu according to the requested cpu
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

func (k *Kubernetes) setProjectServiceName(sg *apistructs.ServiceGroup) {
	for index, service := range sg.Services {
		service.ProjectServiceName = k.composeNewKey([]string{service.Name, "-", sg.ID})
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

// Scale implements update the replica and resources for one service
func (k *Kubernetes) Scale(ctx context.Context, spec interface{}) (interface{}, error) {
	sg, err := ValidateRuntime(spec, "TaskScale")

	if err != nil {
		return nil, err
	}

	if !IsGroupStateful(sg) && sg.ProjectNamespace != "" {
		k.setProjectServiceName(sg)
	}

	// only support scale one service resources
	if len(sg.Services) != 1 {
		return nil, fmt.Errorf("the scaling service count is not equal 1")
	}

	if err = k.scaleDeployment(sg); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return sg, nil
}
