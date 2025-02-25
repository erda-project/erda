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
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/pkg/transport"
	cpb "github.com/erda-project/erda-proto-go/common/pb"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/user"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/runtime"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Service implements pb.RuntimeServiceServer
type Service struct {
	logger           logs.Logger
	bdl              *bundle.Bundle // 这个是否应该引入，因为下面已经存在一个 bundle
	bundle           BundleService
	db               DBService
	evMgr            EventManagerService
	serviceGroupImpl servicegroup.ServiceGroup
	clusterSvc       clusterpb.ClusterServiceServer
	runtime          *runtime.Runtime
	scheduler        *scheduler.Scheduler
}

func (r *Service) CreateRuntime(ctx context.Context, req *pb.RuntimeCreateRequest) (*pb.DeploymentCreateResponse, error) {
	projectInfo, err := r.bdl.GetProject(req.Extra.ProjectId)
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	clusterName, ok := projectInfo.ClusterConfig[req.Extra.Workspace]
	if !ok {
		return nil, apierrors.ErrCreateRuntime.InternalError(errors.New("cluster not found"))
	}
	req.ClusterName = clusterName
	data, err := r.runtime.Create(user.ID(req.Operator), ConvertCreateRuntimePbToDTO(req))

	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	return ConvertDeploymentCreateResponseDTOToPb(data), nil
}

func (r *Service) CreateRuntimeByReleaseAction(ctx context.Context, req *pb.RuntimeReleaseCreateRequest) (*pb.DeploymentCreateResponse, error) {
	operator := apis.GetUserID(ctx)
	data, err := r.runtime.CreateByReleaseID(ctx, user.ID(operator), ConvertRuntimeReleaseCreateRequestToDTO(req))
	if err != nil {
		return nil, errorresp.New().InternalError(err)
	}
	return ConvertDeploymentCreateResponseDTOToPb(data), nil
}

func (r *Service) ListRuntimes(ctx context.Context, req *pb.ListRuntimesRequest) (*pb.ListRuntimeResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(err)
	}
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrGetRuntime.NotLogin()
	}

	v := req.ApplicationID
	appID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrListRuntime.InvalidParameter(strutil.Concat("applicationID: ", v))
	}

	workSpace := req.WorkSpace
	name := req.Name

	data, err := r.runtime.List(user.ID(userID), uint64(orgID), appID, workSpace, name)
	if err != nil {
		return nil, errorresp.New().InternalError(err)
	}

	userIDs := make([]string, 0, len(data))
	for i := range data {
		userIDs = append(userIDs, data[i].LastOperator)
	}
	return &pb.ListRuntimeResponse{
		Data:    ConvertRuntimeSummaryToList(data),
		UserIDs: userIDs,
	}, nil
}

func (r *Service) StopRuntime(ctx context.Context, req *pb.RuntimeStopRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InvalidParameter(err)
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, apierrors.ErrDeployRuntime.NotLogin()
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v))
	}
	data, err := r.runtime.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errorresp.New().InternalError(err)
	}
	return ConvertDeploymentCreateResponseDTOToPb(data), nil
}

func (r *Service) ReDeployRuntimeAction(ctx context.Context, req *pb.ReDeployRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InvalidParameter(err)
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, apierrors.ErrDeployRuntime.NotLogin()
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v))
	}
	data, err := r.runtime.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errorresp.New().InternalError(err)
	}
	return ConvertDeploymentCreateResponseDTOToPb(data), nil
}

func (r *Service) RollBackRuntimeAction(ctx context.Context, req *pb.RollBackRuntimeActionRequest) (*pb.DeploymentCreateResponse, error) {
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, apierrors.ErrRollbackRuntime.InvalidParameter(err)
	}
	operator := apis.GetUserID(ctx)
	if operator == "" {
		return nil, apierrors.ErrRollbackRuntime.NotLogin()
	}
	if req.DeploymentID <= 0 {
		return nil, apierrors.ErrRollbackRuntime.InvalidParameter("deploymentID")
	}
	v := req.RuntimeID
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrRollbackRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v))
	}
	data, err := r.runtime.Redeploy(user.ID(operator), uint64(orgID), runtimeID)
	if err != nil {
		return nil, errorresp.New().InternalError(err)
	}
	return ConvertDeploymentCreateResponseDTOToPb(data), nil
}

