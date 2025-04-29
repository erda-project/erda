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

// Package k8s implements managing the servicegroup by k8s cluster
package k8s

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/conf"
	eventboxapi "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events/eventtypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/canal"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/daemonset"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/elasticsearch"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/mysql"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/redis"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/rocketmq"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon/sourcecov"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/clusterinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/configmap"
	ds "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/daemonset"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/deployment"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/event"
	erdahpa "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/hpa"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/instanceinfosync"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/job"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sservice"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/oversubscriberatio"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/persistentvolume"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/pod"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/resourceinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/scaledobject"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/serviceaccount"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/statefulset"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/storageclass"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/k8sclient"
	k8sclientconfig "github.com/erda-project/erda/pkg/k8sclient/config"
	"github.com/erda-project/erda/pkg/k8sclient/scheme"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind = "K8S"

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
		bdl := bundle.New(bundle.WithErdaServer())
		syncer := instanceinfosync.NewSyncer(clustername, k.addr, dbclient, bdl, k.pod, k.sts, k.deploy, k.event, k.hpa, k.scaledObject)

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
				if err != nil {
					logrus.Errorf("failed to new dlock, executor: %s, (%v)", name, err)
					continue
				}
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
	name               executortypes.Name
	clusterName        string
	cluster            *apistructs.ClusterInfo
	options            map[string]string
	addr               string
	client             *httpclient.HTTPClient
	k8sClient          *k8sclient.K8sClient
	bdl                *bundle.Bundle
	evCh               chan *eventtypes.StatusEvent
	deploy             *deployment.Deployment
	job                *job.Job
	ds                 *ds.Daemonset
	namespace          *namespace.Namespace
	service            *k8sservice.Service
	pvc                *persistentvolumeclaim.PersistentVolumeClaim
	pv                 *persistentvolume.PersistentVolume
	scaledObject       *scaledobject.ErdaScaledObject
	hpa                *erdahpa.ErdaHPA
	sts                *statefulset.StatefulSet
	pod                *pod.Pod
	secret             *secret.Secret
	storageClass       *storageclass.StorageClass
	sa                 *serviceaccount.ServiceAccount
	ClusterInfo        *clusterinfo.ClusterInfo
	resourceInfo       *resourceinfo.ResourceInfo
	event              *event.Event
	overSubscribeRatio oversubscriberatio.Interface

	// operators
	elasticsearchoperator addon.AddonOperator
	redisoperator         addon.AddonOperator
	mysqloperator         addon.AddonOperator
	canaloperator         addon.AddonOperator
	daemonsetoperator     addon.AddonOperator
	sourcecovoperator     addon.AddonOperator
	rocketmqoperator      addon.AddonOperator

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

