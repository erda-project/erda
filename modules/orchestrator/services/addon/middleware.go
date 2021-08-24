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

package addon

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/strutil"
)

func buildMiddlewareFilter(instanceinfo apistructs.InstanceInfoDataList) (addonids []string) {
	addonids = []string{}
	for _, ins := range instanceinfo {
		if ins.AddonID != "" {
			addonids = append(addonids, ins.AddonID)
		}
	}
	return
}

// ListMiddleware 获取 middleware 列表
func (a *Addon) ListMiddleware(orgID uint64, params *apistructs.MiddlewareListRequest) (*apistructs.MiddlewareListResponseData, error) {
	var limited_addonids []string
	if params.InstanceIP != "" {
		instanceinfo, err := a.bdl.GetInstanceInfo(apistructs.InstanceInfoRequest{
			Phases:     []string{"unhealthy", "healthy", "running"},
			InstanceIP: params.InstanceIP,
		})
		if err != nil {
			return nil, err
		}
		if !instanceinfo.Success {
			return nil, fmt.Errorf(instanceinfo.Error.Msg)
		}
		limited_addonids = buildMiddlewareFilter(instanceinfo.Data)
	}
	// 通用查询，包含AddonName、Workspace、Project
	total, addons, err := a.db.ListAddonInstancesByParams(orgID, params)
	if err != nil {
		return nil, err
	}
	if len(*addons) == 0 {
		idQueryParams := apistructs.MiddlewareListRequest{
			InstanceID: params.AddonName,
		}
		ins, err := a.db.GetAddonInstance(params.AddonName)
		if err != nil {
			return nil, err
		}
		if ins == nil {
			return &apistructs.MiddlewareListResponseData{
				Total: 0,
				Overview: apistructs.Overview{
					CPU:   0.0,
					Mem:   0.0,
					Nodes: 0,
				},
				List: []apistructs.MiddlewareListItem{},
			}, nil
		}
		return a.MiddlewareListItem(1, orgID, &idQueryParams, &([]dbclient.AddonInstance{*ins}), limited_addonids)
	} else {
		return a.MiddlewareListItem(total, orgID, params, addons, limited_addonids)
	}
}

func isOperatorAddon(addon dbclient.AddonInstance) bool {
	if addon.AddonName == "terminus-elasticsearch" && addon.Version == "6.8.9" {
		return true
	}
	return false
}

// MiddlewareListItem item抽取代码，通用
func (a *Addon) MiddlewareListItem(total int, orgID uint64,
	params *apistructs.MiddlewareListRequest, addons *[]dbclient.AddonInstance,
	limited_addonids []string) (*apistructs.MiddlewareListResponseData, error) {
	middlewares := make([]apistructs.MiddlewareListItem, 0, len(*addons))
	for _, v := range *addons {
		if limited_addonids != nil && !strutil.Exist(limited_addonids, v.ID) {
			continue
		}
		// 从 node 表获取 cpu & mem
		nodes, err := a.db.GetAddonNodesByInstanceID(v.ID)
		if err != nil {
			return nil, err
		}
		var (
			cpu float64
			mem uint64
		)
		for _, node := range *nodes {
			cpu += node.CPU
			mem += node.Mem
		}
		cpuParse, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cpu)), 64) // 转换为G
		if err != nil {
			return nil, err
		}
		attachCount, err := a.db.GetAttachmentCountByInstanceID(v.ID)
		if err != nil {
			return nil, err
		}
		if v.ProjectID == "" {
			continue
		}
		projectInfos, err := a.getProject(v.ProjectID)
		if err != nil {
			continue
		}
		item := apistructs.MiddlewareListItem{
			InstanceID:  v.ID,
			AddonName:   v.AddonName,
			Env:         v.Workspace,
			Name:        v.AddonName + "-" + v.ID,
			CPU:         cpuParse,
			Mem:         mem,
			ClusterName: v.Cluster,
			ProjectID:   v.ProjectID,
			ProjectName: projectInfos.Name,
			Nodes:       len(*nodes),
			AttachCount: attachCount,
			IsOperator:  isOperatorAddon(v),
		}
		middlewares = append(middlewares, item)
	}

	var overview apistructs.Overview
	allInstances, err := a.db.ListAddonInstancesByParamsWithoutPage(orgID, params)
	if err != nil {
		return nil, err
	}
	for _, v := range *allInstances {
		// 从 node 表获取 cpu & mem
		nodes, err := a.db.GetAddonNodesByInstanceID(v.ID)
		if err != nil {
			return nil, err
		}
		for _, node := range *nodes {
			overview.CPU += node.CPU
			overview.Mem += float64(node.Mem)
			overview.Nodes += 1
		}
	}
	// 转换精度，只取2位小数
	overMem, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(overview.Mem)/float64(1024)), 64) // 转换为G
	overview.Mem = overMem
	// 转换精度，只取2位小数
	overCPU, err := strconv.ParseFloat(fmt.Sprintf("%.2f", overview.CPU), 64)
	overview.CPU = overCPU

	if limited_addonids != nil {
		total = len(middlewares)
	}
	return &apistructs.MiddlewareListResponseData{
		Total:    total,
		Overview: overview,
		List:     middlewares,
	}, nil
}

