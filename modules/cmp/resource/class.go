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

package resource

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
)

const (
	CPU       = "cpu"
	Memory    = "memory"
	Principal = "principal"
	Project   = "Project"
	Cluster   = "Cluster"
	Day       = "Day"
	Week      = "week"
	Month     = "Month"
)

var (
	errResourceTypeNotFound = errors.New("resource type not support")
	errIntervalTypeNotFound = errors.New("date type not support")
)

type PieData struct {
	Series []PieSerie `json:"series"`
}

type PieSerie struct {
	Data []SerieData `json:"data"`
	Name string      `json:"name"`
	Type string      `json:"type"`
}

type HistogramSerie struct {
	Data []float64 `json:"data"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

type SerieData struct {
	Value float64
	Name  string
}

type DailyProjectQuota struct {
	Index      int64 `gorm:"column:idx"`
	CpuQuota   int64 `gorm:"column:cpu_quota"`
	CpuRequest int64 `gorm:"column:cpu_request"`
	MemQuota   int64 `gorm:"column:mem_quota"`
	MemRequest int64 `gorm:"column:mem_request"`
}

type DailyClusterQuota struct {
	Index      int64 `gorm:"column:idx"`
	CpuTotal   int64 `gorm:"column:cpu_total"`
	CpuRequest int64 `gorm:"column:cpu_requested"`
	MemTotal   int64 `gorm:"column:mem_total"`
	MemRequest int64 `gorm:"column:mem_requested"`
}

type Quota struct {
	cpuQuota float64
	memQuota float64
}

func (r *Resource) GetPie(ordId int64, userId string, request *apistructs.ClassRequest) (data map[string]*PieData, err error) {
	var clusters []apistructs.ClusterInfo
	data = make(map[string]*PieData)
	clusters, err = r.Bdl.ListClusters("", uint64(ordId))
	if err != nil {
		return
	}
	request.ClusterName = r.FilterCluster(clusters, request.ClusterName)
	resp, err := r.Bdl.FetchQuotaOnClusters(uint64(ordId), request.ClusterName)
	if err != nil {
		return
	}

	greq := &pb.GetClustersResourcesRequest{}
	greq.ClusterNames = request.ClusterName
	resources, err := r.Server.GetClustersResources(r.Ctx, greq)
	if err != nil {
		return
	}

	// Project
	pie, err := r.GetProjectPie(request.ResourceType, resp)
	if err != nil {
		return
	}
	data[Project] = pie

	// principal
	pie, err = r.GetPrincipalPie(request.ResourceType, resp)
	if err != nil {
		return
	}
	data[Principal] = pie

	// Cluster
	pie, err = r.GetClusterPie(request.ResourceType, resources)
	if err != nil {
		return
	}
	data[Cluster] = pie
	return
}

func (r *Resource) GetProjectPie(resType string, resp *apistructs.GetQuotaOnClustersResponse) (projectPie *PieData, err error) {
	var (
		q  Quota
		ok bool
	)
	projectPie = &PieData{}
	pieSerie := PieSerie{
		Name: r.I18n("distribution by Project"),
		Type: "pie",
	}
	projectMap := make(map[string]Quota)
	for _, owner := range resp.Owners {
		for _, p := range owner.Projects {
			if q, ok = projectMap[p.Name]; !ok {
				q = Quota{}
			}
			q.cpuQuota += p.CPUQuota
			q.memQuota += p.MemQuota
			projectMap[p.Name] = q
		}
	}
	switch resType {

	case Memory:
		for k, v := range projectMap {
			pieSerie.Data = append(pieSerie.Data, SerieData{v.memQuota / G, k})
		}
	default:
		for k, v := range projectMap {
			pieSerie.Data = append(pieSerie.Data, SerieData{v.cpuQuota / MilliCore, k})
		}
	}
	projectPie.Series = append(projectPie.Series, pieSerie)
	return
}

func (r *Resource) GetPrincipalPie(resType string, resp *apistructs.GetQuotaOnClustersResponse) (principalPie *PieData, err error) {
	var (
		q  Quota
		ok bool
	)
	principalMap := make(map[string]Quota)
	for _, owner := range resp.Owners {
		if q, ok = principalMap[owner.Name]; !ok {
			q = Quota{}
		}
		q.cpuQuota = owner.CPUQuota
		q.memQuota = owner.MemQuota
		principalMap[owner.Name] = q
	}
	principalPie = &PieData{}
	serie := PieSerie{
		Name: r.I18n("distribution by principal"),
		Type: "pie",
	}
	switch resType {
	case Memory:
		for k, v := range principalMap {
			serie.Data = append(serie.Data, SerieData{v.memQuota / G, k})
		}
	default:
		for k, v := range principalMap {
			serie.Data = append(serie.Data, SerieData{v.cpuQuota / MilliCore, k})
		}
	}
	principalPie.Series = append(principalPie.Series, serie)
	return
}

func (r *Resource) GetClusterPie(resourceType string, resources *pb.GetClusterResourcesResponse) (clusterPie *PieData, err error) {
	clusterPie = &PieData{}
	serie := PieSerie{
		Name: r.I18n("distribution by Cluster"),
		Type: "pie",
	}
	switch resourceType {
	case Memory:
		for _, c := range resources.List {
			memSum := 0.0
			for _, h := range c.Hosts {
				memSum += float64(h.CpuTotal)
			}
			serie.Data = append(serie.Data, SerieData{memSum / G, c.ClusterName})
		}
	default:
		for _, c := range resources.List {
			cpuSum := 0.0
			for _, h := range c.Hosts {
				cpuSum += float64(h.CpuTotal)
			}
			serie.Data = append(serie.Data, SerieData{cpuSum / MilliCore, c.ClusterName})
		}
	}
	clusterPie.Series = append(clusterPie.Series, serie)
	return
}

func (r *Resource) GetClusterTrend(ordId int64, userId string, request *apistructs.TrendRequest) (td *Histogram, err error) {

	td = &Histogram{}
	td.XAixs = XAixs{
		Type: "category",
	}
	td.YAixs = YAixs{Type: "value"}
	td.Series = make([]HistogramSerie, 2)
	td.Series[0].Type = "bar"
	td.Series[1].Type = "bar"
	td.Series[0].Name = r.I18n("quota")
	td.Series[1].Name = r.I18n("total")
	var (
		pd       []apistructs.ClusterResourceDailyModel
		clusters []apistructs.ClusterInfo
	)
	clusters, err = r.Bdl.ListClusters("", uint64(ordId))
	if err != nil {
		return nil, err
	}
	request.ClusterName = r.FilterCluster(clusters, request.ClusterName)

	db := r.DB.Table("cmp_cluster_resource_daily")
	startTime := time.Unix(request.Start/1e3, request.Start%1e3*1e6)
	endTime := time.Unix(request.End/1e3, request.End%1e3*1e6)
	db.Raw("select SUM(cpu_total),SUM(cpu_requested),SUM(mem_total),SUM(mem_requested) where  updated_at < ? and updated_at >= ? and cluster_name in (?) sort by created_at desc", endTime, startTime, request.ClusterName)
	if err = db.Scan(&pd).Error; err != nil {
		return
	}
	tRes := make(map[int]apistructs.ClusterResourceDailyModel)
	switch request.Interval {
	case Week:
		for _, model := range pd {
			_, wk := model.CreatedAt.ISOWeek()
			if v, ok := tRes[wk]; ok {
				v.CPUTotal += model.CPUTotal
				v.MemTotal += model.MemTotal
				v.CPURequested += model.CPURequested
				v.MemRequested += model.MemRequested
				v.ID = uint64(wk)
			} else {
				tRes[wk] = model
			}
		}
	case Month:
		for _, model := range pd {
			if v, ok := tRes[int(model.CreatedAt.Month())]; ok {
				v.CPUTotal += model.CPUTotal
				v.MemTotal += model.MemTotal
				v.CPURequested += model.CPURequested
				v.MemRequested += model.MemRequested
				v.ID = uint64(model.CreatedAt.Month())
			} else {
				tRes[int(model.CreatedAt.Month())] = model
			}
		}
	default:
		// Day
		for _, model := range pd {
			if v, ok := tRes[model.CreatedAt.Day()]; ok {
				v.CPUTotal += model.CPUTotal
				v.MemTotal += model.MemTotal
				v.CPURequested += model.CPURequested
				v.MemRequested += model.MemRequested
				v.ID = uint64(model.CreatedAt.Day())
			} else {
				tRes[model.CreatedAt.Day()] = model
			}
		}
	}
	pd = make([]apistructs.ClusterResourceDailyModel, 0)
	for _, model := range tRes {
		pd = append(pd, model)
	}
	sort.Slice(pd, func(i, j int) bool {
		return pd[i].ID < pd[i].ID
	})

	switch request.ResourceType {

	case Memory:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toGB(float64(quota.MemRequested)))
			td.Series[1].Data = append(td.Series[1].Data, toGB(float64(quota.MemTotal)))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%s", quota.CreatedAt.String()))
		}
	default:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toCore(float64(quota.CPURequested)))
			td.Series[1].Data = append(td.Series[1].Data, toCore(float64(quota.CPUTotal)))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%s", quota.CreatedAt.String()))
		}
	}
	return
}

func (r *Resource) GetProjectTrend(ordId int64, userId string, request *apistructs.TrendRequest) (td *Histogram, err error) {
	td = &Histogram{}
	td.XAixs = XAixs{
		Type: "category",
	}
	td.YAixs = YAixs{Type: "value"}
	td.Series = make([]HistogramSerie, 2)
	td.Series[0].Type = "bar"
	td.Series[1].Type = "bar"
	td.Series[0].Name = r.I18n("request")
	td.Series[1].Name = r.I18n("quota")
	var (
		pd       []apistructs.ProjectResourceDailyModel
		clusters []apistructs.ClusterInfo
	)
	clusters, err = r.Bdl.ListClusters("", uint64(ordId))
	if err != nil {
		return nil, err
	}
	request.ClusterName = r.FilterCluster(clusters, request.ClusterName)

	db := r.DB.Table("cmp_project_resource_daily")
	db.Raw("select  SUM(cpu_quota),SUM(cpu_request),SUM(mem_quota),SUM(mem_request)  updated_at < ? and updated_at >= ? and cluster_name in (?)  and project_id in (?) sort by created_at desc", request.End, request.Start, request.ClusterName, request.ProjectId)
	if err = db.Scan(&pd).Error; err != nil {
		return
	}

	tRes := make(map[int]apistructs.ProjectResourceDailyModel)
	switch request.Interval {
	case Week:
		for _, model := range pd {
			_, wk := model.CreatedAt.ISOWeek()
			if v, ok := tRes[wk]; ok {
				v.CPURequest += model.CPURequest
				v.CPUQuota += model.CPUQuota
				v.MemRequest += model.MemRequest
				v.MemQuota += model.MemQuota
				v.ID = uint64(wk)
			} else {
				tRes[wk] = model
			}
		}
	case Month:
		for _, model := range pd {
			if v, ok := tRes[int(model.CreatedAt.Month())]; ok {
				v.CPURequest += model.CPURequest
				v.CPUQuota += model.CPUQuota
				v.MemRequest += model.MemRequest
				v.MemQuota += model.MemQuota
				v.ID = uint64(model.CreatedAt.Month())
			} else {
				tRes[int(model.CreatedAt.Month())] = model
			}
		}
	default:
		// Day
		for _, model := range pd {
			if v, ok := tRes[model.CreatedAt.Day()]; ok {
				v.CPURequest += model.CPURequest
				v.CPUQuota += model.CPUQuota
				v.MemRequest += model.MemRequest
				v.MemQuota += model.MemQuota
				v.ID = uint64(model.CreatedAt.Day())
			} else {
				tRes[model.CreatedAt.Day()] = model
			}
		}
	}
	pd = make([]apistructs.ProjectResourceDailyModel, 0)
	for _, model := range tRes {
		pd = append(pd, model)
	}
	sort.Slice(pd, func(i, j int) bool {
		return pd[i].ID < pd[i].ID
	})
	switch request.ResourceType {

	case Memory:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toGB(float64(quota.MemRequest)))
			td.Series[1].Data = append(td.Series[1].Data, toGB(float64(quota.MemQuota)))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%s", quota.CreatedAt.String()))
		}
	default:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toCore(float64(quota.CPURequest)))
			td.Series[1].Data = append(td.Series[1].Data, toCore(float64(quota.CPUQuota)))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%s", quota.CreatedAt.String()))
		}
	}
	return
}

func toCore(mCores float64) float64 {
	return mCores / 1000
}

func toGB(b float64) float64 {
	return b / float64(1<<30)
}