// New new kubernetes executor struct
func New(name executortypes.Name, clusterName string, options map[string]string) (*Kubernetes, error) {
	// get cluster from cluster manager
	bdl := bundle.New(
		bundle.WithClusterManager(),
		bundle.WithErdaServer(),
	)
	cluster, err := bdl.GetCluster(clusterName)
	if err != nil {
		logrus.Errorf("get cluster %s error: %v", clusterName, err)
		return nil, err
	}

	rc, err := k8sclientconfig.ParseManageConfig(clusterName, cluster.ManageConfig)
	if err != nil {
		return nil, errors.Errorf("parse rest.config error: %v", err)
	}

	vpaClient := vpa_clientset.NewForConfigOrDie(rc)
	rc.Timeout = conf.ExecutorClientTimeout()

	k8sClient, err := k8sclient.NewForRestConfig(rc, k8sclient.WithSchemes(scheme.LocalSchemeBuilder...))
	if err != nil {
		return nil, errors.Errorf("failed to get k8s client for cluster %s, %v", clusterName, err)
	}

	addr, client, err := util.GetClient(clusterName, cluster.ManageConfig)
	if err != nil {
		logrus.Errorf("cluster %s get http client and addr error: %v", clusterName, err)
		return nil, err
	}

	logrus.Infof("cluster %s init client success, addr: %s", clusterName, addr)

	deploy := deployment.New(deployment.WithClientSet(k8sClient.ClientSet))
	job := job.New(job.WithCompleteParams(addr, client))
	ds := ds.New(ds.WithClientSet(k8sClient.ClientSet))
	ns := namespace.New(namespace.WithKubernetesClient(k8sClient.ClientSet))
	svc := k8sservice.New(k8sservice.WithCompleteParams(addr, client))
	pvc := persistentvolumeclaim.New(persistentvolumeclaim.WithCompleteParams(addr, client))
	pv := persistentvolume.New(persistentvolume.WithCompleteParams(addr, client))
	scaleObj := scaledobject.New(scaledobject.WithCompleteParams(addr, client), scaledobject.WithVPAClient(vpaClient))
	hpa := erdahpa.New(erdahpa.WithCompleteParams(addr, client))
	sts := statefulset.New(statefulset.WithCompleteParams(addr, client))
	k8spod := pod.New(pod.WithK8sClient(k8sClient.ClientSet))
	k8ssecret := secret.New(secret.WithCompleteParams(addr, client))
	cfgmap := configmap.New(configmap.WithCompleteParams(addr, client))
	k8sstorageclass := storageclass.New(storageclass.WithCompleteParams(addr, client))
	sa := serviceaccount.New(serviceaccount.WithCompleteParams(addr, client))
	event := event.New(event.WithKubernetesClient(k8sClient.ClientSet))
	dbclient := instanceinfo.New(dbengine.MustOpen())

	clusterInfo, err := clusterinfo.New(clusterName, clusterinfo.WithCompleteParams(addr, client), clusterinfo.WithDB(dbclient))
	if err != nil {
		return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
			name, clusterName, err)
	}
	resourceInfo := resourceinfo.New(k8sClient.ClientSet)

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
		istioEngine, err = getIstioEngine(clusterName, clusterInfoData)
		if err != nil {
			return nil, errors.Errorf("failed to get istio engine, executorName:%s, clusterName:%s, err:%v",
				name, clusterName, err)
		}
	}
	evCh := make(chan *eventtypes.StatusEvent, 10)

	// Over subscribe ratios
	osr := oversubscriberatio.New(options)

	// Synchronize cluster info to ETCD (every 10m)
	go clusterInfo.LoopLoadAndSync(context.Background(), true)

	k := &Kubernetes{
		name:               name,
		clusterName:        clusterName,
		cluster:            cluster,
		options:            options,
		addr:               addr,
		client:             client,
		k8sClient:          k8sClient,
		bdl:                bdl,
		evCh:               evCh,
		deploy:             deploy,
		job:                job,
		ds:                 ds,
		namespace:          ns,
		service:            svc,
		pvc:                pvc,
		pv:                 pv,
		scaledObject:       scaleObj,
		hpa:                hpa,
		sts:                sts,
		pod:                k8spod,
		secret:             k8ssecret,
		storageClass:       k8sstorageclass,
		sa:                 sa,
		ClusterInfo:        clusterInfo,
		resourceInfo:       resourceInfo,
		event:              event,
		dbclient:           dbclient,
		overSubscribeRatio: osr,
	}

	if istioEngine != nil {
		k.istioEngine = istioEngine
	}

	elasticsearchoperator := elasticsearch.New(k, sts, ns, svc, osr, k8ssecret, k, cfgmap, client)
	k.elasticsearchoperator = elasticsearchoperator
	redisoperator := redis.New(k, deploy, sts, svc, ns, osr, k8ssecret, client)
	k.redisoperator = redisoperator
	mysqloperator := mysql.New(k, ns, osr, k8ssecret, pvc, client)
	k.mysqloperator = mysqloperator
	canaloperator := canal.New(k, ns, osr, k8ssecret, pvc, client)
	k.canaloperator = canaloperator
	daemonsetoperator := daemonset.New(k, ns, k, k, ds, osr)
	k.daemonsetoperator = daemonsetoperator
	k.sourcecovoperator = sourcecov.New(k, client, osr, ns)
	rocketmqoperator := rocketmq.New(k, ns, client, osr, sts)
	k.rocketmqoperator = rocketmqoperator
	return k, nil
}

