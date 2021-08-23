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

package marathon

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events/eventtypes"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/marathon/instanceinfosync"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/cpupolicy"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind = "MARATHON"
	// default version is 1.6
	defaultVersion = "1.6"
	// mode for version >= 1.5
	defaultNetworkMode = "container"
	// mode for version 1.4.7, BRIDGE, HOST, USER
	defaultNetworkMode147  = "USER"
	defaultNetworkName     = "dcos"
	defaultPrefix          = "/runtimes/v1"
	defaultContainerType   = "DOCKER"
	defaultVipSuffix       = ".marathon.l4lb.thisdcos.directory"
	defaultConstraints     = "dice-role:UNLIKE:platform"
	defaultFetchUris       = ""
	defaultInternalLb      = "marathon-lb-internal-lb.marathon.mesos"
	defaultPublicIp        = "none"
	defaultPublicIpGroup   = "internal"
	defaultCloudVolumePath = "/netdata/volumes"

	// The key of the oversold ratio in the configuration
	CPU_SUBSCRIBE_RATIO = "CPU_SUBSCRIBE_RATIO"

	// 100000  /sys/fs/cgroup/cpu/cpu.cfs_period_us default value
	CPU_CFS_PERIOD_US int = 100000
	// Minimum application cpu value
	MIN_CPU_SIZE = 0.1

	HCMethodCommand = "COMMAND"
	HCMethodTCP     = "TCP"
)

func init() {
	// go instanceinfosync.NewSyncer(instanceinfo.New(dbengine.MustOpen())).Sync()
	executortypes.Register(kind, func(
		name executortypes.Name,
		clusterName string,
		options map[string]string,
		optionsPlus interface{}) (executortypes.Executor, error) {
		go func() {
			for {
				lock, err := dlock.New("/instanceinfosync/marathon", func() {})
				if err != nil {
					logrus.Errorf("failed to new dlock:%v", err)
					continue
				}
				if err := lock.Lock(context.Background()); err != nil {
					logrus.Errorf("failed to get dlock: %v", err)
					continue
				}
				instanceinfosync.NewSyncer(instanceinfo.New(dbengine.MustOpen())).Sync()
				if err := lock.Unlock(); err != nil {
					logrus.Errorf("failed to unlock: %v", err)
				}
			}
		}()
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found marathon address in env variables")
		}
		prefix, ok := options["PREFIX"]
		if !ok {
			prefix = defaultPrefix
		}
		version, ok := options["VERSION"]
		if !ok {
			version = defaultVersion
		}
		ver, err := parseVersion(version)
		if err != nil {
			return nil, err
		}
		publicIp, ok := options["PUBLICIP"]
		if !ok {
			publicIp = defaultPublicIp
		}
		publicIpGroup, ok := options["PUBLICIPGROUP"]
		if !ok {
			publicIpGroup = defaultPublicIpGroup
		}

		client := httpclient.New()

		if _, ok := options["CA_CRT"]; ok {
			logrus.Infof("marathon executor(%s) addr for https: %v", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
				[]byte(options["CLIENT_KEY"]),
				[]byte(options["CA_CRT"])))
		}

		basicAuth, ok := options["BASICAUTH"]
		if ok {
			namePassword := strings.Split(basicAuth, ":")
			if len(namePassword) == 2 {
				client.BasicAuth(namePassword[0], namePassword[1])
			}
		}
		backOff, _ := options["BACKOFF"]
		isUnique := false
		unique_, ok := options["UNIQUE"]
		if ok && unique_ == "true" {
			isUnique = true
		}
		isDisableAutoRestart := false
		disableAutoRestart_, ok := options["ADDONS_DISABLE_AUTO_RESTART"]
		if ok && disableAutoRestart_ == "true" {
			isDisableAutoRestart = true
		}

		cpuSubscribeRatio := 1.0
		if cpuSubscribeRatio_, ok := options[CPU_SUBSCRIBE_RATIO]; ok && len(cpuSubscribeRatio_) > 0 {
			if ratio, err := strconv.ParseFloat(cpuSubscribeRatio_, 64); err == nil && ratio >= 1.0 {
				cpuSubscribeRatio = ratio
				logrus.Debugf("executor(%s) default cpuSubscribeRatio set to %v", name, ratio)
			}
		}

		cpuNumQuota := float64(0)
		if cpuNumQuota_, ok := options["CPU_NUM_QUOTA"]; ok && len(cpuNumQuota_) > 0 {
			if num, err := strconv.ParseFloat(cpuNumQuota_, 64); err == nil && (num >= 0 || num == -1.0) {
				cpuNumQuota = num
				logrus.Debugf("executor(%s) cpuNumQuota set to %v", name, cpuNumQuota)
			}
		}

		go util.GetAndSetTokenAuth(client, string(name))

		enableTag, err := util.ParseEnableTagOption(options, "ENABLETAG", false)
		if err != nil {
			return nil, err
		}
		preserveProjects := util.ParsePreserveProjects(options, "PRESERVEPROJECTS")
		workspaceTags := util.ParsePreserveProjects(options, "WORKSPACETAGS")

		evCh := make(chan *eventtypes.StatusEvent, 200)

		dockerCfg := options["DOCKER_CONFIG"]
		dockerEnv := options["DOCKER_ENV"]

		// cluster info
		js, err := jsonstore.New()
		if err != nil {
			return nil, errors.Errorf("failed to new json store for clusterinfo, executor: %s, (%v)",
				name, err)
		}
		ci := clusterinfo.NewClusterInfoImpl(js)

		m := &Marathon{
			name:             name,
			clusterName:      clusterName,
			options:          options,
			addr:             addr,
			prefix:           prefix,
			pubIp:            publicIp,
			pubGrp:           publicIpGroup,
			version:          ver,
			client:           client,
			enableTag:        enableTag,
			preserveProjects: preserveProjects,
			workspaceTags:    workspaceTags,
			evCh:             evCh,
			backoff:          backOff,
			unique:           isUnique,
			// set addons' service restarting period 3600s
			addonsDisableAutoRestart: isDisableAutoRestart,
			cpuSubscribeRatio:        cpuSubscribeRatio,
			cpuNumQuota:              cpuNumQuota,

			dockerCfg: dockerCfg,
			dockerEnv: dockerEnv,

			db:          instanceinfo.New(dbengine.MustOpen()),
			clusterInfo: ci,
		}

		if disableEvent := os.Getenv("DISABLE_EVENT"); disableEvent == "true" {
			return m, nil
		}
		// The key is {runtimeNamespace}/{runtimeName}, and the value is the address of the corresponding event structure
		lstore := &sync.Map{}
		stopCh := make(chan struct{}, 1)
		registerEventChanAndLocalStore(string(name), evCh, stopCh, lstore)

		collectCh := make(chan string, 100)
		suspend := false
		if forceKill_, ok := options["FORCE_KILL"]; ok && forceKill_ == "true" {
			suspend = true
			go m.SuspendApp(collectCh)
		}
		go m.WaitEvent(options, suspend, collectCh, stopCh)

		go m.initEventAndPeriodSync(string(name), lstore, stopCh)

		return m, nil
	})
}

