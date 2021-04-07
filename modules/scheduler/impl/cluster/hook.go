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

package cluster

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	// ErrInvalidAction 不合法的 clusterevent 的 action
	ErrInvalidAction = errors.New("invalid cluster event action")
	// ErrInvalidClusterType 不合法的 cluster type
	ErrInvalidClusterType = errors.New("invalid cluster type")
	// ErrEmptyClusterName cluster name 为空
	ErrEmptyClusterName = errors.New("empty cluster name")
	// ErrClusterAddrNotSet cluster addr 未设置
	ErrClusterAddrNotSet = errors.New("cluster addr not set")
)

// Hook 接收 clusterevent， 见 `apistructs.ClusterEvent`
func (c *ClusterImpl) Hook(clusterEvent *apistructs.ClusterEvent) error {
	if !strutil.Equal(clusterEvent.Content.Type, clusterTypeDcos, true) {
		if strutil.Equal(clusterEvent.Content.Type, clusterTypeLocaldocker, true) {
			return c.createLocalDocker(clusterEvent)
		}
		if strutil.Equal(clusterEvent.Content.Type, clusterTypeK8S, true) {
			return c.handleK8SCluster(clusterEvent)
		}
		if strutil.Equal(clusterEvent.Content.Type, clusterTypeEdas, true) {
			return c.handleEdasCluster(clusterEvent)
		}
		return errors.Wrap(ErrInvalidClusterType, clusterEvent.Content.Type)
	}
	if clusterEvent.Content.Name == "" {
		return errors.Wrap(ErrEmptyClusterName, fmt.Sprintf("%+v", clusterEvent.Content))
	}
	logrus.Infof("cluster event: %+v, scheduler config: %+v", clusterEvent, *(clusterEvent.Content.SchedConfig))

	if clusterEvent.Action == clusterActionCreate {
		cluster := generateClusterInfoFromEvent(&clusterEvent.Content)
		addr, ok := cluster.Options["ADDR"]
		if !ok {
			return ErrClusterAddrNotSet
		}

		srv := cluster
		if err := createMarathonExecutor(c.js, srv, addr); err != nil {
			return err
		}

		job := cluster
		if err := createMetronomeExecutor(c.js, job, addr); err != nil {
			return err
		}

		logrus.Infof("cluster(%s) %s successfully according to related event", clusterEvent.Action, cluster.ClusterName)
		return nil
	} else if clusterEvent.Action == clusterActionUpdate {
		keys, err := c.findKeysByCluster(clusterEvent.Content.Name)
		if err != nil {
			return err
		}

		for _, k := range keys {
			cluster := ClusterInfo{}
			if err := c.js.Get(context.Background(), k, &cluster); err != nil {
				content := fmt.Sprintf("get cluster(%s, key: %s) from etcd error: %v", clusterEvent.Content.Name, k, err)
				return errors.New(content)
			}
			if err := patchClusterConfig(&cluster, &(clusterEvent.Content)); err != nil {
				content := fmt.Sprintf("parse cluster(%s, key: %s) url error: %v", clusterEvent.Content.Name, k, err)
				return errors.New(content)
			}
			if err := c.js.Put(context.Background(), k, cluster); err != nil {
				content := fmt.Sprintf("patch cluster(%s, key: %s) to etcd error: %v", clusterEvent.Content.Name, k, err)
				return errors.New(content)
			}
		}
		return nil
	}
	return errors.Wrap(ErrInvalidAction, clusterEvent.Action)
}

func (c *ClusterImpl) findKeysByCluster(name string) ([]string, error) {
	keys := make([]string, 0)
	findKeyBelongedToCluster := func(k string, v interface{}) error {
		c, ok := v.(*ClusterInfo)
		if !ok {
			logrus.Errorf("key(%s) related value(%+v) is not type of ClusterInfo", k, v)
			return nil
		}
		if c.ClusterName == name {
			keys = append(keys, k)
		}
		return nil
	}

	if err := c.js.ForEach(context.Background(), clusterPrefix, ClusterInfo{}, findKeyBelongedToCluster); err != nil {
		content := fmt.Sprintf("cluster(%s) range prefix(%s) from etcd got err: %v", name, clusterPrefix, err)
		return keys, errors.New(content)
	}

	if len(keys) == 0 {
		content := fmt.Sprintf("Cannot find executor belonged to cluster(%s)", name)
		return keys, errors.New(content)
	}
	return keys, nil
}