func (r *Service) EpBulkGetRuntimeStatusDetail(ctx context.Context, req *pb.EpBulkGetRuntimeStatusDetailRequest) (*pb.EpBulkGetRuntimeStatusDetailResponse, error) {
	rs_ := req.RuntimeIDs
	rs := strings.Split(rs_, ",")
	var runtimeIds []uint64
	for _, r := range rs {
		runtimeId, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			// TODO: 这里返回的错误类型不知道如何操作
			return nil, errorresp.New().InternalError(err)
		}
		runtimeIds = append(runtimeIds, runtimeId)
	}
	//funcErrMsg := fmt.Sprintf("failed to bulk get runtime StatusDetail, runtimeIds: %v", runtimeIds)
	runtimes, err := r.db.FindRuntimesByIds(runtimeIds)
	if err != nil {
		// TODO: 这里返回的错误类型不知道如何操作
		return nil, errorresp.New().InternalError(err)
	}
	userId := apis.GetUserID(ctx)
	if userId == "" {
		return nil, apierrors.ErrGetRuntime.NotLogin()
	}
	data := make(map[uint64]interface{})
	for _, rt := range runtimes {
		vars := map[string]string{
			"namespace": rt.ScheduleName.Namespace,
			"name":      rt.ScheduleName.Name,
		}
		if status, err := r.scheduler.EpGetRuntimeStatus(context.Background(), vars); err != nil {
			// TODO: 这里返回的错误类型不知道如何操作
			return nil, errorresp.New().InternalError(err)
		} else {
			data[rt.ID] = status
		}
	}
	return ConvertEpBulkGetRuntimeStatusDetailToPb(data), nil
}