// Create implements creating servicegroup based on k8s api
func (k *Kubernetes) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, err := ValidateRuntime(specObj, "Create")
	if err != nil {
		return nil, err
	}
	l := logrus.WithField("namespace/name", runtime.Type+"/"+runtime.ID)

	if runtime.ProjectNamespace != "" {
		k.setProjectServiceName(runtime)
	}

	l.Infof("start to create runtime, serviceGroup: %s", strutil.TryGetJsonStr(runtime))

	ok, reason, err := k.checkQuota(ctx, runtime)
	if err != nil {
		return nil, err
	}
	if !ok {
		quotaErr := NewQuotaError(reason)
		return nil, quotaErr
	}

	operator, ok := runtime.Labels["USE_OPERATOR"]
	if ok {
		l.Infof("use operator %s, to create by operator", operator)
		op, err := k.whichOperator(operator)
		if err != nil {
			return nil, fmt.Errorf("not found addonoperator: %v", operator)
		}
		// addon runtime id
		return nil, addon.Create(op, runtime)
	}

	l.Infof("not use operator, to create runtime normaly")
	if err = k.createRuntime(ctx, runtime); err != nil {
		l.Errorf("failed to create runtime, namespace: %s, name: %s, (%v)",
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
				logrus.Debugf("k8s namespace not found or already deleted, Namespace: %s", ns)
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
		if value, ok := runtime.Labels[pstypes.ErdaPALabelKey]; ok && value == pstypes.ErdaHPALabelValueCancel {
			logrus.Infof("delete pod autoscaler objects in runtime %s on namespace %s", runtime.ID, runtime.ProjectNamespace)
			err = k.cancelErdaPARules(*runtime)
			if err != nil {
				logrus.Errorf("failed to delete runtime resource, delete runtime's pod autoscaler objects error: %v", err)
				return err
			}
		}

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
		if err = k.createRuntime(ctx, runtime); err != nil {
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
	return nil, k.updateRuntime(ctx, runtime)
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

	if serviceName, ok := runtime.Labels["GET_RUNTIME_STATELESS_SERVICE_POD"]; ok && serviceName != "" {
		err = k.getStatelessPodsStatus(runtime, serviceName)
		if err != nil {
			return nil, fmt.Errorf("get pods for servicegroup %+v failed: %v", runtime, err)
		}
		return runtime, nil
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
	r.CanalOperator = k.canaloperator.IsSupported()
	r.DaemonsetOperator = k.daemonsetoperator.IsSupported()
	r.SourcecovOperator = k.sourcecovoperator.IsSupported()
	r.RocketMQOperator = k.rocketmqoperator.IsSupported()
	return r
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

// Scale implements update the replica and resources for one service
func (k *Kubernetes) Scale(ctx context.Context, spec interface{}) (interface{}, error) {
	sg, ok := spec.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.Errorf("invalid servicegroup spec: %#v", spec)
	}

	value, ok := sg.Labels[pstypes.ErdaPALabelKey]
	if !ok {
		return k.manualScale(ctx, spec)
	}

	switch value {
	case pstypes.ErdaHPALabelValueCreate:
		return k.createErdaHPARules(spec)
	case pstypes.ErdaHPALabelValueApply:
		return k.applyErdaHPARules(sg)
	case pstypes.ErdaHPALabelValueCancel:
		return k.cancelErdaHPARules(sg)
	case pstypes.ErdaHPALabelValueReApply:
		return k.reApplyErdaHPARules(sg)
	case pstypes.ErdaVPALabelValueApply:
		return k.applyErdaVPARules(sg)
	case pstypes.ErdaVPALabelValueCancel:
		return k.cancelErdaVPARules(sg)
	case pstypes.ErdaVPALabelValueReApply:
		return k.reApplyErdaVPARules(sg)
	default:
		return nil, errors.Errorf("unknown task scale action")
	}
}