func setDefaultClusterConfig(ci *ClusterInfo) {
	t := "true"
	if _, ok := ci.Options[labelconfig.ENABLETAG]; !ok {
		ci.Options[labelconfig.ENABLETAG] = t
	}
	if _, ok := ci.Options[labelconfig.ENABLE_ORG]; !ok {
		ci.Options[labelconfig.ENABLE_ORG] = t
	}
	if _, ok := ci.Options[labelconfig.ENABLE_WORKSPACE]; !ok {
		ci.Options[labelconfig.ENABLE_WORKSPACE] = t
	}
	if _, ok := ci.Options[labelconfig.CPU_NUM_QUOTA]; !ok {
		// 这里的 "-1" 指按照用户申请的真实cpu来限制其最大cpu，
		// 即如果用户申请的是1个核的cpu，则quota对应的就是1个核
		// 如果设置为 "0"，则cpu quota无限制
		// 具体请参考 marathon 插件中的实现（setFineGrainedCPU 函数中）
		ci.Options[labelconfig.CPU_NUM_QUOTA] = "-1"
	}
}

func create(js jsonstore.JsonStore, key string, c ClusterInfo) error {
	setDefaultClusterConfig(&c)
	if err := js.Put(context.Background(), key, c); err != nil {
		content := fmt.Sprintf("write key(%s) to etcd error: %v", key, err)
		return errors.New(content)
	}

	logrus.Infof("cluster executor(%s) created successfully, content: %+v", c.ExecutorName, c)
	return nil
}

func createEdasExector(js jsonstore.JsonStore, key string, c string) error {
	if err := js.Put(context.Background(), key, c); err != nil {
		content := fmt.Sprintf("write key(%s) to etcd error: %v", key, err)
		return errors.New(content)
	}

	logrus.Infof("cluster executor(%s) created successfully", c)
	return nil
}

func createMarathonExecutor(js jsonstore.JsonStore, cluster ClusterInfo, addr string) error {
	srv := cluster
	srv.Kind = clusterutil.ServiceKindMarathon
	srv.ExecutorName = clusterutil.GenerateExecutorByCluster(cluster.ClusterName, clusterutil.ServiceKindMarathon)
	srv.Options["ADDR"] = addr + marathonAddrSuffix
	// 默认限制最大cpu(cpu quota)
	if _, ok := srv.Options["CPU_NUM_QUOTA"]; !ok {
		// 这里的 "-1" 指按照用户申请的真实cpu来限制其最大cpu，
		// 即如果用户申请的是1个核的cpu，则quota对应的就是1个核
		// 如果设置为 "0"，则cpu quota无限制
		// 具体请参考 marathon 插件中的实现（setFineGrainedCPU 函数中）
		srv.Options["CPU_NUM_QUOTA"] = "-1"
	}

	return create(js, clusterPrefix+cluster.ClusterName+clusterMarathonSuffix, srv)
}

func createMetronomeExecutor(js jsonstore.JsonStore, cluster ClusterInfo, addr string) error {
	job := cluster
	job.Kind = clusterutil.JobKindMetronome
	job.ExecutorName = clusterutil.GenerateExecutorByCluster(cluster.ClusterName, clusterutil.JobKindMetronome)
	job.Options["ADDR"] = addr + metronomeAddrSuffix

	return create(js, clusterPrefix+cluster.ClusterName+clusterMetronomeSuffix, job)
}