func (r *Service) BatchUpdateOverlay(ctx context.Context, req *pb.RuntimeScaleRecords) (*pb.BatchRuntimeScaleResults, error) {
	request := ConvertRuntimeScaleRecordsToDTO(req)
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrUpdateRuntime.NotLogin()
	}

	if len(request.Runtimes) == 0 && len(request.IDs) == 0 {
		// TODO: 这里返回的错误类型不知道如何操作
		return nil, apierrors.ErrUpdateRuntime.InvalidParameter(fmt.Sprintf("failed get diceyml.Object in table ps_v2_pre_builds for runtime ids %#v", request.IDs))
	}

	if len(request.Runtimes) != 0 && len(request.IDs) != 0 {
		return nil, apierrors.ErrUpdateRuntime.InvalidParameter("failed to batch update Overlay, runtimeRecords and ids must only one wtih non-empty values in request body")
	}

	var err error
	var runtimes []dbclient.Runtime
	if len(request.Runtimes) == 0 {
		runtimes, request.Runtimes, err = r.getRuntimeScaleRecordByRuntimeIds(request.IDs)
		if err != nil {
			// TODO: 这里返回的错误类型不知道如何操作
			logrus.Errorf("[batch redeploy] find runtimes by ids in ps_v2_project_runtimes failed, err: %v", err)
			return nil, apierrors.ErrUpdateRuntime.InternalError(err)
		}
	}

	for _, rsr := range request.Runtimes {
		perm, err := r.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   userID,
			Scope:    apistructs.AppScope,
			ScopeID:  rsr.ApplicationId,
			Resource: "runtime-" + strutil.ToLower(rsr.Workspace),
			Action:   apistructs.OperateAction,
		})
		if err != nil {
			return nil, apierrors.ErrUpdateRuntime.InternalError(err)
		}
		if !perm.Access {
			return nil, apierrors.ErrUpdateRuntime.AccessDenied()
		}
	}

	action := req.ScaleAction

	// 根据 action 的取值执行相应操作
	switch action {
	// 批量重新部署
	case apistructs.ScaleActionReDeploy:
		logrus.Infof("[batch redeploy] do batch runtimes redeploy")
		batchRuntimeReDeployResult := r.batchRuntimeReDeploy(ctx, user.ID(userID), runtimes, *request)
		if batchRuntimeReDeployResult.Failed > 0 {
			return nil, apierrors.ErrUpdateRuntime.InternalError()
			return nil, httpserver.NotOkResp(batchRuntimeReDeployResult, http.StatusInternalServerError)
		}
		logrus.Infof("[batch redeploy] redeploy all runtimes successfully")
		return nil, httpserver.OkResp(batchRuntimeReDeployResult)

	// 批量恢复
	case apistructs.ScaleActionUp:
		logrus.Infof("[batch recovery] do batch runtimes recover scale up from replicas 0 to last non-zero, will get non-zero from table ps_v2_pre_builds filed dice_overlay")
		r.batchRuntimeRecovery(request)

	// 批量停止
	case apistructs.ScaleActionDown:
		logrus.Infof("[batch scale] do batch runtimes scale down to replicas 0")
		r.batchRuntimeScaleAddRuntimeIDs(request, apistructs.ScaleActionDown)

	// 批量删除
	case apistructs.ScaleActionDelete:
		logrus.Infof("[batch delete] do batch runtimes delete")
		batchRuntimeDeleteResult := r.batchRuntimeDelete(user.ID(userID), runtimes, *request)
		if batchRuntimeDeleteResult.Failed > 0 {
			return nil, httpserver.NotOkResp(batchRuntimeDeleteResult, http.StatusInternalServerError)
		}
		logrus.Infof("[batch delete] delete all runtimes successfully")
		return nil, httpserver.OkResp(batchRuntimeDeleteResult)

	// 请求字符串指定 scale_action	参数,但对应的值为无效值
	default:
		return apierrors.ErrUpdateRuntime.InvalidParameter("invalid parameter value for parameter " + apistructs.ScaleAction + ", valid value is [scaleUp] [scaleDown] [delete] [reDeploy]").ToResp(), nil
	}

	// scale runtime
	//oldOverlayDataForAudits := make([]apistructs.PreDiceDTO,0)
	oldOverlayDataForAudits := apistructs.BatchRuntimeScaleResults{
		Total:           len(request.Runtimes),
		Successed:       0,
		Faild:           0,
		FailedScales:    make([]apistructs.RuntimeScaleRecord, 0),
		FailedIds:       make([]uint64, 0),
		SuccessedScales: make([]apistructs.PreDiceDTO, 0),
		SuccessedIds:    make([]uint64, 0),
	}
	for _, rsr := range request.Runtimes {
		oldOverlayDataForAudit, err, errMsg := r.processRuntimeScaleRecord(rsr, action)
		if err != nil {
			logrus.Errorf(errMsg)
			oldOverlayDataForAudits.Faild++
			rsr.ErrMsg = errMsg
			oldOverlayDataForAudits.FailedIds = append(oldOverlayDataForAudits.FailedIds, rsr.RuntimeID)
			oldOverlayDataForAudits.FailedScales = append(oldOverlayDataForAudits.FailedScales, rsr)
		} else {
			oldOverlayDataForAudits.Successed++
			oldOverlayDataForAudits.SuccessedIds = append(oldOverlayDataForAudits.SuccessedIds, rsr.RuntimeID)
			oldOverlayDataForAudits.SuccessedScales = append(oldOverlayDataForAudits.SuccessedScales, oldOverlayDataForAudit)
		}
	}

	if oldOverlayDataForAudits.Faild > 0 {
		return nil, nil
		return nil, httpserver.NotOkResp(oldOverlayDataForAudits, http.StatusInternalServerError)
	}
	logrus.Infof("[batch scale] scale all runtimes successfully")
	return httpserver.OkResp(oldOverlayDataForAudits)

}

