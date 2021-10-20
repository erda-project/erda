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

	"github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
)

const (
	CPU       = "cpu"
	Memory    = "memory"
	principal = "principal"
	project   = "project"
	cluster   = "cluster"
	day       = "day"
	week      = "week"
	month     = "month"
)

var (
	errResourceTypeNotFound = errors.New("resource type not support")
	errIntervalTypeNotFound = errors.New("date type not support")
)

type PieData struct {
	series []PieSerie
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
	Index      int     `gorm:"column:idx"`
	CpuQuota   float64 `gorm:"column:cpu_quota"`
	CpuRequest float64 `gorm:"column:cpu_request"`
	MemQuota   float64 `gorm:"column:mem_quota"`
	MemRequest float64 `gorm:"column:mem_request"`
}

type DailyClusterQuota struct {
	Index      int     `gorm:"column:idx"`
	CpuTotal   float64 `gorm:"column:cpu_total"`
	CpuRequest float64 `gorm:"column:cpu_requested"`
	MemTotal   float64 `gorm:"column:mem_total"`
	MemRequest float64 `gorm:"column:mem_requested"`
}

type Quota struct {
	cpuQuota float64
	memQuota float64
}

func (r *Resource) GetPie(ordId string, userId string, request *apistructs.ClassRequest) (data map[string]*PieData, err error) {
	data = make(map[string]*PieData)

	req := &apistructs.GetQuotaOnClustersRequest{}
	req.ClusterNames = request.ClusterNames
	req.OrgID = ordId
	resp, err := bdl.Bdl.FetchQuota(req)
	if err != nil {
		return
	}

	greq := &pb.GetClustersResourcesRequest{}
	greq.ClusterNames = request.ClusterNames
	resources, err := r.Server.GetClustersResources(r.Ctx, greq)
	if err != nil {
		return
	}

	// project
	pie, err := r.GetProjectPie(request.ResourceType, resp)
	if err != nil {
		return
	}
	data[project] = pie

	// principal
	pie, err = r.GetProjectPie(request.ResourceType, resp)
	if err != nil {
		return
	}
	data[principal] = pie

	// cluster
	pie, err = r.GetClusterPie(request.ResourceType, resources)
	if err != nil {
		return
	}
	data[cluster] = pie
	return
}

func (r *Resource) GetProjectPie(resourceType string, resp *apistructs.GetQuotaOnClustersResponse) (projectPie *PieData, err error) {
	var (
		q  Quota
		ok bool
	)
	projectPie = &PieData{}
	serie := PieSerie{
		Name: r.I18n("distribution by project"),
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
	switch resourceType {
	case CPU:
		for k, v := range projectMap {
			serie.Data = append(serie.Data, SerieData{v.cpuQuota, k})
		}
	case Memory:
		for k, v := range projectMap {
			serie.Data = append(serie.Data, SerieData{v.memQuota, k})
		}
	default:
		err = errResourceTypeNotFound
		return
	}
	projectPie.series = append(projectPie.series, serie)
	return
}

func (r *Resource) GetPrincipalPie(resourceType string, resp *apistructs.GetQuotaOnClustersResponse) (projectPie *PieData, err error) {
	var (
		q  Quota
		ok bool
	)
	projectPie = &PieData{}
	principalMap := make(map[string]Quota)
	for _, owner := range resp.Owners {
		if q, ok = principalMap[owner.Name]; !ok {
			q = Quota{}
		}
		q.cpuQuota = owner.CPUQuota
		q.memQuota = owner.MemQuota
		principalMap[owner.Name] = q
	}
	principalPie := &PieData{}
	serie := PieSerie{
		Name: r.I18n("distribution by principal"),
		Type: "pie",
	}
	switch resourceType {
	case CPU:
		for k, v := range principalMap {
			serie.Data = append(serie.Data, SerieData{v.cpuQuota, k})
		}
	case Memory:
		for k, v := range principalMap {
			serie.Data = append(serie.Data, SerieData{v.memQuota, k})
		}
	default:
		err = errResourceTypeNotFound
		return
	}
	principalPie.series = append(principalPie.series, serie)
	return
}

func (r *Resource) GetClusterPie(resourceType string, resources *pb.GetClusterResourcesResponse) (projectPie *PieData, err error) {
	clusterPie := &PieData{}
	serie := PieSerie{
		Name: r.I18n("distribution by cluster"),
		Type: "pie",
	}

	switch resourceType {
	case CPU:
		for _, c := range resources.List {
			cpuSum := 0.0
			for _, h := range c.Hosts {
				cpuSum += float64(h.CpuTotal)
			}
			serie.Data = append(serie.Data, SerieData{cpuSum, c.ClusterName})
		}
	case Memory:
		for _, c := range resources.List {
			memSum := 0.0
			for _, h := range c.Hosts {
				memSum += float64(h.CpuTotal)
			}
			serie.Data = append(serie.Data, SerieData{memSum, c.ClusterName})
		}
	default:
		err = errResourceTypeNotFound
		return
	}
	clusterPie.series = append(clusterPie.series, serie)
	return
}

func (r *Resource) GetClusterTrend(ordId string, userId string, request *apistructs.TrendRequest) (td *Histogram, err error) {
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
		pd []DailyClusterQuota
	)

	db := r.DB.Table("cmp_cluster_resource_daily")
	switch request.Interval {
	case day:
		db.Raw("select date as idx, SUM(cpu_total),SUM(cpu_requested),SUM(mem_total),SUM(mem_requested) where  updated_at < ? and updated_at >= ? and cluster_name in (?)", request.End, request.Start, request.ClusterNames)
		db.Group("date")
	case week:
		db.Raw("select WEEK(MY_DATE, 5)+1 as idx, SUM(cpu_total),SUM(cpu_requested),SUM(mem_total),SUM(mem_requested) where  updated_at < ? and updated_at >= ? and cluster_name in (?)", request.End, request.Start, request.ClusterNames)
		db.Group("WEEK(date, 5)")

	case month:
		db.Raw("select MONTH(date) as idx, SUM(cpu_total),SUM(cpu_requested),SUM(mem_total),SUM(mem_requested) where updated_at < ? and updated_at >= ? and cluster_name in (?)", request.End, request.Start, request.ClusterNames)
		db.Group("MONTH(date)")
	default:
		err = errIntervalTypeNotFound
		return
	}
	if err = db.Group("cluster_name").Scan(&pd).Error; err != nil {
		return
	}
	switch request.ResourceType {
	case CPU:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toCore(quota.CpuRequest))
			td.Series[1].Data = append(td.Series[1].Data, toCore(quota.CpuTotal))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%d", quota.Index))
		}
	case Memory:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toGB(quota.MemRequest))
			td.Series[1].Data = append(td.Series[1].Data, toGB(quota.MemTotal))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%d", quota.Index))
		}
	default:
		err = errResourceTypeNotFound
		return
	}
	return
}