type Marathon struct {
	name             executortypes.Name
	clusterName      string
	options          map[string]string
	addr             string
	prefix           string
	version          Ver
	pubIp            string
	pubGrp           string
	client           *httpclient.HTTPClient
	enableTag        bool
	preserveProjects map[string]struct{}
	workspaceTags    map[string]struct{}
	evCh             chan *eventtypes.StatusEvent
	// Set the backoff factor of marathon for a specific cluster
	backoff string
	// Set up balanced instance scheduling for a specific cluster
	unique bool
	// Set the service for a specific cluster without being pulled up by marathon after it hangs (by setting the backoff time to 3600)
	// To take effect, the following two conditions must be met at the same time:
	// 1, Set ADDONS_DISABLE_AUTO_RESTART to "true" in the cluster configuration
	// 2， Pass in SERVICE_TYPE in the label of ServiceGroup (ServiceGroup): ADDONS
	addonsDisableAutoRestart bool
	// Divide the CPU actually set by the upper layer by a ratio and pass it to the cluster scheduling, the default is 1
	cpuSubscribeRatio float64
	// Set the cpu quota value to cpuNumQuota cpu quota, the default is 0, that is, the cpu quota is not limited
	// When the value is -1, it means that the actual number of cpus is used to set the cpu quota (quota may also be modified by other parameters, such as the number of cpus that pop up)
	cpuNumQuota float64
	// Additional configuration for docker, such as swapiness
	dockerCfg string
	// Set up docker env
	dockerEnv string

	db          *instanceinfo.Client
	clusterInfo clusterinfo.ClusterInfo
}

type Ver []int

func (m *Marathon) Kind() executortypes.Kind {
	return kind
}

func (m *Marathon) Name() executortypes.Name {
	return m.name
}

func (m *Marathon) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("invalid service spec")
	}
	if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
		err := m.probeProxyPortsForCreate(http.MethodPost, &runtime)
		if err != nil {
			return nil, err
		}
		// retrieve and append proxyPorts if needed
		err = m.resolveProxyPorts(&runtime)
		if err != nil {
			return nil, err
		}
		// TODO: refactor it, double put the proxyPort env
		err = m.probeProxyPortsForCreate(http.MethodPut, &runtime)
		if err != nil {
			return nil, err
		}
	}
	// construct marathon Group entity
	mGroup, err := m.buildMarathonGroup(runtime)
	if err != nil {
		return nil, errors.Wrap(err, "build marathon group failed")
	}
	logrus.Infof("mgroup: %+v\n", mGroup) // debug print

	if runtime.ScheduleInfo.HostUnique {
		placeHolderHosts, err := m.createPlaceHolderGroup(&runtime, mGroup)
		if err != nil && err != errNoNeedBuildPlaceHolder {
			return nil, errors.Wrap(err, "create place holder group failed")
		}
		if err != errNoNeedBuildPlaceHolder {
			if err := updateGroupByPlaceHolderHosts(mGroup, placeHolderHosts); err != nil {
				return nil, err
			}
			logrus.Infof("group after modified: %+v", mGroup)
		}
	}

	// POST the constructed Group to marathon
	var method string = http.MethodPost
	if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy || runtime.ScheduleInfo.HostUnique {
		// already created group
		method = http.MethodPut
	}

	ret, err := m.putGroup(method, mGroup, runtime.Force)
	if err != nil {
		return nil, errors.Wrap(err, "post marathon group failed")
	}
	// TODO: return ret
	_ = ret
	return nil, nil
}

func (m *Marathon) Destroy(ctx context.Context, specObj interface{}) error {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return errors.New("invalid service spec")
	}
	// marathon GroupId
	mGroupId := buildMarathonGroupId(m.prefix, runtime.Type, runtime.ID)
	ret, err := m.deleteGroup(mGroupId, runtime.Force)
	if err != nil {
		return errors.Wrapf(err, "cleanup marathon group failed, groupId: %s", mGroupId)
	}
	// TODO: use this ret to track deployment success
	_ = ret
	return nil
}

func (m *Marathon) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var status apistructs.StatusDesc
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return status, errors.New("invalid service spec")
	}
	// marathon GroupId
	mGroupId := buildMarathonGroupId(m.prefix, runtime.Type, runtime.ID)
	mGroup, err := m.getGroupWithDefaultParam(mGroupId)
	if err != nil {
		return status, errors.Wrapf(err, "get marathon group failed")
	}
	queue, err := m.getQueue()
	if err != nil {
		return status, errors.Wrapf(err, "get marathon queue failed")
	}
	mAllAppStatus := getAllAppStatus(&mGroup.Apps, queue)
	return combineRuntimeStatus(mAllAppStatus, &mGroup.Apps), nil
}

func (m *Marathon) Remove(ctx context.Context, specObj interface{}) error {
	// TODO: currently as same as Destroy
	return m.Destroy(ctx, specObj)
}

// Update update marathon group
// NOTE:
// If HOST_UNIQUE=true in the updated servicegroup, return an error directly
// Because after updating the group, the original host may no longer meet the current constraints.
// And re-obtaining the available host like create will invalidate the blue-green release of the service
func (m *Marathon) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("invalid service spec")
	}
	if runtime.ScheduleInfo.HostUnique {
		return nil, errors.New("HOST_UNIQUE is not support in servicegroup update")
	}
	if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
		// retrieve and append proxyPorts if needed
		err := m.resolveProxyPorts(&runtime)
		if err != nil {
			return nil, err
		}
	}
	// construct marathon Group entity
	mGroup, err := m.buildMarathonGroup(runtime)
	if err != nil {
		return nil, errors.Wrap(err, "build marathon group failed")
	}
	// POST the constructed Group to marathon
	ret, err := m.putGroup(http.MethodPut, mGroup, runtime.Force)
	if err != nil {
		return nil, errors.Wrap(err, "put marathon group failed")
	}
	// TODO: return ret
	_ = ret
	return nil, nil
}

func (m *Marathon) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("invalid service spec")
	}
	// marathon GroupId
	mGroupId := buildMarathonGroupId(m.prefix, runtime.Type, runtime.ID)
	mGroup, err := m.getGroupWithDefaultParam(mGroupId)
	if err != nil {
		return nil, errors.Wrapf(err, "get marathon group failed")
	}
	queue, err := m.getQueue()
	if err != nil {
		return nil, errors.Wrapf(err, "get marathon queue failed")
	}
	mAllAppStatus := getAllAppStatus(&mGroup.Apps, queue)
	runtime.StatusDesc = combineRuntimeStatus(mAllAppStatus, &mGroup.Apps)
	for i := range runtime.Services {
		service := &runtime.Services[i]
		mAppId := buildMarathonAppId(mGroupId, service.Name)
		service.StatusDesc = calculateServiceStatus(mAllAppStatus[mAppId])
		mAppLevelVip := buildMarathonVipAppLevelPart(mAppId)
		service.Vip = buildMarathonVip(mAppLevelVip)
		service.ShortVIP = service.Name

		index := -1
		for idx, app := range mGroup.Apps {
			if app.Id == mAppId {
				index = idx
				break
			}
		}

		if index >= 0 {
			service.InstanceInfos = make([]apistructs.InstanceInfo, len(mGroup.Apps[index].Tasks))
			for j, task := range mGroup.Apps[index].Tasks {
				var instance apistructs.InstanceInfo

				instance.Id = task.Id
				instance.Status = task.State
				if len(task.InstanceIpAddresses) > 0 {
					instance.Ip = task.InstanceIpAddresses[0].InstanceIp
				}

				if len(task.HealthCheckResults) > 0 {
					if task.HealthCheckResults[0].Alive {
						instance.Alive = "true"
					} else {
						instance.Alive = "false"
					}
				} else {
					// For instances where the health check is not started for some reason, the status is temporarily deemed to have failed the health check
					if service.NewHealthCheck != nil || service.HealthCheck != nil {
						instance.Alive = "false"
					} else {
						instance.Alive = "noHealthCheckSpecified"
					}
				}

				service.InstanceInfos[j] = instance
			}
		}

		setServiceDetailedResourceInfo(service, queue, mAppId, mAllAppStatus[mAppId])
	}

	if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
		err := m.resolveProxyPorts(&runtime)
		if err != nil {
			return nil, err
		}
	}
	// TODO: currently only status to inspect, support tasks list later
	return &runtime, nil
}