func (r *Service) FullGC(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	go r.runtime.FullGC()
	return nil, nil
}

func (r *Service) ReferCluster(ctx context.Context, req *pb.ReferClusterRequest) (*pb.ReferClusterResponse, error) {
	identity := apis.GetIdentityInfo(ctx)
	if identity == nil {
		return nil, apierrors.ErrReferRuntime.NotLogin()
	}

	v := apis.GetOrgID(ctx)
	orgID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrReferRuntime.InvalidParameter(err)
	}

	clusterName := req.Cluster
	referred := r.runtime.ReferCluster(clusterName, orgID)

	return ConvertReferClusterResponseToPb(referred), nil

}

func (r *Service) RuntimeLogs(ctx context.Context, req *pb.RuntimeLogsRequest) (*pb.DashboardSpotLogData, error) {
	v := apis.GetOrgID(ctx)
	orgID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(err)
	}

	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrCreateAddon.NotLogin()
	}

	source := req.Source
	id := req.Id
	if source == "" {
		return nil, apierrors.ErrGetRuntime.MissingParameter("source")
	}
	if id == "" {
		return nil, apierrors.ErrGetRuntime.MissingParameter("id")
	}
	deploymentID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("deploymentID: ", id))
	}
	request := ConvertRuntimeLogsRequestToDTO(req)
	result, err := r.runtime.RuntimeDeployLogs(user.ID(userID), orgID, apis.GetOrg(ctx), deploymentID, request)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("deploymentID: ", id))
	}
	return ConvertDashboardSpotLogDataToPb(result), nil
}

func (r *Service) ListRuntimeByApps(ctx context.Context, req *pb.ListRuntimeByAppsRequest) (*pb.RuntimeSummary, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Service) ListMyRuntimes(ctx context.Context, req *pb.ListMyRuntimesRequest) (*pb.ListMyRuntimesResponse, error) {
	var (
		appIDs     []uint64
		env        string
		appID2Name = make(map[uint64]string)
	)
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrListRuntime.NotLogin()
	}
	v := apis.GetOrgID(ctx)
	orgID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, apierrors.ErrListRuntime.NotLogin()
	}

	projectIDStr := req.ProjectID
	if projectIDStr == "" {
		return nil, apierrors.ErrListRuntime.MissingParameter("projectID")
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return nil, apierrors.ErrListRuntime.InvalidParameter("projectID")
	}

	appIDStrs := req.AppID
	for _, str := range appIDStrs {
		id, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, apierrors.ErrGetRuntime.InvalidState("appID")
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
	myApps, err := r.bdl.GetMyAppsByProject(userID, orgID, projectID, "")
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

	runtimes, err := r.runtime.ListGroupByApps(targetAppIDs, env)
	if err != nil {
		return nil, apierrors.ErrListRuntime.InternalError(err)
	}

	var res *pb.ListMyRuntimesResponse
	for _, runtimeSummaryList := range runtimes {
		for _, runtimeSummary := range runtimeSummaryList {
			res.Data = append(res.Data, ConvertRuntimeSummaryDTOToPb(runtimeSummary))
		}
	}

	return res, nil
}

