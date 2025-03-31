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

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/pkg/transport"
	cpb "github.com/erda-project/erda-proto-go/common/pb"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	dicehubpb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	releasepb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/user"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	errors "github.com/erda-project/erda/pkg/common/errors"
	perm "github.com/erda-project/erda/pkg/common/permission"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// RuntimeService implements pb.RuntimeServiceServer
type RuntimeService struct {
	logger           logs.Logger
	bundle           *bundle.Bundle
	db               DBService
	evMgr            *events.EventManager
	serviceGroupImpl servicegroup.ServiceGroup
	clusterSvc       clusterpb.ClusterServiceServer
	clusterinfoImpl  clusterinfo.ClusterInfo
	scheduler        *scheduler.Scheduler
	org              org.ClientInterface
	Addon            *addon.Addon
	releaseSvc       releasepb.ReleaseServiceServer
	pipelineSvc      pipelinepb.PipelineServiceServer
}

func (r *RuntimeService) KillPodService(ctx context.Context, req *pb.KillPodRequest) (*pb.KillPodResponse, error) {
	auditData, err := r.KillPod(req.RuntimeID, req.PodName)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.KillPodResponse{
		ApplicationID: auditData["applicationID"],
		Workspace:     auditData["workspace"],
		Runtime:       auditData["runtime"],
		PodName:       auditData["podName"],
		ProjectName:   auditData["projectName"],
		AppName:       auditData["appName"],
	}, nil
}

func (r *RuntimeService) RollBackRuntime(ctx context.Context, req *pb.RollBackRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "not found org-id")
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	if req.DeploymentId <= 0 {
		return nil, errors.NewInvalidParameterError("deploymentID", "not found deploymentID")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("runtimeID", v)
	}
	data, err := r.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) ReDeployRuntime(ctx context.Context, req *pb.ReDeployRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "not found org-id")
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("runtimeID", v)
	}
	data, err := r.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) CreateRuntimeByRelease(ctx context.Context, req *pb.RuntimeReleaseCreateRequest) (*pb.DeploymentCreateResponse, error) {
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}

	release := &apistructs.RuntimeReleaseCreateRequest{
		ReleaseID:     req.ReleaseId,
		Workspace:     req.Workspace,
		ProjectID:     req.ProjectId,
		ApplicationID: req.ApplicationId,
	}

	data, err := r.CreateByReleaseID(ctx, user.ID(operator), release)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) CreateRuntime(ctx context.Context, req *pb.RuntimeCreateRequest) (*pb.DeploymentCreateResponse, error) {
	if req.Extra == nil {
		return nil, errors.NewInvalidParameterError("extra", "extra is nil")
	}
	projectInfo, err := r.bundle.GetProject(req.Extra.ProjectId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	clusterName, ok := projectInfo.ClusterConfig[req.Extra.Workspace]
	if !ok {
		return nil, errors.NewInternalServerErrorMessage("cluster not found")
	}
	req.ClusterName = clusterName

	// transform pb to apistructs
	request, err := ConvertCreatRuntimeRequest(req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	// create Runtime
	data, err := r.Create(user.ID(req.Operator), request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) CreateRuntimeByReleaseAction(ctx context.Context, req *pb.RuntimeReleaseCreateRequest) (*pb.DeploymentCreateResponse, error) {
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}

	release := &apistructs.RuntimeReleaseCreateRequest{
		ReleaseID:     req.ReleaseId,
		Workspace:     req.Workspace,
		ProjectID:     req.ProjectId,
		ApplicationID: req.ApplicationId,
	}

	data, err := r.CreateByReleaseID(ctx, user.ID(operator), release)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) ListRuntimes(ctx context.Context, req *pb.ListRuntimesRequest) (*pb.ListRuntimeResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "")
	}
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}

	v := req.ApplicationID
	appID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("applicationID", v)
	}

	workSpace := req.WorkSpace
	name := req.Name

	data, err := r.List(user.ID(userID), uint64(orgID), appID, workSpace, name)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	userIDs := make([]string, 0, len(data))
	for i := range data {
		userIDs = append(userIDs, data[i].LastOperator)
	}
	return ConvertListRuntimeResponse(data, userIDs), nil
}