func (r *Resource) GetProjectTrend(ordId string, userId string, request *apistructs.TrendRequest) (td *Histogram, err error) {
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
		pd []DailyProjectQuota
	)

	db := r.DB.Table("cmp_project_resource_daily")
	switch request.Interval {
	case day:
		db.Raw("select date as idx, SUM(cpu_quota),SUM(cpu_request),SUM(mem_quota),SUM(mem_request)  updated_at < ? and updated_at >= ? and cluster_name in (?)  and project_id in (?)", request.End, request.Start, request.ClusterNames, request.ProjectIds)
		db.Group("date")
	case week:
		db.Raw("select WEEK(MY_DATE, 5)+1 as idx, SUM(cpu_quota),SUM(cpu_request),SUM(mem_quota),SUM(mem_request)  updated_at < ? and updated_at >= ? and cluster_name in (?)  and project_id in (?)", request.End, request.Start, request.ClusterNames, request.ProjectIds)
		db.Group("WEEK(date, 5)")
	case month:
		db.Raw("select MONTH(date) as idx, SUM(cpu_quota),SUM(cpu_request),SUM(mem_quota),SUM(mem_request)  updated_at < ? and updated_at >= ? and cluster_name in (?) and project_id in (?)", request.End, request.Start, request.ClusterNames, request.ProjectIds)
		db.Group("MONTH(date)")
	default:
		err = errIntervalTypeNotFound
		return
	}

	if err = db.Group("project_id").Scan(&pd).Error; err != nil {
		return
	}
	switch request.ResourceType {
	case CPU:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toCore(quota.CpuRequest))
			td.Series[1].Data = append(td.Series[1].Data, toCore(quota.CpuQuota))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%d", quota.Index))
		}
	case Memory:
		for _, quota := range pd {
			td.Series[0].Data = append(td.Series[0].Data, toGB(quota.MemRequest))
			td.Series[1].Data = append(td.Series[1].Data, toGB(quota.MemQuota))
			td.XAixs.Data = append(td.XAixs.Data, fmt.Sprintf("%d", quota.Index))
		}
	default:
		err = errResourceTypeNotFound
		return
	}
	return
}

func toCore(mCores float64) float64 {
	return mCores / 1000
}

func toGB(b float64) float64 {
	return b / float64(1<<30)
}