func (r *Service) CountPRByWorkspace(ctx context.Context, req *pb.CountPRByWorkspaceRequest) (*pb.CountPRByWorkspaceResponse, error) {
	var (
		l          = logrus.WithField("func", "*Endpoints.CountPRByWorkspace")
		resp       = make(map[string]uint64)
		defaultEnv = []string{"STAGING", "DEV", "PROD", "TEST"}
	)

	userId := apis.GetUserID(ctx)
	if userId == "" {
		return nil, errors.New("invalid user id")
	}

	v := apis.GetOrgID(ctx)
	orgId, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, err
	}

	projectIDStr := req.ProjectID
	if projectIDStr == "" {
		return nil, apierrors.ErrGetRuntime.InvalidParameter("projectId")
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter("projectId")
	}

	appIDStr := req.AppID
	envParam := req.WorkSpace
	if appIDStr != "" {
		appId, err := strconv.ParseUint(appIDStr, 10, 64)
		if err != nil {
			return nil, apierrors.ErrGetRuntime.InvalidParameter("appId")
		}
		if len(envParam) == 0 || envParam[0] == "" {
			for i := 0; i < len(defaultEnv); i++ {
				cnt, err := r.runtime.CountARByWorkspace(appId, defaultEnv[i])
				if err != nil {
					l.WithError(err).Warnf("count runtimes of workspace %s failed", defaultEnv[i])
				}
				resp[defaultEnv[i]] = cnt
			}
		} else {
			env := envParam[0]
			cnt, err := r.runtime.CountARByWorkspace(appId, env)
			if err != nil {
				l.WithError(err).Warnf("count runtimes of workspace %s failed", env)
			}
			resp[env] = cnt
		}
	} else {
		apps, err := r.bdl.GetMyApps(userId, orgId)
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
					cnt, err := r.runtime.CountARByWorkspace(aid, defaultEnv[i])
					if err != nil {
						l.WithError(err).Warnf("count runtimes of app %s failed", defaultEnv[i])
					}
					resp[defaultEnv[i]] += cnt
				}
			}
		} else {
			env := envParam[0]
			for aid := range appIdMap {
				cnt, err := r.runtime.CountARByWorkspace(aid, env)
				if err != nil {
					l.WithError(err).Warnf("count runtimes of workspace %s failed", env)
				}
				resp[env] += cnt
			}
		}
	}
	return ConvertCountPRByWorkspaceResponse(resp), nil
}

func (r *Service) BatchRuntimeService(ctx context.Context, req *pb.BatchRuntimeServiceRequest) (*pb.BatchRuntimeServiceResponse, error) {
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

	serviceMap, err := r.runtime.GetServiceByRuntime(runtimeIDs)
	if err != nil {
		return nil, apierrors.ErrListRuntime.InternalError(err)
	}
	return ConvertBatchRuntimeServiceResponse(serviceMap), nil
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
func (r *Service) Delete(operator user.ID, orgID uint64, runtimeID uint64) (*pb.Runtime, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrDeleteRuntime.InternalError(err)
	}
	// TODO: do not query app
	app, err := r.bundle.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, err
	}

	err = r.checkRuntimeScopePermission(operator, runtime, apistructs.DeleteAction)
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
		return nil, apierrors.ErrDeleteRuntime.InternalError(err)
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

func (r *Service) DelRuntime(ctx context.Context, req *pb.DelRuntimeRequest) (*pb.Runtime, error) {
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
		return nil, apierrors.ErrDeleteRuntime.InvalidParameter(strutil.Concat("runtimeID: ", req.Id))
	}

	return r.Delete(userID, orgID, runtimeID)
}

func (r *Service) findByIDOrName(idOrName string, appIDStr string, workspace string) (*dbclient.Runtime, error) {
	runtimeID, err := strconv.ParseUint(idOrName, 10, 64)
	if err == nil {
		// parse int success, idOrName is an id
		return r.db.GetRuntimeAllowNil(runtimeID)
	}

	// idOrName is a name
	if workspace == "" {
		return nil, apierrors.ErrGetRuntime.MissingParameter("workspace")
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("applicationID: ", appIDStr))
	}

	// TODO: we shall not un-escape runtimeName, after we fix existing data and deny '/'
	name, err := url.PathUnescape(idOrName)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("idOrName: ", idOrName))
	}

	runtime, err := r.db.FindRuntime(spec.RuntimeUniqueId{Name: name, Workspace: workspace, ApplicationId: appID})
	if err == nil && runtime == nil {
		return nil, apierrors.ErrGetRuntime.NotFound()
	}

	return runtime, err
}