func (r *RuntimeService) StopRuntime(ctx context.Context, req *pb.RuntimeStopRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "not found org-id")
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("runtimeID", v)
	}
	data, err := r.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) ReDeployRuntimeAction(ctx context.Context, req *pb.ReDeployRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "not found org-id")
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("runtimeID", v)
	}
	data, err := r.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) RollBackRuntimeAction(ctx context.Context, req *pb.RollBackRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, errors.NewInvalidParameterError("org-id", "not found org-id")
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	if req.DeploymentId <= 0 {
		return nil, errors.NewInvalidParameterError("deploymentID", "not found deploymentID")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("runtimeID", v)
	}
	data, err := r.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return &pb.DeploymentCreateResponse{
		DeploymentId:  data.DeploymentID,
		ApplicationId: data.ApplicationID,
		RuntimeId:     data.RuntimeID,
	}, nil
}

func (r *RuntimeService) FullGC(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	go r.FullGCService()
	return nil, nil
}

func (r *RuntimeService) ListRuntimesGroupByApps(ctx context.Context, req *pb.ListRuntimeByAppsRequest) (*pb.ListRuntimeByAppsResponse, error) {
	var (
		l      = logrus.WithField("func", "*Endpoints.ListRuntimesGroupByApps")
		appIDs []uint64
		env    string
	)

	for _, appID := range req.ApplicationID {
		id, err := strconv.ParseUint(appID, 10, 64)
		if err != nil {
			l.WithError(err).Warnf("failed to parse applicationID: failed to ParseUint: %s", appID)
		}
		appIDs = append(appIDs, id)
	}
	envParam := req.Workspace

	if len(envParam) == 0 {
		env = ""
	} else {
		env = envParam[0]
	}
	runtimes, err := r.ListGroupByApps(appIDs, env)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	resp := &pb.ListRuntimeByAppsResponse{
		Data: make(map[uint64]*structpb.Value),
	}
	for i, runtimeList := range runtimes {

		runtimeListBytes, err := json.Marshal(runtimeList)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}

		var structValue structpb.Value

		if err := json.Unmarshal(runtimeListBytes, &structValue); err != nil {
			return nil, errors.NewInternalServerError(err)
		}

		resp.Data[i] = &structValue
	}

	return resp, nil
}

func (r *RuntimeService) ListMyRuntimes(ctx context.Context, req *pb.ListMyRuntimesRequest) (*pb.ListMyRuntimesResponse, error) {
	var (
		appIDs     []uint64
		env        string
		appID2Name = make(map[uint64]string)
	)
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, errors.NewUnauthorizedError("not Login")
	}
	v := apis.GetOrgID(ctx)
	orgID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, errors.NewUnauthorizedError("not Login")
	}

	projectIDStr := req.ProjectID
	if projectIDStr == "" {
		return nil, errors.NewMissingParameterError("projectID")
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("projectID", projectIDStr)
	}

	appIDStrs := req.AppID
	for _, str := range appIDStrs {
		id, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, errors.NewInvalidParameterError("appID", str)
		}
		appIDs = append(appIDs, id)
	}

	envParam := req.WorkSpace
	if len(envParam) == 0 {
		env = ""
	} else {
		env = envParam[0]
	}

	var myAppIDs []uint64
	myApps, err := r.bundle.GetMyAppsByProject(userID, orgID, projectID, "")
	for i := range myApps.List {
		myAppIDs = append(myAppIDs, myApps.List[i].ID)
		appID2Name[myApps.List[i].ID] = myApps.List[i].Name
	}

	var targetAppIDs []uint64
	if len(appIDs) == 0 {
		targetAppIDs = myAppIDs
	} else {
		for _, id := range appIDs {
			if _, ok := appID2Name[id]; ok {
				targetAppIDs = append(targetAppIDs, id)
			}
		}
	}

	runtimes, err := r.ListGroupByApps(targetAppIDs, env)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	resp := &pb.ListMyRuntimesResponse{}

	for _, runtimeSummaryList := range runtimes {
		for _, runtimeSummary := range runtimeSummaryList {
			var data *pb.RuntimeSummary
			dataBytes, err := json.Marshal(runtimeSummary)
			if err != nil {
				return nil, errors.NewInternalServerError(err)
			}
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return nil, errors.NewInternalServerError(err)
			}
			resp.Data = append(resp.Data, data)
		}
	}

	return resp, nil
}