func generateClusterInfoFromEvent(eventCluster *apistructs.ClusterInfo) ClusterInfo {
	configCluster := ClusterInfo{
		ClusterName:  eventCluster.Name,
		ExecutorName: clusterutil.GenerateExecutorByCluster(eventCluster.Name, clusterutil.ServiceKindMarathon),
		Kind:         clusterutil.ServiceKindMarathon,
		Options:      make(map[string]string),
	}
	if eventCluster.SchedConfig == nil {
		return configCluster
	}

	configCluster.Options["ADDR"] = eventCluster.SchedConfig.MasterURL
	configCluster.Options["ENABLETAG"] = strconv.FormatBool(eventCluster.SchedConfig.EnableTag)
	if eventCluster.SchedConfig.AuthType == "basic" {
		configCluster.Options["BASICAUTH"] = eventCluster.SchedConfig.AuthUsername + ":" + eventCluster.SchedConfig.AuthPassword
	}
	if len(eventCluster.SchedConfig.CACrt) > 0 {
		configCluster.Options["CA_CRT"] = eventCluster.SchedConfig.CACrt
		configCluster.Options["CLIENT_CRT"] = eventCluster.SchedConfig.ClientCrt
		configCluster.Options["CLIENT_KEY"] = eventCluster.SchedConfig.ClientKey
	}
	return configCluster
}

// TODO: 需要讨论开放给用户修改的内容
// TODO: 支持 patch 精细化配置
func patchClusterConfig(local *ClusterInfo, request *apistructs.ClusterInfo) error {
	if request.SchedConfig != nil {
		if request.SchedConfig.MasterURL != "" {
			u, err := url.Parse(request.SchedConfig.MasterURL)
			if err != nil {
				return errors.Errorf("cluster addr(%s) is invalid url", request.SchedConfig.MasterURL)
			}
			if local.Kind == clusterutil.ServiceKindMarathon {
				u.Path = path.Join(u.Path, marathonAddrSuffix)
				local.Options["ADDR"] = u.String()
			} else if local.Kind == clusterutil.JobKindMetronome {
				u.Path = path.Join(u.Path, metronomeAddrSuffix)
				local.Options["ADDR"] = u.String()
			}
		}
		if request.SchedConfig.AuthUsername != "" && request.SchedConfig.AuthPassword != "" {
			local.Options["BASICAUTH"] = request.SchedConfig.AuthUsername + ":" + request.SchedConfig.AuthPassword
		}
		if request.SchedConfig.CACrt != "" && request.SchedConfig.ClientCrt != "" && request.SchedConfig.ClientKey != "" {
			local.Options["CA_CRT"] = request.SchedConfig.CACrt
			local.Options["CLIENT_CRT"] = request.SchedConfig.ClientCrt
			local.Options["CLIENT_KEY"] = request.SchedConfig.ClientKey
		}
	}
	return nil
}

// 临时代码，用于支持 localdocker 的 executor 创建
func (c *ClusterImpl) createLocalDocker(clusterEvent *apistructs.ClusterEvent) error {
	srv := ClusterInfo{
		ClusterName: clusterEvent.Content.Name,
		Kind:        "LOCALDOCKER",
		// 用 MARATHON 类型模拟生成 localdocker 的服务类型 executor
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.ServiceKindMarathon),
	}
	key := clusterPrefix + clusterEvent.Content.Name + clusterMarathonSuffix
	if err := c.js.Put(context.Background(), key, srv); err != nil {
		content := fmt.Sprintf("write key(%s) to etcd error: %v", key, err)
		return errors.New(content)
	}

	job := ClusterInfo{
		ClusterName: clusterEvent.Content.Name,
		Kind:        "LOCALJOB",
		// 用 METRONOME 类型模拟生成 localdocker 的 job 类型 executor
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.JobKindMetronome),
	}
	key = clusterPrefix + clusterEvent.Content.Name + clusterMetronomeSuffix
	if err := c.js.Put(context.Background(), key, job); err != nil {
		content := fmt.Sprintf("write key(%s) to etcd error: %v", key, err)
		return errors.New(content)
	}

	logrus.Infof("successed to create local cluster, name: %s", clusterEvent.Content.Name)
	return nil
}