func (m *Marathon) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {
	runtime, ok := specObj.(apistructs.ServiceGroup)
	if !ok {
		return nil, errors.New("invalid service spec")
	}

	var deployments Deployments
	resp, err := m.client.Get(m.addr).Path("/v2/deployments").Do().JSON(&deployments)
	if err != nil {
		return nil, errors.Errorf("marathon get deployments list error: %v", err)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("marathon get deployments list failed, statusCode=%d", resp.StatusCode())
	}

	name := runtime.ID
	found := false
	var deploymentID string

	for _, v := range deployments {
		if found {
			break
		}
		for _, app := range v.AffectedApps {
			if strings.Contains(app, name) {
				deploymentID = v.ID
				found = true
				break
			}
		}
	}
	if !found {
		return "", nil
	}
	logrus.Infof("got deployment id(%v) to delete", deploymentID)
	resp, err = m.client.Delete(m.addr).Path("/v2/deployments/" + deploymentID).Do().DiscardBody()
	if err != nil {
		return nil, errors.Errorf("marathon delete deployment(%s) error: %v", deploymentID, err)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("marathon delete deployments(%s) failed, statusCode=%d",
			deploymentID, resp.StatusCode())
	}
	return deploymentID, nil
}
func (m *Marathon) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

func (m *Marathon) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

func (m *Marathon) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{
		ProdCPUOverCommit:    1.0,
		DevCPUOverCommit:     1.0,
		TestCPUOverCommit:    1.0,
		StagingCPUOverCommit: 1.0,
		ProdMEMOverCommit:    1.0,
		DevMEMOverCommit:     1.0,
		TestMEMOverCommit:    1.0,
		StagingMEMOverCommit: 1.0,
		Nodes: map[string]*apistructs.NodeResourceInfo{
			"1.1.1.1": {
				Labels:         []string{},
				IgnoreLabels:   true,
				Ready:          true,
				CPUAllocatable: 999999999999999.0,
				MemAllocatable: 999999999999999,
			},
		},
	}, nil
}

func getAllAppStatus(mApps *[]App, queue *Queue) map[string]AppStatus {
	ret := make(map[string]AppStatus)
	for i := range *mApps {
		mApp := (*mApps)[i]
		// https://mesosphere.github.io/marathon/docs/marathon-ui.html
		// https://github.com/mesosphere/marathon-ui/blob/master/src/js/stores/AppsStore.js
		var mAppStatus AppStatus
		for _, offer := range queue.Queue {
			if mApp.Id == offer.App.Id {
				if offer.Delay.Overdue {
					mAppStatus = AppStatusWaiting
				} else {
					mAppStatus = AppStatusDelayed
				}
				break
			}
		}
		if mAppStatus == "" {
			if len(mApp.Deployments) > 0 {
				mAppStatus = AppStatusDeploying
				// TODO: support readiness ?
			} else if mApp.Instances == 0 && mApp.TasksRunning == 0 {
				mAppStatus = AppStatusHealthy
			} else if mApp.TasksRunning > 0 {
				// Only when the health check is set and there are unhealthy instances, the status is AppStatusRunning
				if mApp.TasksUnhealthy > 0 {
					mAppStatus = AppStatusRunning
				} else {
					mAppStatus = AppStatusHealthy
				}
			}
		}
		// TODO: also check app health status
		ret[mApp.Id] = mAppStatus
	}
	return ret
}

func calculateServiceStatus(mAppStatus AppStatus) apistructs.StatusDesc {
	switch mAppStatus {
	case AppStatusDelayed:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "service failed",
		}
	case AppStatusWaiting:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "service have not resource yet",
		}
	case AppStatusSuspended:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "service stopped",
		}
	case AppStatusDeploying:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "we are progressing, soon to ready!",
		}
	case AppStatusRunning:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "service is not healthy",
		}
	case AppStatusHealthy:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusReady,
			LastMessage: "service is ready to serve",
		}
	default:
		return apistructs.StatusDesc{
			Status:      apistructs.StatusUnknown,
			LastMessage: "service unknown condition",
		}
	}
}

func combineRuntimeStatus(mAllAppStatus map[string]AppStatus, mApps *[]App) apistructs.StatusDesc {
	counts := make(map[AppStatus]int)
	for _, v := range mAllAppStatus {
		counts[v] += 1
	}
	if counts[AppStatusDelayed] > 0 {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusFailing,
			LastMessage: "some service failed",
		}
	}
	if counts[AppStatusWaiting] > 0 {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusFailing,
			LastMessage: "some service have not resource yet",
		}
	}
	if counts[AppStatusSuspended] > 0 {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusFailing,
			LastMessage: "some service stopped",
		}
	}
	if counts[AppStatusDeploying] > 0 {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "we are progressing, soon to ready!",
		}
	}
	if counts[AppStatusHealthy] == len(*mApps) {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusReady,
			LastMessage: "services are ready to serve",
		}
	} else if counts[AppStatusRunning] > 0 {
		return apistructs.StatusDesc{
			Status:      apistructs.StatusProgressing,
			LastMessage: "health check not ready",
		}
	}
	return apistructs.StatusDesc{
		Status:      apistructs.StatusUnknown,
		LastMessage: "count of running services not match",
	}
}

func (m *Marathon) probeProxyPortsForCreate(method string, runtime *apistructs.ServiceGroup) error {
	// construct marathon Group entity
	mGroup, err := m.buildMarathonGroup(*runtime)
	if err != nil {
		return errors.Wrap(err, "build marathon group failed")
	}
	for i := range mGroup.Apps {
		mGroup.Apps[i].Instances = 0
	}
	_, err = m.putGroup(method, mGroup, runtime.Force)
	return err
}

func (m *Marathon) resolveProxyPorts(runtime *apistructs.ServiceGroup) error {
	mGroupId := buildMarathonGroupId(m.prefix, runtime.Type, runtime.ID)
	mGroup, err := m.getGroup(mGroupId)
	if err != nil {
		return err
	}
	mp := make(map[string][]int)
	for _, mApp := range mGroup.Apps {
		var portMappings []AppContainerPortMapping
		if lessThan(m.version, Ver{1, 5, 0}) {
			portMappings = mApp.Container.Docker.PortMappings
		} else {
			portMappings = mApp.Container.PortMappings
		}
		proxyPorts := make([]int, 0)
		for _, mapping := range portMappings {
			proxyPorts = append(proxyPorts, mapping.ServicePort)
		}
		mp[mApp.Id] = proxyPorts
	}
	for i := range runtime.Services {
		mAppId := buildMarathonAppId(mGroupId, runtime.Services[i].Name)
		runtime.Services[i].ProxyIp = defaultInternalLb
		runtime.Services[i].ProxyPorts = mp[mAppId]
		if runtime.Services[i].Labels["X_ENABLE_PUBLIC_IP"] == "true" {
			runtime.Services[i].PublicIp = m.pubIp
		}
	}
	return nil
}