func (r *RuntimeService) CheckCountPRByWorkspacePerm(userID string, applicationId uint64, resources []string) error {
	scope := perm.ScopeApp
	for _, resource := range resources {
		// 鉴权
		resp, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   userID,
			Scope:    perm.ScopeApp,
			ScopeID:  applicationId,
			Resource: resource,
			Action:   perm.ActionGet,
		})
		if err != nil {
			return errors.NewServiceInvokingError("CheckPermission", err)
		}
		if !resp.Access {
			return errors.NewPermissionError(fmt.Sprintf("user/%s/%s/%s/resource/%s", userID, scope, applicationId, resource), perm.ActionGet, "")
		}
	}
	return nil
}

func (r *RuntimeService) CountPRByWorkspace(ctx context.Context, req *pb.CountPRByWorkspaceRequest) (*pb.CountPRByWorkspaceResponse, error) {
	var (
		l          = logrus.WithField("func", "*Endpoints.CountPRByWorkspace")
		resp       = make(map[string]uint64)
		defaultEnv = []string{"STAGING", "DEV", "PROD", "TEST"}
	)

	userId := apis.GetUserID(ctx)
	if userId == "" {
		return nil, errors.NewInvalidParameterError("userId", "invalid userId")
	}

	v := apis.GetOrgID(ctx)
	if v == "" {
		return nil, errors.NewInvalidParameterError("orgId", "invalid orgId")
	}
	orgId, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, err
	}

	projectIDStr := req.ProjectId
	if projectIDStr == "" {
		return nil, errors.NewMissingParameterError("projectID")
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("projectID", projectIDStr)
	}

	appIDStr := req.AppID
	envParam := req.WorkSpace
	if appIDStr != "" {
		appId, err := strconv.ParseUint(appIDStr, 10, 64)
		if err != nil {
			return nil, errors.NewInvalidParameterError("appID", appIDStr)
		}
		if len(envParam) == 0 || envParam[0] == "" {
			// 鉴权
			if err := r.CheckCountPRByWorkspacePerm(userId, appId, []string{"runtime-staging", "runtime-dev", "runtime-prod", "runtime-test"}); err != nil {
				return nil, err
			}

			for i := 0; i < len(defaultEnv); i++ {
				cnt, err := r.CountARByWorkspace(appId, defaultEnv[i])
				if err != nil {
					l.WithError(err).Warnf("count runtimes of workspace %s failed", defaultEnv[i])
				}
				resp[defaultEnv[i]] = cnt
			}
		} else {
			// 鉴权
			if err := r.CheckCountPRByWorkspacePerm(userId, appId, []string{GetRuntimeResource(envParam[0])}); err != nil {
				return nil, err
			}
			env := envParam[0]
			cnt, err := r.CountARByWorkspace(appId, env)
			if err != nil {
				l.WithError(err).Warnf("count runtimes of workspace %s failed", env)
			}
			resp[env] = cnt
		}
	} else {
		apps, err := r.bundle.GetMyApps(userId, orgId)
		if err != nil {
			return nil, err
		}
		appIdMap := make(map[uint64]bool)
		for i := 0; i < len(apps.List); i++ {
			if apps.List[i].ProjectID == projectID {
				appIdMap[apps.List[i].ID] = true
			}
		}
		if len(envParam) == 0 || envParam[0] == "" {
			for i := 0; i < len(defaultEnv); i++ {
				for aid := range appIdMap {
					// 鉴权
					if err := r.CheckCountPRByWorkspacePerm(userId, aid, []string{GetRuntimeResource(defaultEnv[i])}); err != nil {
						return nil, err
					}
					cnt, err := r.CountARByWorkspace(aid, defaultEnv[i])
					if err != nil {
						l.WithError(err).Warnf("count runtimes of app %s failed", defaultEnv[i])
					}
					resp[defaultEnv[i]] += cnt
				}
			}
		} else {
			env := envParam[0]
			for aid := range appIdMap {
				// 鉴权
				if err := r.CheckCountPRByWorkspacePerm(userId, aid, []string{GetRuntimeResource(env)}); err != nil {
					return nil, err
				}
				cnt, err := r.CountARByWorkspace(aid, env)
				if err != nil {
					l.WithError(err).Warnf("count runtimes of workspace %s failed", env)
				}
				resp[env] += cnt
			}
		}
	}
	return &pb.CountPRByWorkspaceResponse{Data: resp}, nil
}