func (c *ClusterImpl) handleK8SCluster(clusterEvent *apistructs.ClusterEvent) error {
	logrus.Infof("cluster hook of k8s, name: %s", clusterEvent.Content.Name)

	if clusterEvent.Content.SchedConfig == nil {
		return errors.Wrap(ErrInvalidClusterType, clusterEvent.Content.Type)
	}

	if clusterEvent.Action == clusterActionCreate {
		return c.createK8SExecutor(clusterEvent)
	} else if clusterEvent.Action == clusterActionUpdate {
		return c.updateK8SExecutor(clusterEvent)
	}
	return nil
}

func (c *ClusterImpl) createK8SExecutor(clusterEvent *apistructs.ClusterEvent) error {
	serviceCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.ServiceKindMarathon),
		Kind:         clusterutil.ServiceKindK8S,
		Options: map[string]string{
			"cluster":                     clusterEvent.Content.Name,
			"ADDR":                        clusterEvent.Content.SchedConfig.MasterURL,
			"DEV_CPU_SUBSCRIBE_RATIO":     clusterEvent.Content.SchedConfig.CPUSubscribeRatio,
			"TEST_CPU_SUBSCRIBE_RATIO":    clusterEvent.Content.SchedConfig.CPUSubscribeRatio,
			"STAGING_CPU_SUBSCRIBE_RATIO": clusterEvent.Content.SchedConfig.CPUSubscribeRatio,
		},
	}
	setDefaultClusterConfig(&serviceCfg)

	svcPath := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SSuffix)
	if err := create(c.js, svcPath, serviceCfg); err != nil {
		logrus.Errorf("failed to create k8s service executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	jobCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.JobKindMetronome),
		Kind:         clusterutil.JobKindK8S,
		Options: map[string]string{
			"ADDR": clusterEvent.Content.SchedConfig.MasterURL,
		},
	}
	setDefaultClusterConfig(&jobCfg)

	jobPath := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SJobSuffix)
	if err := create(c.js, jobPath, jobCfg); err != nil {
		logrus.Errorf("failed to create k8s job executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	if err := createK8SFlinkExecutor(clusterEvent, c.js); err != nil {
		logrus.Errorf("failed to create k8s flink executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	if err := createK8SSparkExecutor(clusterEvent, c.js); err != nil {
		logrus.Errorf("failed to create k8s spark executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	logrus.Infof("succeed to create k8s cluster, name: %s", clusterEvent.Content.Name)
	return nil
}

func (c *ClusterImpl) updateK8SExecutor(clusterEvent *apistructs.ClusterEvent) error {
	keys, err := c.findKeysByCluster(clusterEvent.Content.Name)
	if err != nil {
		return err
	}
	for _, k := range keys {
		cluster := ClusterInfo{}
		if err := c.js.Get(context.Background(), k, &cluster); err != nil {
			content := fmt.Sprintf("get cluster(%s, key: %s) from etcd error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
		if err := patchK8SConfig(&cluster, &(clusterEvent.Content)); err != nil {
			content := fmt.Sprintf("parse cluster(%s, key: %s) url error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
		if err := c.js.Put(context.Background(), k, cluster); err != nil {
			content := fmt.Sprintf("patch cluster(%s, key: %s) to etcd error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
	}
	return nil
}

// todo: 暂只支持修改地址
func patchK8SConfig(local *ClusterInfo, request *apistructs.ClusterInfo) error {
	if request.SchedConfig != nil {
		if request.SchedConfig.MasterURL != "" {
			u, err := url.Parse(request.SchedConfig.MasterURL)
			if err != nil {
				return errors.Errorf("k8s cluster addr is invalid, addr: %s", request.SchedConfig.MasterURL)
			}
			local.Options["ADDR"] = u.String()
			c, err := url.Parse(request.SchedConfig.CPUSubscribeRatio)
			if err != nil {
				return errors.Errorf("k8s cluster addr is invalid, addr: %s", request.SchedConfig.MasterURL)
			}
			local.Options["DEV_CPU_SUBSCRIBE_RATIO"] = c.String()
			local.Options["TEST_CPU_SUBSCRIBE_RATIO"] = c.String()
			local.Options["STAGING_CPU_SUBSCRIBE_RATIO"] = c.String()
		}
	}
	return nil
}

func (c *ClusterImpl) handleEdasCluster(clusterEvent *apistructs.ClusterEvent) error {
	logrus.Infof("cluster hook of edas, name: %s", clusterEvent.Content.Name)

	if clusterEvent.Content.SchedConfig == nil {
		return errors.Wrap(ErrInvalidClusterType, clusterEvent.Content.Type)
	}

	if clusterEvent.Action == clusterActionCreate {
		return c.createEdasExecutor(clusterEvent)
	} else if clusterEvent.Action == clusterActionUpdate {
		return c.updateEdasExecutor(clusterEvent)
	}
	return nil
}

func (c *ClusterImpl) createEdasExecutor(clusterEvent *apistructs.ClusterEvent) error {
	serviceExecutorName := clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.ServiceKindEdas)
	jobExecutorName := clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.JobKindK8S)
	addonExecutorName := clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.ServiceKindK8S)
	//写入edas service excutor
	svcExecutorPath := strutil.Concat(edasExecutorPrefix, clusterEvent.Content.Name, "/", clusterutil.ServiceKindMarathon)
	if err := createEdasExector(c.js, svcExecutorPath, serviceExecutorName); err != nil {
		logrus.Errorf("failed to create edas service executor, executor: %s", serviceExecutorName)
		return err
	}
	//写edas service 配置
	serviceCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: serviceExecutorName,
		Kind:         clusterutil.ServiceKindEdas,
		Options: map[string]string{
			"ADDR":            clusterEvent.Content.SchedConfig.EdasConsoleAddr,
			"ACCESSKEY":       clusterEvent.Content.SchedConfig.AccessKey,
			"ACCESSSECRET":    clusterEvent.Content.SchedConfig.AccessSecret,
			"CLUSTERID":       clusterEvent.Content.SchedConfig.ClusterID,
			"REGIONID":        clusterEvent.Content.SchedConfig.RegionID,
			"LOGICALREGIONID": clusterEvent.Content.SchedConfig.LogicalRegionID,
			"KUBEADDR":        clusterEvent.Content.SchedConfig.K8sAddr,
			"REGADDR":         clusterEvent.Content.SchedConfig.RegAddr,
		},
	}
	svcPath := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterEdasSuffix)
	if err := create(c.js, svcPath, serviceCfg); err != nil {
		logrus.Errorf("failed to create edas service executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}
	//写入edas job excutor
	jobExecutorPath := strutil.Concat(edasExecutorPrefix, clusterEvent.Content.Name, "/", clusterutil.JobKindMetronome)
	if err := createEdasExector(c.js, jobExecutorPath, jobExecutorName); err != nil {
		logrus.Errorf("failed to create edas service executor, executor: %s", jobExecutorName)
		return err
	}
	//写edas job 配置
	jobCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: jobExecutorName,
		Kind:         clusterutil.JobKindK8S,
		Options: map[string]string{
			"ADDR": clusterEvent.Content.SchedConfig.K8sAddr,
		},
	}
	setDefaultClusterConfig(&jobCfg)

	jobPath := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SJobSuffix)
	if err := create(c.js, jobPath, jobCfg); err != nil {
		logrus.Errorf("failed to create edas job executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	//写入edas addon excutor
	addonExecutorPath := strutil.Concat(edasExecutorPrefix, clusterEvent.Content.Name, "/", clusterutil.EdasKindInK8s)
	if err := createEdasExector(c.js, addonExecutorPath, addonExecutorName); err != nil {
		logrus.Errorf("failed to create edas service executor, executor: %s", addonExecutorName)
		return err
	}
	//写edas addon 配置
	addonCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: addonExecutorName,
		Kind:         clusterutil.ServiceKindK8S,
		Options: map[string]string{
			"ADDR":    clusterEvent.Content.SchedConfig.K8sAddr,
			"IS_EDAS": "true",
		},
	}
	setDefaultClusterConfig(&addonCfg)

	addonPath := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SSuffix)
	if err := create(c.js, addonPath, addonCfg); err != nil {
		logrus.Errorf("failed to create edas addon executor, cluster: %s", clusterEvent.Content.Name)
		return err
	}

	logrus.Infof("succeed to create edas cluster, name: %s", clusterEvent.Content.Name)
	return nil
}

func (c *ClusterImpl) updateEdasExecutor(clusterEvent *apistructs.ClusterEvent) error {
	keys, err := c.findKeysByCluster(clusterEvent.Content.Name)
	if err != nil {
		return err
	}
	for _, k := range keys {
		cluster := ClusterInfo{}
		if err := c.js.Get(context.Background(), k, &cluster); err != nil {
			content := fmt.Sprintf("get cluster(%s, key: %s) from etcd error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
		var pErr error
		if cluster.Kind == clusterutil.ServiceKindEdas {
			pErr = patchEdasConfig(&cluster, &(clusterEvent.Content))
		} else {
			pErr = patchK8SConfig(&cluster, &(clusterEvent.Content))
		}
		if pErr != nil {
			content := fmt.Sprintf("parse cluster(%s, key: %s) url error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
		if err := c.js.Put(context.Background(), k, cluster); err != nil {
			content := fmt.Sprintf("patch cluster(%s, key: %s) to etcd error: %v", clusterEvent.Content.Name, k, err)
			return errors.New(content)
		}
	}
	return nil
}

// 修改edas调度器
func patchEdasConfig(local *ClusterInfo, request *apistructs.ClusterInfo) error {
	if request.SchedConfig != nil {
		local.Options["ADDR"] = request.SchedConfig.EdasConsoleAddr
		local.Options["ACCESSKEY"] = request.SchedConfig.AccessKey
		local.Options["ACCESSSECRET"] = request.SchedConfig.AccessSecret
		local.Options["CLUSTERID"] = request.SchedConfig.ClusterID
		local.Options["REGIONID"] = request.SchedConfig.RegionID
		local.Options["LOGICALREGIONID"] = request.SchedConfig.LogicalRegionID
		local.Options["KUBEADDR"] = request.SchedConfig.K8sAddr
		local.Options["REGADDR"] = request.SchedConfig.RegAddr
	}
	return nil
}

func createK8SFlinkExecutor(clusterEvent *apistructs.ClusterEvent, js jsonstore.JsonStore) error {
	flinkCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.K8SKindFlink),
		Kind:         clusterutil.K8SKindFlink,
		Options: map[string]string{
			"ADDR": clusterEvent.Content.SchedConfig.MasterURL,
		},
		OptionsPlus: nil,
	}
	flinkKey := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SFlinkSuffix)
	return create(js, flinkKey, flinkCfg)
}

func createK8SSparkExecutor(clusterEvent *apistructs.ClusterEvent, js jsonstore.JsonStore) error {
	sparkCfg := ClusterInfo{
		ClusterName:  clusterEvent.Content.Name,
		ExecutorName: clusterutil.GenerateExecutorByCluster(clusterEvent.Content.Name, clusterutil.K8SKindSpark),
		Kind:         clusterutil.K8SKindSpark,
		Options: map[string]string{
			"ADDR":          clusterEvent.Content.SchedConfig.MasterURL,
			"SPARK_VERSION": clusterutil.K8SSparkVersion,
		},
		OptionsPlus: nil,
	}
	sparkKey := strutil.Concat(clusterPrefix, clusterEvent.Content.Name, clusterK8SSparkSuffix)
	return create(js, sparkKey, sparkCfg)
}
