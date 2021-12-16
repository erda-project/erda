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

package apistructs

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ApplicationsResourcesRequest struct {
	OrgID     string
	UserID    string
	ProjectID string
	Query     *ApplicationsResourceQuery
}

func (arr ApplicationsResourcesRequest) Validate() error {
	if _, err := arr.GetOrgID(); err != nil {
		return errors.Wrap(err, "Org-ID is invalid")
	}
	if _, err := arr.GetUserID(); err != nil {
		return errors.Wrap(err, "User-ID is invalid")
	}
	if _, err := arr.GetProjectID(); err != nil {
		return errors.Wrap(err, "projectID is invalid")
	}

	return nil
}

func (arr ApplicationsResourcesRequest) GetOrgID() (uint64, error) {
	return strconv.ParseUint(arr.OrgID, 10, 64)
}

func (arr ApplicationsResourcesRequest) GetUserID() (uint64, error) {
	return strconv.ParseUint(arr.OrgID, 10, 64)
}

func (arr ApplicationsResourcesRequest) GetProjectID() (uint64, error) {
	return strconv.ParseUint(arr.ProjectID, 10, 64)
}

type ApplicationsResourceQuery struct {
	AppsIDs  []string
	OwnerIDs []string
	OrderBy  []string
	PageNo   string
	PageSize string
}

func (arq ApplicationsResourceQuery) GetAppIDs() []uint64 {
	return arq.uin64Slice(arq.AppsIDs)
}

func (arq ApplicationsResourceQuery) GetOwnerIDs() []uint64 {
	return arq.uin64Slice(arq.OwnerIDs)
}

func (arq ApplicationsResourceQuery) uin64Slice(ss []string) []uint64 {
	var result []uint64
	for _, s := range ss {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}

func (arq ApplicationsResourceQuery) GetPageNo() int64 {
	i, err := strconv.ParseInt(arq.PageNo, 10, 64)
	if err != nil || i < 1 {
		return 1
	}
	return i
}

func (arq ApplicationsResourceQuery) GetPageSize() int64 {
	i, err := strconv.ParseInt(arq.PageNo, 10, 64)
	if err != nil || i < 1 {
		return 15
	}
	return i
}

type ApplicationsResourcesResponse struct {
	Total int                          `json:"total"`
	List  []*ApplicationsResourcesItem `json:"list"`
}

func (r *ApplicationsResourcesResponse) OrderBy(conditions ...string) {
	for i := len(conditions) - 1; i >= 0; i-- {
		switch condition := conditions[i]; {
		case strings.EqualFold(condition, "-podsCount"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].PodsCount > r.List[j].PodsCount
			})
		case strings.EqualFold(condition, "podsCount"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].PodsCount < r.List[j].PodsCount
			})
		case strings.EqualFold(condition, "-cpuRequest"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].CPURequest > r.List[j].CPURequest
			})
		case strings.EqualFold(condition, "cpuRequest"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].CPURequest < r.List[j].CPURequest
			})
		case strings.EqualFold(condition, "-memRequest"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].MemRequest > r.List[j].MemRequest
			})
		case strings.EqualFold(condition, "memRequest"):
			sort.Slice(r.List, func(i, j int) bool {
				return r.List[i].MemRequest < r.List[j].MemRequest
			})
		}
	}
}

func (r *ApplicationsResourcesResponse) Paging(pageSize, pageNo int64) {
	if pageNo < 1 {
		pageNo = 1
	}
	i := (pageNo - 1) * pageSize
	j := i + pageSize
	if int(i) >= len(r.List) {
		r.List = nil
		return
	}
	if int(j) >= len(r.List) {
		r.List = r.List[i:]
		return
	}
	r.List = r.List[i:j]
}

type ApplicationsResourcesItem struct {
	ID                uint64 `json:"id"` // the application primary
	Name              string `json:"name"`
	DisplayName       string `json:"displayName"`
	OwnerUserID       uint64 `json:"ownerUserID"`
	OwnerUserName     string `json:"ownerUserName"`
	OwnerUserNickname string `json:"ownerUserNickname"`
	PodsCount         uint64 `json:"podsCount"`
	CPURequest        uint64 `json:"cpuRequest"`
	MemRequest        uint64 `json:"memRequest"`
	ProdCPURequest    uint64 `json:"prodCPURequest"`
	ProdMemRequest    uint64 `json:"prodMemRequest"`
	ProdPodsCount     uint64 `json:"prodPodsCount"`
	StagingCPURequest uint64 `json:"stagingCPURequest"`
	StagingMemRequest uint64 `json:"stagingMemRequest"`
	StagingPodsCount  uint64 `json:"stagingPodsCount"`
	TestCPURequest    uint64 `json:"testCPURequest"`
	TestMemRequest    uint64 `json:"testMemRequest"`
	TestPodsCount     uint64 `json:"testPodsCount"`
	DevCPURequest     uint64 `json:"devCPURequest"`
	DevMemRequest     uint64 `json:"devMemRequest"`
	DevPodsCount      uint64 `json:"devPodsCount"`
}