func (r *RuntimeService) BatchRuntimeService(ctx context.Context, req *pb.BatchRuntimeServiceRequest) (*pb.BatchRuntimeServiceResponse, error) {
	var (
		l          = logrus.WithField("func", "*Endpoints.BatchRuntimeServices")
		runtimeIDs []uint64
	)

	for _, runtimeID := range req.RuntimeID {
		id, err := strconv.ParseUint(runtimeID, 10, 64)
		if err != nil {
			l.WithError(err).Warnf("failed to parse applicationID: failed to ParseUint: %s", runtimeID)
		}
		runtimeIDs = append(runtimeIDs, id)
	}

	serviceMap, err := r.GetServiceByRuntime(runtimeIDs)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp := &pb.BatchRuntimeServiceResponse{
		Data: make(map[uint64]*structpb.Value),
	}
	for k, v := range serviceMap {
		var structValue structpb.Value
		vBytes, err := json.Marshal(v)
		if err != nil {
			return nil, errors.NewInternalServerErrorMessage("Marshal failed")
		}
		if err := json.Unmarshal(vBytes, &structValue); err != nil {
			return nil, errors.NewInternalServerErrorMessage("UnMarshal failed")
		}
		resp.Data[k] = &structValue
	}

	return resp, nil
}

func ConvertListRuntimeResponse(data []apistructs.RuntimeSummaryDTO, userIds []string) *pb.ListRuntimeResponse {
	if data == nil {
		return nil
	}
	var tmp = make(map[string]interface{})
	tmp["data"] = data
	tmp["userIDs"] = userIds
	dataBytes, err := json.Marshal(tmp)
	var result *pb.ListRuntimeResponse
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(dataBytes, &result)
	return result
}

func ConvertCreatRuntimeRequest(req *pb.RuntimeCreateRequest) (*apistructs.RuntimeCreateRequest, error) {
	addonActions := make(map[string]interface{})
	addonActionsBytes, err := json.Marshal(req.Extra.AddonActions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(addonActionsBytes, &addonActions)
	return &apistructs.RuntimeCreateRequest{
		Name:        req.Name,
		ReleaseID:   req.ReleaseId,
		Operator:    req.Operator,
		ClusterName: req.ClusterName,
		Source:      apistructs.RuntimeSource(req.Source),
		Extra: apistructs.RuntimeCreateRequestExtra{
			OrgID:           req.Extra.OrgId,
			ProjectID:       req.Extra.ProjectId,
			ApplicationID:   req.Extra.ApplicationId,
			ApplicationName: req.Extra.ApplicationName,
			Workspace:       req.Extra.Workspace,
			BuildID:         req.Extra.BuildId,
			DeployType:      req.Extra.DeployType,
			InstanceID:      json.Number(req.Extra.InstanceId),
			ClusterId:       json.Number(req.Extra.ClusterId),
			AddonActions:    addonActions,
		},
		SkipPushByOrch:    req.SkipPushByOrch,
		Param:             req.Param,
		DeploymentOrderId: req.DeploymentOrderId,
		ReleaseVersion:    req.ReleaseVersion,
		ExtraParams:       req.ExtraParams,
	}, nil
}

func convertRuntimeToPB(runtime *dbclient.Runtime, app *apistructs.ApplicationDTO) *pb.Runtime {
	return &pb.Runtime{
		Name:            runtime.Name,
		GitBranch:       runtime.Name,
		Workspace:       runtime.Workspace,
		ClusterName:     runtime.ClusterName,
		Status:          runtime.Status,
		ClusterID:       runtime.ClusterId,
		ApplicationID:   runtime.ApplicationID,
		ApplicationName: app.Name,
		ProjectID:       app.ProjectID,
		ProjectName:     app.ProjectName,
		OrgID:           app.OrgID,
		Id:              runtime.ID,
	}
}

// Delete turn status of runtime to be Deleting
func (r *RuntimeService) Delete(operator user.ID, orgID uint64, runtimeID uint64) (*pb.Runtime, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	// TODO: do not query app
	app, err := r.bundle.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, err
	}

	if runtime.LegacyStatus == dbclient.LegacyStatusDeleting {
		// already marked
		return convertRuntimeToPB(runtime, app), nil
	}
	// set status to DELETING
	runtime.LegacyStatus = dbclient.LegacyStatusDeleting
	if err := r.db.UpdateRuntime(runtime); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	event := events.RuntimeEvent{
		EventName: events.RuntimeDeleting,
		Runtime:   dbclient.ConvertRuntimeDTO(runtime, app),
		Operator:  operator.String(),
	}
	r.evMgr.EmitEvent(&event)
	// TODO: should emit RuntimeDeleted after really deleted or RuntimeDeleteFailed if failed
	return convertRuntimeToPB(runtime, app), nil
}