// GetMiddleware 获取 middleware 详情
func (a *Addon) GetMiddleware(orgID uint64, userID, instanceID string) (*apistructs.MiddlewareFetchResponseData, error) {

	instance, err := a.db.GetAddonInstance(instanceID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.Errorf("not found instance: %s", instanceID)
	}
	references, err := a.ListReferencesByInstanceID(orgID, userID, instanceID)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	if instance.Config != "" {
		if err = json.Unmarshal([]byte(instance.Config), &config); err != nil {
			return nil, err
		}
	}
	// config中的password需要过滤
	for k := range config {
		if strings.Contains(strings.ToLower(k), "pass") || strings.Contains(strings.ToLower(k), "secret") {
			config[k] = ""
		}
	}
	proj, err := a.getProject(instance.ProjectID)
	if err != nil {
		return nil, err
	}
	response := apistructs.MiddlewareFetchResponseData{
		Name:           instance.AddonName + "-" + instance.ID,
		IsOperator:     isOperatorAddon(*instance),
		InstanceID:     instance.ID,
		AddonName:      instance.AddonName,
		Plan:           instance.Plan,
		Version:        instance.Version,
		Category:       instance.Category,
		Cluster:        instance.Cluster,
		Workspace:      instance.Workspace,
		ProjectID:      instance.ProjectID,
		ProjectName:    proj.Name,
		Status:         instance.Status,
		AttachCount:    len(*references),
		Config:         config,
		ReferenceInfos: *references,
		CreatedAt:      instance.CreatedAt,
		UpdatedAt:      instance.UpdatedAt,
	}

	if v, ok := AddonInfos.Load(instance.AddonName); ok {
		response.LogoURL = v.(apistructs.Extension).LogoUrl
	} else {
		logrus.Warnf("failed to fetch addon info: %s", instance.AddonName)
	}

	return &response, nil
}

// GetMiddlewareAddonClassification 获取 middleware addon分类使用情况
func (a *Addon) GetMiddlewareAddonClassification(orgID uint64, params *apistructs.MiddlewareListRequest) (*map[string]apistructs.MiddlewareResourceItem, error) {
	result := map[string]apistructs.MiddlewareResourceItem{}
	addons, err := a.db.ListAddonInstancesByParamsWithoutPage(orgID, params)
	if err != nil {
		return nil, err
	}
	if len(*addons) == 0 {
		if params.AddonName == "" {
			return &result, nil
		}
		addons, err = a.db.ListAddonInstancesByParamsWithoutPage(orgID, &apistructs.MiddlewareListRequest{
			InstanceID: params.AddonName,
		})
		if err != nil {
			return nil, err
		}
	}
	if len(*addons) == 0 {
		return &result, nil
	}
	instanceIDs := []string{}
	addonInstanceMap := map[string]string{}
	for _, v := range *addons {
		if v.Category != "custom" {
			instanceIDs = append(instanceIDs, v.ID)
			addonInstanceMap[v.ID] = v.AddonName
		}
	}
	instanceNodes, err := a.db.GetAddonNodesByInstanceIDs(instanceIDs)
	for _, v := range *instanceNodes {
		if _, ok := addonInstanceMap[v.InstanceID]; !ok {
			continue
		}
		addonName := addonInstanceMap[v.InstanceID]
		if _, ok := result[addonName]; !ok {
			result[addonName] = apistructs.MiddlewareResourceItem{
				CPU: v.CPU,
				Mem: float64(v.Mem),
			}
		} else {
			oldResource := result[addonName]
			oldResource.CPU += v.CPU
			oldResource.Mem += float64(v.Mem)
			result[addonName] = oldResource
		}
	}

	// 转换mem的单位为G
	conversionMap := map[string]apistructs.MiddlewareResourceItem{}
	for k, v := range result {
		memCpu, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", v.CPU), 64)
		memValue, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", v.Mem/1024), 64)
		conversionMap[strings.Replace(k, "terminus-", "", 1)] = apistructs.MiddlewareResourceItem{
			CPU: memCpu,
			Mem: memValue,
		}
	}

	return &conversionMap, nil
}

