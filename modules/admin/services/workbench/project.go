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

package workbench

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func GetProjectNum(bdl *bundle.Bundle, identity apistructs.Identity, query string) (int, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		Query:    query,
		PageNo:   1,
		PageSize: 1,
	}
	projectDTO, err := bdl.ListMyProject(identity.UserID, req)
	if err != nil {
		return 0, err
	}
	if projectDTO == nil {
		return 0, nil
	}
	return projectDTO.Total, nil
}

func ListProjectWorkbenchData(bdl *bundle.Bundle, identity apistructs.Identity, page apistructs.PageRequest, projectIDs []uint64) (*apistructs.WorkbenchResponseData, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}
	req := apistructs.WorkbenchRequest{
		OrgID:      uint64(orgID),
		PageSize:   page.PageSize,
		PageNo:     page.PageNo,
		ProjectIDs: projectIDs,
	}
	res, err := bdl.GetWorkbenchData(identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func ListSubProjWorkbenchData(bdl *bundle.Bundle, identity apistructs.Identity) (*apistructs.WorkbenchResponseData, error) {
	subList, err := bdl.ListSubscribes(identity.UserID, identity.OrgID, apistructs.GetSubscribeReq{Type: apistructs.ProjectSubscribe})
	if err != nil {
		return nil, err
	}
	pIDList := make([]uint64, len(subList.List))
	for _, v := range subList.List {
		pIDList = append(pIDList, v.TypeID)
	}
	page := apistructs.PageRequest{
		PageNo:   1,
		PageSize: len(pIDList),
	}
	return ListProjectWorkbenchData(bdl, identity, page, pIDList)
}

func ListQueryProjWorkbenchData(bdl *bundle.Bundle, identity apistructs.Identity, page apistructs.PageRequest, query string) (*apistructs.WorkbenchResponseData, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Query:    query,
	}
	projectDTO, err := bdl.ListMyProject(identity.UserID, req)
	if err != nil {
		return nil, err
	}
	if projectDTO == nil {
		return nil, nil
	}
	pIDList := make([]uint64, len(projectDTO.List))
	for _, v := range projectDTO.List {
		pIDList = append(pIDList, v.ID)
	}
	return ListProjectWorkbenchData(bdl, identity, page, pIDList)
}