func (r *RuntimeService) DelRuntime(ctx context.Context, req *pb.DelRuntimeRequest) (*pb.Runtime, error) {
	var (
		userID    user.ID
		orgID     uint64
		err       error
		runtimeID uint64
	)

	if userID, _, err = r.getUserAndOrgID(ctx); err != nil {
		return nil, err
	}

	runtimeID, err = strconv.ParseUint(req.Id, 10, 64)

	if err != nil {
		return nil, errors.NewInvalidParameterError("runtime_id", req.Id)
	}

	return r.Delete(userID, orgID, runtimeID)
}

func (r *RuntimeService) findByIDOrName(idOrName string, appIDStr string, workspace string) (*dbclient.Runtime, error) {
	runtimeID, err := strconv.ParseUint(idOrName, 10, 64)
	if err == nil {
		// parse int success, idOrName is an id
		return r.db.GetRuntimeAllowNil(runtimeID)
	}

	// idOrName is a name
	if workspace == "" {
		return nil, errors.NewMissingParameterError("workspace")
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("applicationID: ", appIDStr)
	}

	// TODO: we shall not un-escape runtimeName, after we fix existing data and deny '/'
	name, err := url.PathUnescape(idOrName)
	if err != nil {
		return nil, errors.NewInvalidParameterError("idOrName: ", idOrName)
	}

	runtime, err := r.db.FindRuntime(spec.RuntimeUniqueId{Name: name, Workspace: workspace, ApplicationId: appID})
	if err == nil && runtime == nil {
		return nil, errors.NewNotFoundError("runtime")
	}

	return runtime, err
}

func (r *RuntimeService) getUserAndOrgID(ctx context.Context) (userID user.ID, orgID uint64, err error) {
	orgIntID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		err = errors.NewInvalidParameterError("org-id", "invalid org-id")
		return
	}

	orgID = uint64(orgIntID)

	userID = user.ID(apis.GetUserID(ctx))
	if userID.Invalid() {
		err = errors.NewUnauthorizedError("not Login")
		return
	}

	return
}

func (r *RuntimeService) getRuntimeByRequest(request *pb.GetRuntimeRequest) (*dbclient.Runtime, error) {
	runtime, err := r.findByIDOrName(request.NameOrID, request.AppID, request.Workspace)

	if err != nil {
		return nil, err
	}

	if runtime == nil {
		return nil, errors.NewNotFoundError("runtime")
	}

	return runtime, nil
}