func (r *Service) getUserAndOrgID(ctx context.Context) (userID user.ID, orgID uint64, err error) {
	orgIntID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		err = apierrors.ErrGetRuntime.InvalidParameter(errors.New("Org-ID"))
		return
	}

	orgID = uint64(orgIntID)

	userID = user.ID(apis.GetUserID(ctx))
	if userID.Invalid() {
		err = apierrors.ErrGetRuntime.NotLogin()
		return
	}

	return
}

func (r *Service) checkRuntimeScopePermission(userID user.ID, runtime *dbclient.Runtime, action string) error {
	perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   action,
	})

	if err != nil {
		return err
	}

	if !perm.Access {
		return apierrors.ErrGetRuntime.AccessDenied()
	}

	return nil
}

func (r *Service) getRuntimeByRequest(request *pb.GetRuntimeRequest) (*dbclient.Runtime, error) {
	runtime, err := r.findByIDOrName(request.NameOrID, request.AppID, request.Workspace)

	if err != nil {
		return nil, err
	}

	if runtime == nil {
		return nil, apierrors.ErrGetRuntime.NotFound()
	}

	return runtime, nil
}

func (r *Service) getDeployment(runtimeID uint64) (*dbclient.Deployment, error) {
	deployment, err := r.db.FindLastDeployment(runtimeID)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InternalError(err)
	}

	if deployment == nil {
		return nil, apierrors.ErrGetRuntime.InvalidState("last deployment not found")
	}

	return deployment, nil
}

// GetRuntime Get detail information of a single runtime
func (r *Service) GetRuntime(ctx context.Context, request *pb.GetRuntimeRequest) (*pb.RuntimeInspect, error) {
	var (
		err        error
		userID     user.ID
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

	if userID, _, err = r.getUserAndOrgID(ctx); err != nil {
		return nil, err
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
	if err = r.checkRuntimeScopePermission(userID, runtime, apistructs.GetAction); err != nil {
		return nil, err
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
		return nil, apierrors.ErrGetRuntime.InvalidState(strutil.Concat("dice.json invalid: ", err.Error()))
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

func (r *Service) CheckRuntimeExist(ctx context.Context, req *pb.CheckRuntimeExistReq) (*pb.CheckRuntimeExistResp, error) {
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

func (r *Service) Healthz(ctx context.Context, req *cpb.VoidRequest) (*cpb.VoidResponse, error) {
	return &cpb.VoidResponse{}, nil
}

func (r *Service) inspectDomains(id uint64) (map[string][]string, error) {
	domains, err := r.db.FindDomainsByRuntimeId(id)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InternalError(err)
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

type ServiceOption func(*Service) *Service

func WithBundleService(s BundleService) ServiceOption {
	return func(service *Service) *Service {
		service.bundle = s
		return service
	}
}

func WithDBService(db DBService) ServiceOption {
	return func(service *Service) *Service {
		service.db = db
		return service
	}
}

func WithEventManagerService(evMgr EventManagerService) ServiceOption {
	return func(service *Service) *Service {
		service.evMgr = evMgr
		return service
	}
}

func WithServiceGroupImpl(serviceGroupImpl servicegroup.ServiceGroup) ServiceOption {
	return func(service *Service) *Service {
		service.serviceGroupImpl = serviceGroupImpl
		return service
	}
}

func WithClusterSvc(clusterSvc clusterpb.ClusterServiceServer) ServiceOption {
	return func(service *Service) *Service {
		service.clusterSvc = clusterSvc
		return service
	}
}

var server pb.RuntimeServiceServer = &Service{}

func NewRuntimeService(options ...ServiceOption) pb.RuntimeServiceServer {
	s := &Service{}

	for _, option := range options {
		option(s)
	}

	return s
}