// GetMiddlewareAddonDaily 获取 middleware 每日资源使用情况
func (a *Addon) GetMiddlewareAddonDaily(orgID uint64, params *apistructs.MiddlewareListRequest) (*map[string]interface{}, error) {

	timesArr := []string{}
	result := []apistructs.MiddlewareResourceItem{}
	// 获取今天的addon资源占用情况
	currentTime := time.Now()

	// 获取大大前天的addon资源占用情况
	LastFourTime := currentTime.AddDate(0, 0, -4)
	params.EndTime = &LastFourTime
	LastFourTimeResult, err := a.GetTotalResourceWithTime(orgID, params)
	if err != nil {
		return nil, err
	}
	timesArr = append(timesArr, LastFourTime.Format("2006-01-02"))
	result = append(result, *LastFourTimeResult)

	// 获取大前天的addon资源占用情况
	LastThreeTime := currentTime.AddDate(0, 0, -3)
	params.EndTime = &LastThreeTime
	LastThreeTimeResult, err := a.GetTotalResourceWithTime(orgID, params)
	if err != nil {
		return nil, err
	}
	timesArr = append(timesArr, LastThreeTime.Format("2006-01-02"))
	result = append(result, *LastThreeTimeResult)

	// 获取前天的addon资源占用情况
	LastTwoTime := currentTime.AddDate(0, 0, -2)
	params.EndTime = &LastTwoTime
	LastTwoTimeResult, err := a.GetTotalResourceWithTime(orgID, params)
	if err != nil {
		return nil, err
	}
	timesArr = append(timesArr, LastTwoTime.Format("2006-01-02"))
	result = append(result, *LastTwoTimeResult)

	// 获取昨天的addon资源占用情况
	lastOneTime := currentTime.AddDate(0, 0, -1)
	params.EndTime = &lastOneTime
	lastOneResult, err := a.GetTotalResourceWithTime(orgID, params)
	if err != nil {
		return nil, err
	}
	timesArr = append(timesArr, lastOneTime.Format("2006-01-02"))
	result = append(result, *lastOneResult)

	// 获取今天的addon资源占用情况
	params.EndTime = &currentTime
	currentResult, err := a.GetTotalResourceWithTime(orgID, params)
	if err != nil {
		return nil, err
	}
	timesArr = append(timesArr, currentTime.Format("2006-01-02"))
	result = append(result, *currentResult)

	return &map[string]interface{}{"abscissa": timesArr, "resource": result}, nil
}

// InnerGetMiddleware 内部获取middleware详情
func (a *Addon) InnerGetMiddleware(instanceID string) (*apistructs.MiddlewareFetchResponseData, error) {
	instance, err := a.db.GetAddonInstance(instanceID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.Errorf("not found instance: %s", instanceID)
	}
	response := apistructs.MiddlewareFetchResponseData{
		Name:       instance.AddonName + "-" + instance.ID,
		InstanceID: instance.ID,
		AddonName:  instance.AddonName,
		Plan:       instance.Plan,
		Version:    instance.Version,
		Category:   instance.Category,
		Cluster:    instance.Cluster,
		Workspace:  instance.Workspace,
		ProjectID:  instance.ProjectID,
		Status:     instance.Status,
		CreatedAt:  instance.CreatedAt,
		UpdatedAt:  instance.UpdatedAt,
	}

	return &response, nil
}

// GetMiddlewareResource 获取 middleware 资源详情
func (a *Addon) GetMiddlewareResource(instanceID string) ([]apistructs.MiddlewareResourceFetchResponseData, error) {
	resp, err := a.bdl.GetInstanceInfo(apistructs.InstanceInfoRequest{
		AddonID: instanceID,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.Errorf("not found instance: %s", instanceID)
	}

	middlewares := make([]apistructs.MiddlewareResourceFetchResponseData, 0, len(resp.Data))
	for _, v := range resp.Data {
		item := apistructs.MiddlewareResourceFetchResponseData{
			InstanceID:  instanceID,
			ContainerID: v.ContainerID,
			ContainerIP: v.ContainerIP,
			ClusterName: v.Cluster,
			HostIP:      v.HostIP,
			Image:       v.Image,
			CPURequest:  v.CpuRequest,
			CPULimit:    v.CpuLimit,
			MemRequest:  uint64(v.MemRequest),
			MemLimit:    uint64(v.MemLimit),
			Status:      v.Phase,
			StartedAt:   v.StartedAt,
		}
		middlewares = append(middlewares, item)
	}

	return middlewares, nil
}

// GetTotalResourceWithTime 根据截止时间来查询addon资源占用情况
func (a *Addon) GetTotalResourceWithTime(orgID uint64, params *apistructs.MiddlewareListRequest) (*apistructs.MiddlewareResourceItem, error) {
	allInstances, err := a.db.ListAddonInstancesByParamsWithoutPage(orgID, params)
	if err != nil {
		return nil, err
	}
	if len(*allInstances) == 0 {
		if params.AddonName == "" {
			return &apistructs.MiddlewareResourceItem{}, nil
		}
		allInstances, err = a.db.ListAddonInstancesByParamsWithoutPage(orgID, &apistructs.MiddlewareListRequest{
			InstanceID: params.AddonName,
			EndTime:    params.EndTime,
		})
		if err != nil {
			return nil, err
		}
		if len(*allInstances) == 0 {
			return &apistructs.MiddlewareResourceItem{}, nil
		}
	}
	item := apistructs.MiddlewareResourceItem{}
	for _, v := range *allInstances {
		// 从 node 表获取 cpu & mem
		nodes, err := a.db.GetAddonNodesByInstanceID(v.ID)
		if err != nil {
			return nil, err
		}
		for _, node := range *nodes {
			item.CPU += node.CPU
			item.Mem += float64(node.Mem)
		}
	}
	item.CPU, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.CPU), 64)
	item.Mem, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Mem/1024), 64)
	return &item, nil
}