func (r *RuntimeService) getDeployment(runtimeID uint64) (*dbclient.Deployment, error) {
	deployment, err := r.db.FindLastDeployment(runtimeID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	if deployment == nil {
		return nil, errors.NewWarnError("last deployment not found")
	}

	return deployment, nil
}

// GetRuntime Get detail information of a single runtime
func (r *RuntimeService) GetRuntime(ctx context.Context, request *pb.GetRuntimeRequest) (*pb.RuntimeInspect, error) {
	var (
		err        error
		runtime    *dbclient.Runtime
		deployment *dbclient.Deployment
		domainMap  map[string][]string
		cluster    *clusterpb.ClusterInfo
		sg         *apistructs.ServiceGroup
		app        *apistructs.ApplicationDTO
		ri         *pb.RuntimeInspect
		dice       diceyml.Object
	)

	ri = &pb.RuntimeInspect{
		Services:    make(map[string]*pb.Service),
		LastMessage: make(map[string]*pb.StatusMap),
	}

	if runtime, err = r.getRuntimeByRequest(request); err != nil {
		return nil, err
	}

	hpaRules, err := r.db.GetRuntimeHPARulesByRuntimeId(runtime.ID)
	if err != nil {
		logrus.Warnf("[GetRuntime] get hpa rules for runtimeId %d failed.", runtime.ID)
	}

	vpaRules, err := r.db.GetRuntimeVPARulesByRuntimeId(runtime.ID)
	if err != nil {
		logrus.Warnf("[GetRuntime] get vpa rules for runtimeId %d failed.", runtime.ID)
	}

	if deployment, err = r.getDeployment(runtime.ID); err != nil {
		return nil, err
	}

	if domainMap, err = r.inspectDomains(runtime.ID); err != nil {
		return nil, err
	}

	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "cmp"}))
	resp, err := r.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{IdOrName: runtime.ClusterName})
	if err != nil {
		return nil, err
	}
	cluster = resp.Data

	if runtime.ScheduleName.Name != "" {
		sg, _ = r.serviceGroupImpl.InspectServiceGroupWithTimeout(runtime.ScheduleName.Args())
	}

	if app, err = r.bundle.GetApp(runtime.ApplicationID); err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(deployment.Dice), &dice); err != nil {
		return nil, errors.NewWarnError(strutil.Concat("dice.json invalid: ", err.Error()))
	}

	fillInspectBase(ri, runtime, sg)
	fillInspectByDeployment(ri, runtime, deployment, cluster)
	fillInspectByApp(ri, runtime, app)
	fillInspectDataWithServiceGroup(ri, dice.Services, dice.Jobs, sg, domainMap, string(deployment.Status))
	updateStatusToDisplay(ri)
	if deployment.Status == apistructs.DeploymentStatusDeploying {
		updateStatusWhenDeploying(ri)
	}
	updatePARuleEnabledStatusToDisplay(hpaRules, vpaRules, ri)

	return ri, nil
}

func (r *RuntimeService) GetRuntimeWorkspaceByID(RuntimeId uint64) (*dbclient.Runtime, error) {
	return nil, nil
}