func (m *Marathon) putGroup(method string, mGroup *Group, force bool) (*GroupPutResponse, error) {
	// do POST (create) or PUT (update) over marathon api
	var req *httpclient.Request
	if method == http.MethodPost {
		req = m.client.Post(m.addr).Path("/v2/groups/" + mGroup.Id).JSONBody(mGroup)
	} else if method == http.MethodPut {
		req = m.client.Put(m.addr).Path("/v2/groups/" + mGroup.Id).JSONBody(mGroup)
	} else {
		return nil, errors.Errorf("method(%v) not supported, only POST or PUT available", method)
	}

	if force {
		// + "?force=true"
		req = req.Param("force", "true")
	}
	var obj GroupPutResponse
	resp, err := req.Do().JSON(&obj)
	if err != nil {
		return nil, errors.Wrapf(err, "marathon schedule service: %s", mGroup.Id)
	}
	// handle response
	return processGroupPutResponse(&obj, resp.StatusCode())
}

func (m *Marathon) deleteGroup(mGroupId string, force bool) (*GroupPutResponse, error) {
	req := m.client.Delete(m.addr).Path("/v2/groups/" + mGroupId)
	if force {
		req = req.Param("force", "true")
	}
	var obj GroupPutResponse
	resp, err := req.Do().JSON(&obj)
	if err != nil {
		if resp.IsNotfound() {
			return &obj, nil
		}
		return nil, errors.Wrapf(err, "marathon delete failed, groupId: %s", mGroupId)
	}

	// If not found, directly return to success
	if resp.IsNotfound() {
		return &obj, nil
	}

	// handle response
	return processGroupPutResponse(&obj, resp.StatusCode())
}

func (m *Marathon) getGroupWithDefaultParam(groupID string) (*Group, error) {
	return m.getGroup(groupID,
		"group.apps",
		"group.apps.tasks",
		"group.apps.counts",
		"group.apps.deployments",
		"group.apps.readiness",
		"group.apps.lastTaskFailure",
		"group.apps.taskStats")
}

func (m *Marathon) getGroup(mGroupId string, embed ...string) (*Group, error) {
	req := m.client.Get(m.addr).Path("/v2/groups/" + mGroupId)
	for _, em := range embed {
		req = req.Param("embed", em)
	}
	var obj GroupHTTPResult
	resp, err := req.Do().JSON(&obj)
	if err != nil {
		return nil, errors.Wrapf(err, "marathon get group failed, groupId: %s", mGroupId)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return &(obj.Group), nil
	default:
		return nil, errors.New(obj.GetErrorResponse.Message)
	}
}

func (m *Marathon) getQueue() (*Queue, error) {
	var obj QueueHTTPResult

	resp, err := m.client.Get(m.addr).Path("/v2/queue").Do().JSON(&obj)
	if err != nil {
		return nil, errors.Wrap(err, "marathon get queue failed")
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return &(obj.Queue), nil
	default:
		return nil, errors.New(obj.GetErrorResponse.Message)
	}
}

func processGroupPutResponse(resp *GroupPutResponse, code int) (*GroupPutResponse, error) {
	switch code {
	case http.StatusOK, http.StatusCreated:
		// TODO: returns unified version
		return resp, nil
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusConflict,
		http.StatusUnprocessableEntity:
		logrus.Errorf("marathon put group failed, error=%v", resp)
		return nil, errors.New(resp.ToString())
	default:
		logrus.Errorf("marathon put group failed, error=%v", resp)
		return nil, errors.New(resp.ToString())
	}
}