func (i *ApplicationsResourcesItem) AddResource(workspace string, pods, cpu, mem uint64) {
	switch strings.ToUpper(workspace) {
	case "PROD":
		i.ProdPodsCount += pods
		i.ProdCPURequest += cpu
		i.ProdMemRequest += mem
	case "STAGING":
		i.StagingPodsCount += pods
		i.StagingCPURequest += cpu
		i.StagingMemRequest += mem
	case "TEST":
		i.TestPodsCount += pods
		i.TestCPURequest += cpu
		i.TestMemRequest += mem
	case "DEV":
		i.DevPodsCount += pods
		i.DevCPURequest += cpu
		i.DevMemRequest += mem
	}
	i.PodsCount = i.ProdPodsCount + i.StagingPodsCount + i.TestPodsCount + i.DevPodsCount
	i.CPURequest = i.ProdCPURequest + i.StagingCPURequest + i.TestCPURequest + i.DevCPURequest
	i.MemRequest = i.ProdMemRequest + i.StagingMemRequest + i.TestMemRequest + i.DevMemRequest
}

type GetResourceApplicationTrendReq struct {
	OrgID     string
	UserID    string
	ProjectID string

	Query *GetResourceApplicationTrendReqQuery
}

func (req GetResourceApplicationTrendReq) Validate() error {
	if _, err := req.GetOrgID(); err != nil {
		return errors.Wrap(err, "invalid OrgID")
	}
	if _, err := req.GetUserID(); err != nil {
		return errors.Wrap(err, "invalid UserID")
	}
	if _, err := req.GetProjectID(); err != nil {
		return errors.Wrap(err, "invalid ProjectID")
	}
	if req.Query == nil {
		return errors.New("the Query is empty")
	}
	return req.Query.Validate()
}

func (req GetResourceApplicationTrendReq) GetOrgID() (uint64, error) {
	return strconv.ParseUint(req.OrgID, 10, 32)
}

func (req GetResourceApplicationTrendReq) GetUserID() (uint64, error) {
	return strconv.ParseUint(req.UserID, 10, 32)
}

func (req GetResourceApplicationTrendReq) GetProjectID() (uint64, error) {
	return strconv.ParseUint(req.ProjectID, 10, 32)
}

type GetResourceApplicationTrendReqQuery struct {
	ApplicationID string // 要统计的应用 id
	Start         string // 统计值起始时间, 13 位时间戳
	End           string // 统计值结束时间, 13 位时间戳
	Interval      string // 统计聚合维度: 周期, 枚举: day, week, month
	ResourceType  string // 资源类型, 枚举: cpu, mem
}

func (rq GetResourceApplicationTrendReqQuery) Validate() error {
	if _, err := rq.GetApplicationID(); err != nil {
		return errors.Wrap(err, "invalid ApplicationID")
	}
	if _, err := rq.GetStart(); err != nil {
		return errors.Wrap(err, "invalid Start")
	}
	if _, err := rq.GetEnd(); err != nil {
		return errors.Wrap(err, "invalid End")
	}
	return nil
}

func (rq GetResourceApplicationTrendReqQuery) GetApplicationID() (uint64, error) {
	return strconv.ParseUint(rq.ApplicationID, 10, 32)
}

func (rq GetResourceApplicationTrendReqQuery) GetStart() (uint64, error) {
	return rq.getTime(rq.Start)
}

func (rq GetResourceApplicationTrendReqQuery) GetEnd() (uint64, error) {
	return rq.getTime(rq.End)
}

func (rq GetResourceApplicationTrendReqQuery) GetInterval() string {
	switch i := strings.ToLower(rq.Interval); i {
	case "day", "week", "month":
		return i
	default:
		return "day"
	}
}

func (rq GetResourceApplicationTrendReqQuery) getTime(t string) (uint64, error) {
	v, err := strconv.ParseUint(t, 10, 64)
	if err != nil {
		return 0, err
	}
	if v < 1_000_000_000_000 || v > 9_999_999_999_999 {
		return 0, errors.New("timestamp is invalid")
	}
	return v, nil
}