func (r *RuntimeService) CheckRuntimeExist(ctx context.Context, req *pb.CheckRuntimeExistReq) (*pb.CheckRuntimeExistResp, error) {
	runtime, err := r.db.GetRuntimeAllowNil(req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.CheckRuntimeExistResp{Ok: runtime != nil}, nil
}

func updateStatusWhenDeploying(runtime *pb.RuntimeInspect) {
	if runtime == nil {
		return
	}
	if runtime.Status == "UnHealthy" {
		runtime.Status = "Progressing"
	}
	for _, v := range runtime.Services {
		if v.Status == "UnHealthy" {
			v.Status = "Progressing"
		}
	}
}

func updateStatusToDisplay(runtime *pb.RuntimeInspect) {
	if runtime == nil {
		return
	}
	runtime.Status = isStatusForDisplay(runtime.Status)
	for key := range runtime.Services {
		runtime.Services[key].Status = isStatusForDisplay(runtime.Services[key].Status)
	}
}

func isStatusForDisplay(status string) string {
	switch status {
	case apistructs.RuntimeStatusHealthy, apistructs.RuntimeStatusUnHealthy, apistructs.RuntimeStatusInit:
		return status
	case "Ready", "ready":
		return apistructs.RuntimeStatusHealthy
	default:
		return status
	}
}

func fillInspectByApp(data *pb.RuntimeInspect, runtime *dbclient.Runtime, app *apistructs.ApplicationDTO) {
	data.ProjectID = app.ProjectID
	data.ApplicationName = app.Name
	data.CreatedAt = timestamppb.New(runtime.CreatedAt)
	data.UpdatedAt = timestamppb.New(runtime.UpdatedAt)
	data.TimeCreated = timestamppb.New(runtime.CreatedAt)
}

func fillInspectByDeployment(data *pb.RuntimeInspect, runtime *dbclient.Runtime, deployment *dbclient.Deployment, cluster *clusterpb.ClusterInfo) {
	data.DeployStatus = string(deployment.Status)
	if deployment.Status == apistructs.DeploymentStatusDeploying ||
		deployment.Status == apistructs.DeploymentStatusWaiting ||
		deployment.Status == apistructs.DeploymentStatusInit ||
		deployment.Status == apistructs.DeploymentStatusWaitApprove {
		data.Status = apistructs.RuntimeStatusInit
	}

	if runtime.LegacyStatus == "DELETING" {
		data.DeleteStatus = "DELETING"
	}
	data.ReleaseID = deployment.ReleaseId
	data.ClusterID = runtime.ClusterId
	data.ClusterName = runtime.ClusterName
	data.ClusterType = cluster.Type
	data.Extra = &pb.Extra{
		ApplicationID: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		BuildID:       deployment.BuildId,
	}
}

func fillInspectDataWithServiceGroup(data *pb.RuntimeInspect, targetService diceyml.Services, targetJob diceyml.Jobs,
	sg *apistructs.ServiceGroup, domainMap map[string][]string, status string) {
	statusServiceMap := map[string]string{}
	replicaMap := map[string]int{}
	resourceMap := map[string]*pb.Resources{}
	statusMap := make(map[string]*pb.StatusMap)
	if sg != nil {
		if sg.Status != apistructs.StatusReady && sg.Status != apistructs.StatusHealthy {
			for _, serviceItem := range sg.Services {
				statusMap[serviceItem.Name] = &pb.StatusMap{
					Msg:    serviceItem.LastMessage,
					Reason: serviceItem.Reason,
				}
			}
		}

		data.LastMessage = statusMap

		for _, v := range sg.Services {
			statusServiceMap[v.Name] = string(v.StatusDesc.Status)
			replicaMap[v.Name] = int(v.DesiredReplicas)
			resourceMap[v.Name] = &pb.Resources{
				Cpu:  v.Resources.Cpu,
				Mem:  int64(int(v.Resources.Mem)),
				Disk: int64(int(v.Resources.Disk)),
			}
		}
	}

	// TODO: no diceJson and no overlay, we just read dice from releaseId
	for k, v := range targetService {
		var expose []string
		var svcPortExpose bool
		// serv.Expose will abandoned, serv.Ports.Expose is recommended
		for _, svcPort := range v.Ports {
			if svcPort.Expose {
				svcPortExpose = true
			}
		}
		if len(v.Expose) != 0 || svcPortExpose {
			expose = domainMap[k]
		}

		runtimeInspectService := &pb.Service{
			Resources: &pb.Resources{
				Cpu:  v.Resources.CPU,
				Mem:  int64(v.Resources.Mem),
				Disk: int64(v.Resources.Disk),
			},
			Envs:        v.Envs,
			Addrs:       convertInternalAddrs(sg, k),
			Expose:      expose,
			Status:      status,
			Deployments: &pb.Deployments{Replicas: 0},
		}

		if sgStatus, ok := statusServiceMap[k]; ok {
			runtimeInspectService.Status = sgStatus
		}
		if sgReplicas, ok := replicaMap[k]; ok {
			runtimeInspectService.Deployments.Replicas = uint64(sgReplicas)
		}
		if sgResources, ok := resourceMap[k]; ok {
			runtimeInspectService.Resources = sgResources
		}

		data.Services[k] = runtimeInspectService
	}
	for k, v := range targetJob {
		runtimeInspectService := &pb.Service{
			Resources: &pb.Resources{
				Cpu:  v.Resources.CPU,
				Mem:  int64(v.Resources.Mem),
				Disk: int64(v.Resources.Disk),
			},
			Envs:        v.Envs,
			Type:        "job",
			Status:      statusServiceMap[k],
			Deployments: &pb.Deployments{Replicas: 1},
		}
		if sgReplicas, ok := replicaMap[k]; ok {
			runtimeInspectService.Deployments.Replicas = uint64(sgReplicas)
		}
		data.Services[k] = runtimeInspectService
	}
	data.Resources = &pb.Resources{Cpu: 0, Mem: 0, Disk: 0}
	for _, v := range data.Services {
		if v.Type == "job" {
			continue
		}
		data.Resources.Cpu += v.Resources.Cpu * float64(v.Deployments.Replicas)
		data.Resources.Mem += v.Resources.Mem * int64(v.Deployments.Replicas)
		data.Resources.Disk += v.Resources.Disk * int64(v.Deployments.Replicas)
	}
}

func convertInternalAddrs(sg *apistructs.ServiceGroup, serviceName string) []string {
	addrs := make([]string, 0)
	if sg == nil {
		return addrs
	}
	for _, s := range sg.Services {
		if s.Name != serviceName {
			continue
		}
		for _, p := range s.Ports {
			addrs = append(addrs, s.Vip+":"+strconv.Itoa(p.Port))
		}
	}
	return addrs
}

func fillInspectBase(data *pb.RuntimeInspect, runtime *dbclient.Runtime, sg *apistructs.ServiceGroup) {
	data.Id = runtime.ID
	data.Name = runtime.Name
	data.ServiceGroupNamespace = runtime.ScheduleName.Namespace
	data.ServiceGroupName = runtime.ScheduleName.Name
	data.Source = string(runtime.Source)
	data.Status = apistructs.RuntimeStatusUnHealthy
	if runtime.ScheduleName.Namespace != "" && runtime.ScheduleName.Name != "" && sg != nil {
		if sg.Status == "Ready" || sg.Status == "Healthy" {
			data.Status = apistructs.RuntimeStatusHealthy
		}
	}
	if runtime.LegacyStatus == "DELETING" {
		data.DeleteStatus = "DELETING"
	}
}

func (r *RuntimeService) Healthz(ctx context.Context, req *cpb.VoidRequest) (*cpb.VoidResponse, error) {
	return &cpb.VoidResponse{}, nil
}

func (r *RuntimeService) inspectDomains(id uint64) (map[string][]string, error) {
	domains, err := r.db.FindDomainsByRuntimeId(id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	domainMap := make(map[string][]string)
	for _, d := range domains {
		if domainMap[d.EndpointName] == nil {
			domainMap[d.EndpointName] = make([]string, 0)
		}
		domainMap[d.EndpointName] = append(domainMap[d.EndpointName], "http://"+d.Domain)
	}

	return domainMap, nil
}

// 显示 service 对应是否开启 HPA, VPA
func updatePARuleEnabledStatusToDisplay(hpaRules []dbclient.RuntimeHPA, vpaRules []dbclient.RuntimeVPA, runtime *pb.RuntimeInspect) {
	if runtime == nil {
		return
	}

	for svc := range runtime.Services {
		runtime.Services[svc].HpaEnabled = pstypes.RuntimePARuleCanceled
		runtime.Services[svc].VpaEnabled = pstypes.RuntimePARuleCanceled
	}

	for _, rule := range hpaRules {
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			runtime.Services[rule.ServiceName].HpaEnabled = pstypes.RuntimePARuleApplied
		}
	}

	for _, rule := range vpaRules {
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			runtime.Services[rule.ServiceName].VpaEnabled = pstypes.RuntimePARuleApplied
		}
	}
}

