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

package project

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	dashboardPb "github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/cache"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (p *Project) ApplicationsResources(ctx context.Context, req *apistructs.ApplicationsResourcesRequest) (
	*apistructs.ApplicationsResourcesResponse, *errorresp.APIError) {
	l := logrus.WithField("func", "*Application.ApplicationsResources")
	if req == nil {
		err := errors.New("request can not be nil")
		l.WithError(err).Errorln("the req is nil")
		return nil, apierrors.ErrApplicationsResources.InvalidParameter(err)
	}
	var (
		applicationFilter = make(map[uint64]struct{})
		ownerFilter       = map[uint64]struct{}{0: {}}
	)
	for _, applicationID := range req.Query.ApplicationsIDs {
		applicationFilter[applicationID] = struct{}{}
	}
	for _, ownerID := range req.Query.OwnerIDs {
		ownerFilter[ownerID] = struct{}{}
	}

	var data apistructs.ApplicationsResourcesResponse

	// query project info from core-services
	project, err := p.bdl.GetProject(req.ProjectID)
	if err != nil {
		l.WithError(err).Errorf("failed to GetProject(%v)", req.ProjectID)
		return nil, apierrors.ErrApplicationsResources.InternalError(err)
	}
	if len(project.ClusterConfig) == 0 {
		l.WithField("projectID", req.ProjectID).Warnln("clusters not found")
		return &data, nil
	}

	// query applications list from core-services
	// todo: pageSize
	apps, err := p.bdl.GetAppsByProject(req.ProjectID, req.OrgID, req.UserID)
	if err != nil {
		l.WithError(err).Errorln("failed to GetAppsByProject")
		return nil, apierrors.ErrApplicationsResources.InternalError(err)
	}
	if apps.Total == 0 || len(apps.List) == 0 {
		return &data, nil
	}
	var (
		applicationsIDs []uint64
		items           = make(map[uint64]*apistructs.ApplicationsResourcesItem)
	)
	for _, application := range apps.List {
		if _, ok := applicationFilter[application.ID]; !ok {
			continue
		}
		var item = new(apistructs.ApplicationsResourcesItem)
		item.ID = application.ID
		item.Name = application.Name
		item.DisplayName = application.DisplayName
		if cacheItem, _ := p.appOwnerCache.LoadWithUpdate(application.ID); cacheItem != nil {
			owners := cacheItem.Object.(*memberCacheObject)
			owner, ok := owners.hasMemberIn(ownerFilter)
			if !ok {
				continue
			}
			item.OwnerUserID = owner.ID
			item.OwnerUserName = owner.Name
			item.OwnerUserNickname = owner.Nick
		}
		applicationsIDs = append(applicationsIDs, application.ID)
		items[application.ID] = item
		data.List = append(data.List, item)
	}
	if len(applicationsIDs) == 0 {
		return &data, nil
	}

	// query runtimes list from orchestrator
	runtimesM, err := p.bdl.GetApplicationsRuntimes(req.OrgID, req.UserID, applicationsIDs)
	if err != nil {
		l.WithError(err).Errorln("failed to GetApplicationsRuntimes")
		return nil, apierrors.ErrApplicationsResources.InternalError(err)
	}

	// serviceGroupID group by applicationID and workspace
	var (
		// {applicationID: {workspace: []serviceGroupID}}
		applicationWorkspaceServiceGroups = make(map[uint64]map[string][]string)
	)
	for applicationID, runtimes := range runtimesM {
		workspaceServiceGroups, ok := applicationWorkspaceServiceGroups[applicationID]
		if !ok {
			workspaceServiceGroups = make(map[string][]string)
		}
		for _, runtime := range runtimes {
			if runtime.Extra != nil {
				workspace := strings.ToUpper(runtime.Extra.Workspace)
				workspaceServiceGroups[workspace] = append(workspaceServiceGroups[workspace], runtime.ServiceGroupName)
			}
		}
		applicationWorkspaceServiceGroups[applicationID] = workspaceServiceGroups
	}

	// query pods by the label "servicegroup-id" for every application and workspace
	for applicationsID, workspaceServiceGroup := range applicationWorkspaceServiceGroups {
		item := items[applicationsID]

		for workspace, serviceGroupIDs := range workspaceServiceGroup {
			getPodsRequest := dashboardPb.GetPodsByLabelsRequest{
				Cluster: project.ClusterConfig[workspace],
				Labels:  []string{fmt.Sprintf("servicegroup-id in (%s)", strings.Join(serviceGroupIDs, ","))},
			}
			pods, err := p.cmp.GetPodsByLabels(ctx, &getPodsRequest)
			if err != nil {
				l.WithError(err).Errorf("failed to GetPodsByLabels, request: %+v", getPodsRequest)
			}
			for _, pod := range pods.List {
				item.AddResource(workspace, 1, pod.CpuRequest, pod.MemRequest)
			}
		}
	}

	data.Total = len(data.List)
	data.OrderBy(req.Query.OrderBy...)
	data.Paging(req.Query.PageSize, req.Query.PageNo)
	return &data, nil
}

func (p *Project) updateMemberCache(key interface{}) (*cache.Item, bool) {
	l := logrus.WithField("func", "*Project.updateMemberCache")

	unknown := &memberItem{
		ID:   0,
		Name: "OwnerUnknown",
		Nick: "OwnerUnknown",
	}
	object := newMemberCacheObject()
	cacheItem := &cache.Item{Object: object}

	applicationID := key.(uint64)
	if _, err := p.bdl.GetApp(applicationID); err != nil {
		return nil, false
	}

	members, err := p.bdl.GetMembers(apistructs.MemberListRequest{
		ScopeType: "app",
		ScopeID:   int64(applicationID),
		Roles:     []string{"Owner"},
		Labels:    nil,
		Q:         "",
		PageNo:    1,
		PageSize:  999,
	})
	if err != nil || len(members.List) == 0 {
		l.WithError(err).Errorln("failed to GetMembers")
		object.m[unknown.ID] = unknown
		return cacheItem, true
	}

	for _, member := range members.List {
		userID, err := strconv.ParseUint(member.UserID, 10, 64)
		if err != nil {
			l.WithError(err).Warnf("failed to ParseUint, member userID: %s", members.List[0].UserID)
			continue
		}
		object.m[userID] = &memberItem{
			ID:   userID,
			Name: member.Name,
			Nick: member.Nick,
		}
	}
	if len(object.m) == 0 {
		object.m[unknown.ID] = unknown
	}

	return cacheItem, true
}