func (m *Marathon) buildMarathonGroup(runtime apistructs.ServiceGroup) (*Group, error) {
	// marathon GroupId
	mGroupId := buildMarathonGroupId(m.prefix, runtime.Type, runtime.ID)
	fetchUris, ok := m.options["FETCHURIS"]
	if !ok {
		fetchUris = defaultFetchUris
	}
	var mFetch []AppFetch
	if fetchUris != "" {
		uris := strings.Split(fetchUris, ",")
		for _, f := range uris {
			mFetch = append(mFetch, AppFetch{
				Uri: f,
			})
		}
	}

	// get cluster info
	clusterInfo, err := m.clusterInfo.Info(m.clusterName)
	if err != nil {
		return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)",
			m.clusterName, err)
	}

	// build apps first
	mApps := make([]App, len(runtime.Services))
	for i, service := range runtime.Services {
		// marathon AppId
		mAppId := buildMarathonAppId(mGroupId, service.Name)
		// appLevelPart VIP
		mAppLevelVip := buildMarathonVipAppLevelPart(mAppId)
		mVip := buildMarathonVip(mAppLevelVip)
		// save back vip into service spec for next work: depends resolve
		runtime.Services[i].Vip = mVip
		mApps[i] = App{
			Id:        mAppId,
			Instances: service.Scale,
			Cmd:       service.Cmd,
			Cpus:      service.Resources.Cpu,
			Mem:       service.Resources.Mem,
			Disk:      service.Resources.Disk,
			Container: AppContainer{
				Type: defaultContainerType,
				Docker: AppContainerDocker{
					Image: service.Image,
					// TODO: support a parameter to control force pull image
					ForcePullImage: false,
				},
			},
			Dependencies: convertDepends(mGroupId, service.Depends),
			Labels:       service.Labels,
			// backoff config: initial 10 seconds, with a factor 4, up to 3600 seconds
			BackoffSeconds:        15,
			BackoffFactor:         4,
			MaxLaunchDelaySeconds: 3600,
			// fetch uris config
			Fetch: mFetch,
		}

		m.addDockerConfg(&mApps[i])

		binds, err := convertBinds(service.Binds, clusterInfo)
		if err != nil {
			return nil, err
		}

		volumes, err := convertVolumes(service.Volumes, clusterInfo)
		if err != nil {
			return nil, err
		}

		mApps[i].Container.Volumes = append(binds, volumes...)

		// Set fine-grained CPU, including:
		// 1，Apply for cpu value
		// 2， cpu oversold
		// 3，Maximum cpu value
		cpulimit, err := m.setFineGrainedCPU(&mApps[i], runtime.Extra)
		if err != nil {
			return nil, err
		}

		if len(m.backoff) > 0 {
			if b, err := strconv.ParseFloat(m.backoff, 32); err == nil {
				mApps[i].BackoffFactor = float32(b)
				logrus.Debugf("executor(%s) app(%s) backoff factor set to %v", m.name, mApps[i].Id, b)
			}
		}

		if runtime.Labels["SERVICE_TYPE"] == "ADDONS" && m.addonsDisableAutoRestart {
			mApps[i].BackoffSeconds = 3600
			logrus.Debugf("executor(%s) app(%s) backoff seconds set to 3600", m.name, mApps[i].Id)
		}

		handleAddonsLabel(&runtime, &mApps[i])

		// scheduleInfo
		constrains := constructConstrains(&runtime.ScheduleInfo, &service)

		// constraints
		if constrains == nil {
			mApps[i].Constraints = []Constraint{}
		} else {
			mApps[i].Constraints = constrains
		}

		// emergency exit
		if v := os.Getenv("FORCE_OLD_LABEL_SCHEDULE"); v == "true" {
			serviceLabels := util.CombineLabels(runtime.Labels, service.Labels)
			builtConstraints := util.BuildDcosConstraints(m.enableTag,
				serviceLabels, m.preserveProjects, m.workspaceTags)
			for constraint := range builtConstraints {
				mApps[i].Constraints = append(mApps[i].Constraints,
					Constraint(builtConstraints[constraint]))
			}
		}

		if m.unique {
			mApps[i].Constraints = append(mApps[i].Constraints, []string{"hostname", "UNIQUE"})
		}

		// multiple version compatibility
		if lessThan(m.version, Ver{1, 5, 0}) {
			// check version lessThan 1.5.0
			mApps[i].Container.Docker.Network = defaultNetworkMode147
			mApps[i].Container.Docker.PortMappings =
				convertPortToPortMapping(diceyml.ComposeIntPortsFromServicePorts(service.Ports), mAppLevelVip)
			mApps[i].IpAddress = &AppIpAddress{
				NetworkName: defaultNetworkName,
				Labels:      make(map[string]string, 0),
				Discovery:   AppIpAddressDiscovery{Ports: []AppIpAddressDiscoveryPort{}},
				Groups:      []string{},
			}
		} else {
			mApps[i].Networks = []AppNetwork{
				{
					Mode: defaultNetworkMode,
					Name: defaultNetworkName,
				},
			}
			mApps[i].Container.PortMappings =
				convertPortToPortMapping(diceyml.ComposeIntPortsFromServicePorts(service.Ports), mAppLevelVip)
		}
		// TODO: add custom host using add-host
		vipDnsSearch := buildMarathonVipDnsSearch(mVip, service.Name)
		if vipDnsSearch != "" {
			mApps[i].Container.Docker.Parameters =
				append(mApps[i].Container.Docker.Parameters,
					AppContainerDockerParameter{
						Key: "dns-search", Value: vipDnsSearch,
					})
		}
		// TODO: refactor it, we just inject self_host (vip) to 127.0.0.1
		mApps[i].Container.Docker.Parameters =
			append(mApps[i].Container.Docker.Parameters,
				AppContainerDockerParameter{
					Key: "add-host", Value: mVip + ":127.0.0.1",
				}, AppContainerDockerParameter{
					Key: "add-host", Value: service.Name + ":127.0.0.1",
				})
		if service.Hosts != nil {
			addHosts := parseAddHost(service.Hosts)
			if addHosts != nil {
				for _, addHost := range addHosts {
					mApps[i].Container.Docker.Parameters =
						append(mApps[i].Container.Docker.Parameters, addHost)
				}
			}
		}
		// convert health check
		hc, err := convertHealthCheck(service, m.version)
		if err != nil {
			return nil, err
		}
		if hc != nil {
			mApps[i].HealthChecks = []AppHealthCheck{*hc}
		}
		// build basic env
		env := make(map[string]string)

		if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
			if mApps[i].Labels == nil {
				mApps[i].Labels = make(map[string]string)
			}
			e, ok := mApps[i].Labels["HAPROXY_GROUP"]
			if ok {
				if e == "external" {
					mApps[i].Labels["HAPROXY_GROUP"] = "external,internal"
				}
			} else {
				if runtime.Services[i].Labels["X_ENABLE_PUBLIC_IP"] == "true" && m.pubGrp != "internal" {
					mApps[i].Labels["HAPROXY_GROUP"] = m.pubGrp + ",internal"
				} else {
					mApps[i].Labels["HAPROXY_GROUP"] = "internal"
				}
			}
		}
		var selfIp string
		var selfPorts []int
		if runtime.ServiceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
			selfIp = runtime.Services[i].ProxyIp
			selfPorts = runtime.Services[i].ProxyPorts
		} else {
			selfIp = runtime.Services[i].Vip
			selfPorts = diceyml.ComposeIntPortsFromServicePorts(runtime.Services[i].Ports)
		}
		env["SELF_HOST"] = selfIp
		for i, port := range selfPorts {
			portStr := strconv.Itoa(port)
			if i == 0 {
				// special port env, SELF_PORT == SELF_PORT0
				env["SELF_PORT"] = portStr
				// TODO: we should deprecate this SELF_URL
				// TODO: we don't care what http is
				env["SELF_URL"] = "http://" + selfIp + ":" + portStr
			}
			env["SELF_PORT"+strconv.Itoa(i)] = portStr
		}
		env["DICE_CPU_ORIGIN"] = fmt.Sprintf("%f", service.Resources.Cpu)
		env["DICE_CPU_REQUEST"] = fmt.Sprintf("%f", mApps[i].Cpus)
		env["DICE_CPU_LIMIT"] = fmt.Sprintf("%f", cpulimit)
		env["DICE_MEM_ORIGIN"] = fmt.Sprintf("%f", service.Resources.Mem)
		env["DICE_MEM_REQUEST"] = fmt.Sprintf("%f", mApps[i].Mem)
		env["DICE_MEM_LIMIT"] = fmt.Sprintf("%f", mApps[i].Mem)
		env["IS_K8S"] = "false"

		mApps[i].Env = env
		m.addDockerEnv(&mApps[i], service.Env)

		// TODO: dirty approach for restart runtime, disable it later!
		lastRestartTime := runtime.Extra["lastRestartTime"]
		if lastRestartTime != "" {
			if mApps[i].Labels == nil {
				mApps[i].Labels = make(map[string]string)
			}
			mApps[i].Labels["LAST_RESTART_TIME"] = lastRestartTime
		}
	}
	// do resolve depends
	err = resolveDepends(&mApps, &runtime)
	if err != nil {
		return nil, err
	}
	// try expand env
	expandEnv(&mApps)
	// construct marathon Group entity
	mGroup := Group{
		Id:   mGroupId,
		Apps: mApps,
	}
	return &mGroup, nil
}

// an O(n^2) (n = len(mApps)) to resolve depends and injects env
func resolveDepends(mApps *[]App, runtime *apistructs.ServiceGroup) error {
	if len(*mApps) != len(runtime.Services) {
		return errors.New("illegal args, not match len of mApps and runtime.Services")
	}
	// build dependency env
	serviceTable := make(map[string]*apistructs.Service)
	mAppTable := make(map[string]*App)
	serviceNames := make([]string, 0)
	for i := range runtime.Services {
		service := &runtime.Services[i]
		serviceTable[service.Name] = service
		mAppTable[service.Name] = &(*mApps)[i]
		serviceNames = append(serviceNames, service.Name)
	}
	okSet := make(map[string]bool)
	for {
		var dead = true
		for i := range runtime.Services {
			service := &runtime.Services[i]
			if okSet[service.Name] {
				// skip if already resolved depends
				continue
			}
			mApp, exists := mAppTable[service.Name]
			if !exists {
				return errors.New("illegal args, no mApp exist though the relating service existing")
			}
			if containsAll(okSet, service.Depends) {
				dead = false
				okSet[service.Name] = true
				// following the additional work, to handle env injection
				var envInjectList []string
				if runtime.ServiceDiscoveryMode == "GLOBAL" {
					envInjectList = serviceNames
				} else {
					envInjectList = service.Depends
				}
				for _, dep := range envInjectList {
					depService := serviceTable[dep]
					env := resolveDependsEnv(runtime.ServiceDiscoveryKind, depService, service)
					for k, v := range env {
						mApp.Env[k] = v
					}
				}
			}
		}
		if dead {
			return errors.New("unresolved depends")
		}
		if len(okSet) == len(*mApps) {
			break
		}
	}
	return nil
}