type ServiceOption func(*RuntimeService) *RuntimeService

func WithBundleService(s *bundle.Bundle) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.bundle = s
		return service
	}
}

func WithDBService(db DBService) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.db = db
		return service
	}
}

func WithEventManagerService(evMgr *events.EventManager) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.evMgr = evMgr
		return service
	}
}

func WithServiceGroupImpl(serviceGroupImpl servicegroup.ServiceGroup) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.serviceGroupImpl = serviceGroupImpl
		return service
	}
}

func WithClusterSvc(clusterSvc clusterpb.ClusterServiceServer) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.clusterSvc = clusterSvc
		return service
	}
}

func WithReleaseSvc(releaseSvc dicehubpb.ReleaseServiceServer) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.releaseSvc = releaseSvc
		return service
	}
}

func WithOrg(org org.ClientInterface) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.org = org
		return service
	}
}

func WithClusterInfoImpl(clusterInfo clusterinfo.ClusterInfo) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.clusterinfoImpl = clusterInfo
		return service
	}
}

func WithScheduler(scheduler *scheduler.Scheduler) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.scheduler = scheduler
		return service
	}
}

func WithAddon(addon *addon.Addon) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.Addon = addon
		return service
	}
}

func WithPipelineSvc(pipelineSvc pipelinepb.PipelineServiceServer) ServiceOption {
	return func(service *RuntimeService) *RuntimeService {
		service.pipelineSvc = pipelineSvc
		return service
	}
}

func NewRuntimeService(options ...ServiceOption) *RuntimeService {
	s := &RuntimeService{}

	for _, option := range options {
		option(s)
	}

	return s
}