// TODO: need refactor, remove `thisService`s
func resolveDependsEnv(serviceDiscoveryKind string,
	depService *apistructs.Service, thisService *apistructs.Service) map[string]string {
	depEnvPrefix := buildServiceDiscoveryEnvPrefix(depService.Name)
	var depIp string
	var depPorts []int
	if serviceDiscoveryKind == apistructs.ServiceDiscoveryKindProxy {
		depIp = depService.ProxyIp
		depPorts = depService.ProxyPorts
	} else {
		depIp = depService.Vip
		depPorts = diceyml.ComposeIntPortsFromServicePorts(depService.Ports)
	}
	env := make(map[string]string)
	env[depEnvPrefix+"_HOST"] = depIp
	for i, port := range depPorts {
		portStr := strconv.Itoa(port)
		if i == 0 {
			if thisService.Labels["IS_ENDPOINT"] == "true" {
				// check real in dep
				if containsInArray(thisService.Depends, depService.Name) {
					// TODO: we should deprecate BACKEND_URL usage
					env["BACKEND_URL"] = depIp + ":" + portStr
				}
			}
			env[depEnvPrefix+"_PORT"] = portStr
		}
		env[depEnvPrefix+"_PORT"+strconv.Itoa(i)] = portStr
	}
	return env
}

func expandEnv(mApps *[]App) {
	for i := range *mApps {
		env := &(*mApps)[i].Env
		for k, v := range *env {
			(*env)[k] = expandOneEnv(v, env)
		}
	}
}

func expandOneEnv(v string, env *map[string]string) string {
	for {
		lh := strings.Index(v, "${")
		if lh < 0 {
			break
		}
		rh := strings.Index(v, "}")
		if rh < 0 || !(lh+2 < rh) {
			// invalid env
			break
		}
		inner := v[lh+2 : rh]
		innerValue := (*env)[inner]
		v = v[:lh] + innerValue + v[rh+1:]
	}
	return v
}

func containsAll(set map[string]bool, values []string) bool {
	for _, v := range values {
		if !set[v] {
			return false
		}
	}
	return true
}

func containsInArray(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func (m *Marathon) buildGroupID(sg *apistructs.ServiceGroup) string {
	return buildMarathonGroupId(m.prefix, sg.Type, sg.ID)
}

func buildMarathonGroupId(prefix, namespace, name string) string {
	return strings.ToLower(prefix + "/" + namespace + "/" + name)
}

func buildMarathonAppId(mGroupId, serviceName string) string {
	return strings.ToLower(mGroupId + "/" + serviceName)
}

func buildMarathonVipAppLevelPart(mAppId string) string {
	var ret string
	for _, part := range strings.Split(mAppId, "/") {
		if part == "" {
			continue
		}
		if ret == "" {
			ret = part
		} else {
			ret = part + "." + ret
		}
	}
	return ret
}

// TODO: we should do build VIP Name in orchestrator
func buildMarathonStickyVipAppLevelPart(labels map[string]string) string {
	ret := labels["DICE_SERVICE_NAME"]
	runtimeName := labels["DICE_RUNTIME_NAME"]
	if !hasAnyPrefix(runtimeName, "dev.", "test.", "staging.", "prod.") {
		// compatible with old DICE_RUNTIME_NAME
		runtimeName = strings.Replace(
			strings.Replace(runtimeName, "-", ".", 1),
			"/", ".", -1)
	}
	ret += "." + strings.Replace(runtimeName, ".", "", -1) // be a part of domain
	ret += "." + labels["DICE_APPLICATION_NAME"]
	ret += "." + labels["DICE_PROJECT_NAME"]
	ret += "." + labels["DICE_ORG_NAME"]
	ret += ".runtimes"
	return strings.ToLower(ret)
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func buildMarathonVip(appLevelPartVip string) string {
	return appLevelPartVip + defaultVipSuffix
}

func buildMarathonVipDnsSearch(vip, serviceName string) string {
	cutFrom := len(serviceName) + 1 // plus 1 to remove the dot (.)
	if cutFrom > len(vip) {
		return ""
	}
	return vip[cutFrom:]
}

func buildServiceDiscoveryEnvPrefix(serviceName string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", "/", "").Replace(serviceName))
}

func convertPortToPortMapping(ports []int, mAppLevelVips ...string) []AppContainerPortMapping {
	mapping := make([]AppContainerPortMapping, len(ports))
	idx := -1
	for i, port := range ports {
		mapping[i] = AppContainerPortMapping{
			ContainerPort: port,
			// TODO: static servicePort
			// TODO: support config other protocol like udp
			Protocol: "tcp",
			// always assign a VIP whether use it or not
			Labels: make(map[string]string),
		}
		for _, mAppLevelVip := range mAppLevelVips {
			if len(mAppLevelVip) == 0 {
				continue
			}
			idx++
			mapping[i].Labels["VIP_"+strconv.Itoa(idx)] = mAppLevelVip + ":" + strconv.Itoa(port)
		}
	}
	return mapping
}

func convertDepends(mGroupId string, depends []string) []string {
	deps := make([]string, len(depends))
	for i, dep := range depends {
		deps[i] = buildMarathonAppId(mGroupId, dep)
	}
	return deps
}

func convertVolumes(volumes []apistructs.Volume, ciEnvs apistructs.ClusterInfoData) ([]AppContainerVolume, error) {
	appVolumes := []AppContainerVolume{}
	for _, volume := range volumes {
		switch volume.VolumeType {
		case apistructs.LocalVolume:
			appVolumes = append(appVolumes, AppContainerVolume{
				ContainerPath: volume.ContainerPath,
				HostPath:      volume.VolumePath,
				Mode:          "RW",
			}, AppContainerVolume{
				ContainerPath: volume.VolumePath,
				Mode:          "RW",
				Persistent: &apistructs.PersistentVolume{
					Type: "root",
					Size: volume.Size * 1024,
				},
			})
		case apistructs.NasVolume:
			hostPath, err := clusterinfo.ParseJobHostBindTemplate(volume.VolumePath, ciEnvs)
			if err != nil {
				return nil, err
			}

			appVolumes = append(appVolumes, AppContainerVolume{
				ContainerPath: volume.ContainerPath,
				HostPath:      hostPath,
				Mode:          "RW",
			})
		}
	}
	return appVolumes, nil

}
func convertBinds(binds []apistructs.ServiceBind, ciEnvs apistructs.ClusterInfoData) ([]AppContainerVolume, error) {
	appVolumes := []AppContainerVolume{}
	for _, bind := range binds {
		var mode string
		if bind.ReadOnly {
			mode = "RO"
		} else {
			mode = "RW"
		}

		hostPath, err := clusterinfo.ParseJobHostBindTemplate(bind.HostPath, ciEnvs)
		if err != nil {
			return nil, err
		}

		appVolumes = append(appVolumes, AppContainerVolume{
			ContainerPath: bind.ContainerPath,
			HostPath:      hostPath,
			Mode:          mode,
			// TODO: refactor it, remove it
			Persistent: bind.Persistent,
		})
	}
	return appVolumes, nil
}

// convert health check to marathon format
func convertHealthCheck(service apistructs.Service, ver Ver) (hc *AppHealthCheck, err error) {
	// dice Mandatory to set the health check less than 7 minutes to 7 minutes
	diceDuration := apistructs.HealthCheckDuration
	interval := 15

	newHC := service.NewHealthCheck
	if newHC != nil && (newHC.ExecHealthCheck != nil || newHC.HttpHealthCheck != nil) {
		logrus.Infof("in newhealthCheck for service(%s)", service.Name)
		hc = new(AppHealthCheck)
		hc.GracePeriodSeconds = 0
		// The interval of each health check
		hc.IntervalSeconds = interval
		hc.TimeoutSeconds = 10
		// The number of consecutive health check failures before the kill container
		hc.MaxConsecutiveFailures = diceDuration / interval
		// Wait for DelaySeconds seconds to start health check
		hc.DelaySeconds = 0

		if service.NewHealthCheck.HttpHealthCheck != nil {
			hc.Protocol = "MESOS_HTTP"
			hc.Path = service.NewHealthCheck.HttpHealthCheck.Path
			hc.Port = service.NewHealthCheck.HttpHealthCheck.Port
			if service.NewHealthCheck.HttpHealthCheck.Duration > diceDuration {
				hc.MaxConsecutiveFailures =
					service.NewHealthCheck.HttpHealthCheck.Duration / hc.IntervalSeconds
			}
		} else if service.NewHealthCheck.ExecHealthCheck != nil {
			hc.Protocol = HCMethodCommand
			hc.Command = &AppHealthCheckCommand{Value: service.NewHealthCheck.ExecHealthCheck.Cmd}
			if service.NewHealthCheck.ExecHealthCheck.Duration > diceDuration {
				hc.MaxConsecutiveFailures =
					service.NewHealthCheck.ExecHealthCheck.Duration / hc.IntervalSeconds
			}
		}
		return hc, nil
	}

	if service.HealthCheck == nil {
		hc = new(AppHealthCheck)
		if len(service.Ports) > 0 {
			if lessThan(ver, Ver{1, 5, 0}) {
				hc.Protocol = "TCP"
			} else {
				hc.Protocol = "MESOS_TCP"
			}
			hc.Port = service.Ports[0].Port
			hc.IntervalSeconds = interval
			hc.TimeoutSeconds = 10
			hc.MaxConsecutiveFailures = diceDuration / interval
			hc.DelaySeconds = 0
		} else {
			// Configure a default health check for services that are neither configured with health checks nor exposed ports
			hc.Protocol = HCMethodCommand
			hc.Command = &AppHealthCheckCommand{Value: "echo 1"}
			hc.MaxConsecutiveFailures = diceDuration / interval
		}
		return hc, nil
	}

	hc = new(AppHealthCheck)
	// allow app running hc.GracePeriodSeconds first, then we start health checking, and count failures
	hc.GracePeriodSeconds = 0
	hc.DelaySeconds = 0
	hc.IntervalSeconds = 15
	hc.TimeoutSeconds = 10
	hc.MaxConsecutiveFailures = 20

	switch service.HealthCheck.Kind {
	case "HTTP", "https":
		hc.Protocol = "MESOS_" + service.HealthCheck.Kind
		hc.Path = service.HealthCheck.Path
		hc.Port = service.HealthCheck.Port
	case "TCP":
		if lessThan(ver, Ver{1, 5, 0}) {
			hc.Protocol = HCMethodTCP
		} else {
			hc.Protocol = "MESOS_TCP"
		}
		hc.Port = service.HealthCheck.Port
	default:
		hc.Protocol = HCMethodCommand
		hc.Command = &AppHealthCheckCommand{Value: service.HealthCheck.Command}
	}
	return hc, err
}

func findPortIndex(ports []int, port int) (int, error) {
	for i, p := range ports {
		if p == port {
			return i, nil
		}
	}
	return -1, errors.New("port not found")
}

func parseVersion(strV string) (Ver, error) {
	v := make(Ver, 0)
	for _, x := range strings.Split(strV, ".") {
		i, err := strconv.Atoi(x)
		if err != nil {
			return nil, errors.Wrapf(err, "parse marathon version failed, %s", strV)
		}
		v = append(v, i)
	}
	return v, nil
}

func lessThan(v1, v2 Ver) bool {
	var l1, l2 = len(v1), len(v2)
	var minLen = l1
	if l2 < minLen {
		minLen = l2
	}
	for i := 0; i < minLen; i++ {
		if v1[i] == v2[i] {
			continue
		}
		return v1[i] < v2[i]
	}
	return l1 < l2
}

func parseAddHost(hosts []string) (addHosts []AppContainerDockerParameter) {
	for _, record := range hosts {
		parts := strings.SplitN(record, " ", 2)
		if len(parts) != 2 {
			// fail tolerance
			logrus.Warnf("failed to parse add-host")
			continue
		}
		addHosts = append(addHosts, AppContainerDockerParameter{"add-host", parts[1] + ":" + parts[0]})
	}
	return
}

// B -> MiB
func byteToMebibyte(b int64) int64 {
	return b / 1024 / 1024
}

func buildVolumeCloudHostPath(prefix string, mAppId string, containerPath string) string {
	var ret = prefix
	if !strings.HasPrefix(mAppId, "/") {
		ret += "/"
	}
	ret += mAppId
	if !strings.HasPrefix(containerPath, "/") {
		ret += "/"
	}
	ret += containerPath
	return ret
}

func (m *Marathon) addDockerConfg(app *App) {
	// Default value of each cluster
	// TODO: All done in etcd cluster configuration
	app.Container.Docker.Parameters = append(app.Container.Docker.Parameters,
		AppContainerDockerParameter{Key: "log-driver", Value: "json-file"},
		AppContainerDockerParameter{Key: "log-opt", Value: "max-size=100m"},
		AppContainerDockerParameter{Key: "log-opt", Value: "max-file=10"},
		AppContainerDockerParameter{Key: "cap-add", Value: "SYS_PTRACE"},
		// --init
		AppContainerDockerParameter{Key: "init", Value: "true"},
		// --ulimit
		// nofile: file descriptors
		// nproc:  processes
		AppContainerDockerParameter{Key: "ulimit", Value: "nofile=102400"},
		AppContainerDockerParameter{Key: "ulimit", Value: "nproc=8192"},
		//{Key: "security-opt", Value: "seccomp:unconfined"},
	)

	if len(m.dockerCfg) > 0 {
		configs := strings.Split(m.dockerCfg, ",")
		for _, cfg := range configs {
			if kv := strings.Split(cfg, "="); len(kv) == 2 {
				app.Container.Docker.Parameters = append(app.Container.Docker.Parameters,
					AppContainerDockerParameter{Key: kv[0], Value: kv[1]})
			}
		}
	}
}

func (m *Marathon) addDockerEnv(app *App, srvEnv map[string]string) {
	// force rotate size, default 2MB is too small
	app.Env["CONTAINER_LOGGER_LOGROTATE_MAX_STDERR_SIZE"] = "100MB"
	app.Env["CONTAINER_LOGGER_LOGROTATE_MAX_STDOUT_SIZE"] = "100MB"
	app.Env["CONTAINER_LOGGER_LOGROTATE_STDERR_OPTIONS"] = "rotate 9"
	app.Env["CONTAINER_LOGGER_LOGROTATE_STDOUT_OPTIONS"] = "rotate 9"

	if len(m.dockerEnv) > 0 {
		configs := strings.Split(m.dockerEnv, ",")
		for _, cfg := range configs {
			if kv := strings.Split(cfg, "="); len(kv) == 2 {
				app.Env[kv[0]] = kv[1]
			}
		}
	}

	for key, value := range srvEnv {
		app.Env[key] = value
	}
}

func (m *Marathon) setFineGrainedCPU(app *App, extra map[string]string) (float64, error) {
	// 1, Processing request cpu value
	requestCPU := app.Cpus
	if requestCPU < MIN_CPU_SIZE {
		return 0, errors.Errorf("app(%s) request cpu is %v, which is lower than min cpu(%v)",
			app.Id, requestCPU, MIN_CPU_SIZE)
	}

	// 2, Dealing with cpu oversold
	ratio := cpupolicy.CalcCPUSubscribeRatio(m.cpuSubscribeRatio, extra)
	app.Cpus = requestCPU / ratio

	// 3, Processing the maximum cpu, that is, the corresponding cpu quota, the default is not limited cpu quota, that is, the value corresponding to cpu.cfs_quota_us under the cgroup is -1
	quota := int64(0)

	// Set the maximum cpu according to the requested cpu
	if m.cpuNumQuota == -1.0 {
		maxCPU := cpupolicy.AdjustCPUSize(requestCPU)
		// Set the cpu value corresponding to cpu quota to maxCPU
		quota = int64(maxCPU * float64(CPU_CFS_PERIOD_US))
	} else if m.cpuNumQuota > 0 {
		quota = int64(m.cpuNumQuota * float64(CPU_CFS_PERIOD_US))
	}

	app.Container.Docker.Parameters = append(
		app.Container.Docker.Parameters,
		AppContainerDockerParameter{"cpu-quota", strconv.FormatInt(quota, 10)},
	)
	logrus.Debugf("app(%s) set cpu from %v to %v, subscribe ratio: %v, cpu quota: %v",
		app.Id, requestCPU, app.Cpus, ratio, quota)
	return float64(quota) / float64(CPU_CFS_PERIOD_US), nil
}

func handleAddonsLabel(runtime *apistructs.ServiceGroup, app *App) {
	if runtime.Labels["SERVICE_TYPE"] != "ADDONS" {
		return
	}
	if app.Labels == nil {
		app.Labels = make(map[string]string)
	}
	if app.Env == nil {
		app.Env = make(map[string]string)
	}
	for k, v := range runtime.Labels {
		// Copy the label prefixed with DICE on the runtime to the label and env of the service
		if strings.HasPrefix(k, "DICE") {
			app.Labels[k] = v
			app.Env[k] = v
		}
	}
}

// Set specific information about insufficient service resources
func setServiceDetailedResourceInfo(service *apistructs.Service, queue *Queue, appID string, status AppStatus) {
	if status != AppStatusWaiting {
		return
	}
	unScheduledReasons := &service.StatusDesc.UnScheduledReasons
	for _, offer := range queue.Queue {
		if appID == offer.App.Id && offer.Delay.Overdue {
			for _, lastOffer := range offer.ProcessedOffersSummary.RejectSummaryLastOffers {
				// All the submitted offers do not meet the conditions, indicating that the resources are lacking
				if lastOffer.Declined == lastOffer.Processed && lastOffer.Processed > 0 {
					unScheduledReasons.AddResourceInfo(lastOffer.Reason)
					logrus.Infof("service(%s) has resource insufficiency: %v",
						service.Name, unScheduledReasons)
				}
			}
			break
		}
	}
}

// TODO: This function needs to be refactored
func constructConstrains(r *apistructs.ScheduleInfo, service *apistructs.Service) []Constraint {
	var constrains []Constraint

	cons := NewConstraints()
	location, ok := r.Location[service.Name].(diceyml.Selector)
	if !ok {
		cons.NewUnlikeRule(labelconfig.DCOS_ATTRIBUTE).
			OR(AND(apistructs.TagLocationOnly))
	} else if location.Not {
		rule := cons.NewUnlikeRule(labelconfig.DCOS_ATTRIBUTE)
		fullLocations := strutil.Map(location.Values,
			func(s string) string { return strutil.Concat(apistructs.TagLocationPrefix, s) })
		for _, l := range fullLocations {
			rule.OR(AND(l))
		}
	} else {
		rule := cons.NewLikeRule(labelconfig.DCOS_ATTRIBUTE)
		fullLocations := strutil.Map(location.Values,
			func(s string) string { return strutil.Concat(apistructs.TagLocationPrefix, s) })
		for _, l := range fullLocations {
			rule.OR(AND(l))
		}
	}
	constrains = append(constrains, cons.Generate()...)

	if r.IsPlatform {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + apistructs.TagPlatform + `\b.*`})
		if r.IsUnLocked {
			constrains = append(constrains,
				[]string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + apistructs.TagLocked + `\b.*`})
		}
		return constrains
	}

	// Do not schedule to the node marked with the prefix of the label
	for _, unlikePrefix := range r.UnLikePrefixs {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + unlikePrefix + `[^,]+\b.*`})
	}
	// Not scheduled to the node with this label
	unlikes := []string{}
	copy(unlikes, r.UnLikes)
	if !r.IsPlatform {
		unlikes = append(unlikes, apistructs.TagPlatform)
	}
	if r.IsUnLocked {
		unlikes = append(unlikes, apistructs.TagLocked)
	}
	for _, unlike := range unlikes {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + unlike + `\b.*`})
	}
	// Specify scheduling to the node labeled with the prefix
	// Currently no such label
	for _, likePrefix := range r.LikePrefixs {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + likePrefix + `\b.*`})
	}
	// Specify to be scheduled to the node with this label, not coexisting with any
	for _, exclusiveLike := range r.ExclusiveLikes {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + exclusiveLike + `\b.*`})
	}
	// Specify to be scheduled to the node with this label, if the any label is enabled, the any label is attached
	for _, like := range r.Likes {
		if r.Flag {
			constrains = append(constrains,
				[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE",
					`.*\b` + apistructs.TagAny + `\b.*|.*\b` + like + `\b.*`})
		} else {
			constrains = append(constrains,
				[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + like + `\b.*`})
		}
	}

	// Specify scheduling to the node with this label, allowing multiple OR operations
	if len(r.InclusiveLikes) > 0 {
		constrain := []string{labelconfig.DCOS_ATTRIBUTE, "LIKE"}
		var sentence string
		for i, inclusiveLike := range r.InclusiveLikes {
			if i == len(r.InclusiveLikes)-1 {
				sentence = sentence + `.*\b` + inclusiveLike + `\b.*`
				constrain = append(constrain, sentence)
				constrains = append(constrains, constrain)
			} else {
				sentence = sentence + `.*\b` + inclusiveLike + `\b.*|`
			}
		}
	}

	if len(r.SpecificHost) != 0 {
		constrains = append(constrains,
			Constraint([]string{"hostname", "LIKE", strings.Join(r.SpecificHost, "|")}))
	}

	return constrains
}
func (*Marathon) CleanUpBeforeDelete() {}
func (*Marathon) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for marathon")
}

func (*Marathon) KillPod(podname string) error {
	return fmt.Errorf("not support for marathon")
}

func (*Marathon) Scale(ctx context.Context, spec interface{}) (interface{}, error) {
	return apistructs.ServiceGroup{}, fmt.Errorf("scale not support for marathon")
}
