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

package deployment_order

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	FirstBatch        = iota + 1
	DeployModeEnvName = "ERDA_DEPLOY_MODES"
	DefaultMode       = "default"
)

func (d *DeploymentOrder) Create(ctx context.Context, req *apistructs.DeploymentOrderCreateRequest) (*apistructs.DeploymentOrderCreateResponse, error) {
	// generate order id
	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	// parse release id
	releaseId, err := d.getReleaseIdFromReq(req)
	if err != nil {
		logrus.Errorf("failed to get release id, err: %v", err)
		return nil, err
	}

	// get release info
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: releaseId})
	if err != nil {
		logrus.Errorf("failed to get release %s, err: %v", releaseId, err)
		return nil, err
	}

	releaseData := releaseResp.GetData()

	// workspace
	if req.Workspace != "" {
		req.Workspace = strings.ToUpper(req.Workspace)
	} else {
		req.Workspace, err = d.getWorkspaceFromBranch(req.ProjectId, releaseData.GetLabels()[gitBranchLabel])
		if err != nil {
			logrus.Errorf("failed to get workspace, err: %v", err)
			return nil, err
		}
	}

	var (
		deployList [][]*pb.ApplicationReleaseSummary
		// application map: id -> name
		// id: for check permission
		// name: for return application name tips
		appsInfo = make(map[int64]string)
	)

	switch parseOrderType(releaseData.GetIsProjectRelease()) {
	case apistructs.TypeProjectRelease:
		if len(req.Modes) == 0 {
			if _, ok := releaseResp.Data.Modes[DefaultMode]; !ok {
				return nil, errors.Errorf("default mode does not exist, please select modes")
			}
			req.Modes = []string{DefaultMode}
		}
		for _, modeName := range req.Modes {
			if _, ok := releaseResp.Data.Modes[modeName]; !ok {
				return nil, errors.Errorf("mode %s does not exist in release %s", modeName, releaseId)
			}
		}
		for _, mode := range req.Modes {
			if _, ok := releaseData.GetModes()[mode]; !ok {
				return nil, errors.Errorf("mode %s does not exist in release modes list", mode)
			}
		}
		deployList, err = d.renderDeployListWithCrossProject(req.Modes, req.ProjectId, req.Operator, releaseData)
		if err != nil {
			return nil, errors.Errorf("failed to render deploy list with cross project, err: %v", err)
		}
		appsInfo = d.parseAppsInfoWithDeployList(deployList)
	case apistructs.TypeApplicationRelease:
		appsInfo[releaseData.GetApplicationID()] = releaseData.GetApplicationName()
	default:

	}

	// permission check
	if err := d.batchCheckExecutePermission(req.Operator, req.Workspace, appsInfo); err != nil {
		return nil, apierrors.ErrCreateDeploymentOrder.InternalError(err)
	}

	// compose deployment order
	order, deployListStr, err := d.composeDeploymentOrder(releaseData, req, deployList)
	if err != nil {
		logrus.Errorf("failed to compose deployment order, error: %v", err)
		return nil, err
	}

	// save order to db
	if err := d.db.UpdateDeploymentOrder(order); err != nil {
		logrus.Errorf("failed to update deployment order, err: %v", err)
		return nil, err
	}

	createResp := &apistructs.DeploymentOrderCreateResponse{
		Id:              order.ID,
		Name:            utils.ParseOrderName(order.ID),
		Type:            order.Type,
		ReleaseId:       order.ReleaseId,
		ProjectId:       order.ProjectId,
		ProjectName:     order.ProjectName,
		ApplicationId:   order.ApplicationId,
		ApplicationName: order.ApplicationName,
		Status:          utils.ParseDeploymentOrderStatus(nil),
		DeployList:      deployListStr,
	}

	if !req.AutoRun {
		return createResp, nil
	}

	order.CurrentBatch = FirstBatch
	executeDeployResp := make(map[string]*apistructs.DeploymentCreateResponseDTO)

	switch order.Type {
	case apistructs.TypeProjectRelease:
		executeDeployResp, err = d.deployProjectRelease(order, deployList, req.Source)
	case apistructs.TypeApplicationRelease:
		executeDeployResp, err = d.deployApplicationRelease(order, releaseData, req.Source, parseRuntimeNameFromBranch(req))
	}

	if err != nil {
		logrus.Errorf("failed to executeDeploy, err: %v", err)
		// order had been created, return error with order context
		return createResp, err
	}

	createResp.Deployments = executeDeployResp
	return createResp, nil
}

func (d *DeploymentOrder) Deploy(ctx context.Context, req *apistructs.DeploymentOrderDeployRequest) (*dbclient.DeploymentOrder, error) {
	order, err := d.db.GetDeploymentOrder(req.DeploymentOrderId)
	if err != nil {
		logrus.Errorf("failed to get deployment order, err: %v", err)
		return nil, err
	}

	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: order.ReleaseId})
	if err != nil {
		logrus.Errorf("failed to get release, err: %v", err)
		return nil, err
	}

	var (
		appsInfo   = make(map[int64]string)
		deployList [][]*pb.ApplicationReleaseSummary
	)

	switch order.Type {
	case apistructs.TypeProjectRelease:
		deployList, err = d.renderDeployListWithCrossProject(strings.Split(order.Modes, ","), order.ProjectId,
			req.Operator, releaseResp.GetData())
		if err != nil {
			return nil, errors.Errorf("failed to render deploy list with cross project, err: %v", err)
		}
		appsInfo = d.parseAppsInfoWithDeployList(deployList)
	case apistructs.TypeApplicationRelease:
		appsInfo[order.ApplicationId] = order.ApplicationName
	default:
		return nil, errors.Errorf("unsupported deployment order type %s", order.Type)
	}

	// permission check
	if err := d.batchCheckExecutePermission(req.Operator, order.Workspace, appsInfo); err != nil {
		logrus.Errorf("failed to check execute permission, err: %v", err)
		return nil, apierrors.ErrDeployDeploymentOrder.InternalError(err)
	}

	order.Operator = user.ID(req.Operator)

	// deploy interface will means execute from the first batch
	// if current batch is not zero, means it is a retry at current batch, deploy will continue from the current batch
	if order.CurrentBatch == 0 {
		order.CurrentBatch = FirstBatch
	}

	switch order.Type {
	case apistructs.TypeProjectRelease:
		_, err = d.deployProjectRelease(order, deployList, apistructs.SourceDeployCenter)
	case apistructs.TypeApplicationRelease:
		_, err = d.deployApplicationRelease(order, releaseResp.GetData(), apistructs.SourceDeployCenter, false)
	}

	if err != nil {
		logrus.Errorf("failed to execute deploy, order id: %s, err: %v", req.DeploymentOrderId, err)
		return nil, err
	}

	return order, nil
}

// ContinueDeployOrder deploy from queue compensation
func (d *DeploymentOrder) ContinueDeployOrder(orderId string) error {
	order, err := d.db.GetDeploymentOrder(orderId)
	if err != nil {
		logrus.Errorf("failed to get deployment order, err: %v", err)
		return err
	}

	ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := d.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: order.ReleaseId})
	if err != nil {
		logrus.Errorf("failed to get release, err: %v", err)
		return err
	}

	deployList, err := d.renderDeployListWithCrossProject(strings.Split(order.Modes, ","), order.ProjectId,
		order.Operator.String(), releaseResp.GetData())
	if err != nil {
		return errors.Errorf("failed to render deploy list with cross project, err: %v", err)
	}

	switch order.Type {
	case apistructs.TypeProjectRelease:
		_, err = d.deployProjectRelease(order, deployList, apistructs.SourceDeployCenter)
	case apistructs.TypeApplicationRelease:
		_, err = d.deployApplicationRelease(order, releaseResp.GetData(), apistructs.SourceDeployCenter, false)
	}

	if err != nil {
		logrus.Errorf("failed to execute deploy, order id: %s, err: %v", orderId, err)
		return err
	}

	return nil
}

func (d *DeploymentOrder) deployProjectRelease(order *dbclient.DeploymentOrder, deployList [][]*pb.ApplicationReleaseSummary,
	source string) (map[string]*apistructs.DeploymentCreateResponseDTO, error) {
	var runtimeSource = apistructs.RuntimeSource(source)

	if source != apistructs.SourceDeployPipeline {
		runtimeSource = release
	}
	if deployList == nil {
		return nil, errors.Errorf("project release deployment order deploy list is nil")
	}
	// compose runtime create requests with order current batch
	rtCreateReqs, err := d.composeRtReqWithDeployList(order, deployList, runtimeSource)
	if err != nil {
		logrus.Errorf("failed to compose runtime create request, err: %v", err)
		return nil, err
	}
	return d.deployWithRtCreateRequest(order, rtCreateReqs)
}

func (d *DeploymentOrder) deployApplicationRelease(order *dbclient.DeploymentOrder, releaseData *pb.ReleaseGetResponseData,
	source string, isRuntimeNameFromBranch bool) (map[string]*apistructs.DeploymentCreateResponseDTO, error) {
	var runtimeSource = apistructs.RuntimeSource(source)

	if source != apistructs.SourceDeployPipeline {
		runtimeSource = release
	}
	if releaseData == nil {
		return nil, errors.Errorf("application release deployment order deploy list is nil")
	}
	// compose runtime create requests with order current batch
	rtCreateReqs, err := d.composeRtReqWithRelease(order, releaseData, runtimeSource, isRuntimeNameFromBranch)
	if err != nil {
		logrus.Errorf("failed to compose runtime create request, err: %v", err)
		return nil, err
	}
	return d.deployWithRtCreateRequest(order, rtCreateReqs)
}

func (d *DeploymentOrder) deployWithRtCreateRequest(order *dbclient.DeploymentOrder,
	rtCreateReqs []*apistructs.RuntimeCreateRequest) (map[string]*apistructs.DeploymentCreateResponseDTO, error) {

	applicationsStatus := make(apistructs.DeploymentOrderStatusMap)
	// redeploy
	if order.CurrentBatch != FirstBatch && order.StatusDetail != "" {
		if err := json.Unmarshal([]byte(order.StatusDetail), &applicationsStatus); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to deployment order status (%s), err: %v",
				order.ID, err)
		}
	} else {
		order.StartedAt = time.Now()
	}

	// compose init status
	for _, rtCreateReq := range rtCreateReqs {
		applicationsStatus[rtCreateReq.Extra.ApplicationName] = apistructs.DeploymentOrderStatusItem{
			DeploymentStatus: apistructs.DeploymentStatusInit,
		}
	}

	orderStatus := apistructs.DeploymentStatusDeploying
	var failedReason string

	deployResponse := make(map[string]*apistructs.DeploymentCreateResponseDTO)
	// create runtimes
	for _, rtCreateReq := range rtCreateReqs {
		runtimeCreateResp, err := d.rt.Create(order.Operator, rtCreateReq)
		if err != nil {
			logrus.Errorf("failed to create runtime %s, cluster: %s, release id: %s, err: %v",
				rtCreateReq.Name, rtCreateReq.ClusterName, rtCreateReq.ReleaseID, err)
			applicationsStatus[rtCreateReq.Extra.ApplicationName] = apistructs.DeploymentOrderStatusItem{
				DeploymentStatus: apistructs.DeploymentStatusFailed,
			}
			orderStatus = apistructs.DeploymentStatusFailed
			failedReason = fmt.Sprintf("application: %s, failed reason: %v", rtCreateReq.Extra.ApplicationName, err)
			break
		}
		deployResponse[rtCreateReq.Extra.ApplicationName] = runtimeCreateResp
	}

	// marshal applications status
	jsonAppStatus, err := json.Marshal(applicationsStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal applications status, err: %v", err)
	}

	order.StatusDetail = string(jsonAppStatus)
	order.Status = string(orderStatus)

	if err := d.db.UpdateDeploymentOrder(order); err != nil {
		logrus.Errorf("failed to update deployment order, err: %v", err)
		return nil, err
	}

	if orderStatus == apistructs.DeploymentStatusFailed {
		return nil, fmt.Errorf("deploy failed, reason: %s", failedReason)
	}

	return deployResponse, nil
}

func (d *DeploymentOrder) composeDeploymentOrder(release *pb.ReleaseGetResponseData,
	req *apistructs.DeploymentOrderCreateRequest, deployList [][]*pb.ApplicationReleaseSummary) (*dbclient.DeploymentOrder, string, error) {
	var (
		orderId       = req.Id
		orderType     = parseOrderType(release.GetIsProjectRelease())
		workspace     = req.Workspace
		list          = make([][]string, len(deployList))
		deployListStr string
		err           error
	)

	order := &dbclient.DeploymentOrder{
		ID:        orderId,
		Type:      orderType,
		Workspace: workspace,
		Operator:  user.ID(req.Operator),
		ReleaseId: release.ReleaseID,
		BatchSize: uint64(len(deployList)),
	}

	switch orderType {
	case apistructs.TypeProjectRelease:
		if req.ProjectId == 0 {
			return nil, "", errors.Errorf("project id is empty")
		}

		modes := strings.Join(req.Modes, ",")
		for i, l := range deployList {
			for _, summary := range l {
				list[i] = append(list[i], summary.ReleaseID)
			}
		}
		deployListStr, err = marshalDeployList(list)
		if err != nil {
			return nil, "", errors.Errorf("failed to marshal deploy list for release %s, %v", release.ReleaseID, err)
		}

		projectInfo, err := d.bdl.GetProject(req.ProjectId)
		if err != nil {
			return nil, "", errors.Errorf("failed to get project %d, %v", req.ProjectId, err)
		}

		order.Modes = modes
		order.DeployList = deployListStr
		order.ProjectId = projectInfo.ID
		order.ProjectName = projectInfo.Name
	case apistructs.TypeApplicationRelease:
		order.BatchSize = 1
		order.ProjectId, order.ProjectName = uint64(release.ProjectID), release.ProjectName
		order.ApplicationId, order.ApplicationName = release.ApplicationID, release.ApplicationName
	default:
	}

	params, err := d.fetchApplicationsParams(release, deployList, workspace)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch deployment params, err: %v", err)
	}

	paramsJson, err := json.Marshal(params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal params, err: %v", err)
	}

	order.Params = string(paramsJson)

	return order, deployListStr, nil
}

func (d *DeploymentOrder) fetchApplicationsParams(r *pb.ReleaseGetResponseData, deployList [][]*pb.ApplicationReleaseSummary,
	workspace string) (map[string]*apistructs.DeploymentOrderParam, error) {
	ret := make(map[string]*apistructs.DeploymentOrderParam)

	if r.IsProjectRelease {
		for i := 0; i < len(deployList); i++ {
			if deployList[i] == nil {
				continue
			}
			for _, ar := range deployList[i] {
				params, err := d.fetchDeploymentParams(ar.ApplicationID, workspace)
				if err != nil {
					return nil, err
				}
				ret[ar.ApplicationName] = params
			}
		}
	} else {
		params, err := d.fetchDeploymentParams(r.ApplicationID, workspace)
		if err != nil {
			return nil, err
		}
		ret[r.ApplicationName] = params
	}

	return ret, nil
}

func (d *DeploymentOrder) FetchDeploymentConfigDetail(namespace string) ([]apistructs.EnvConfig, []apistructs.EnvConfig, error) {
	envConfigs, err := d.envConfig.GetDeployConfigs(namespace)
	if err != nil {
		return nil, nil, err
	}
	envs := make([]apistructs.EnvConfig, 0)
	files := make([]apistructs.EnvConfig, 0)
	for _, c := range envConfigs {
		if c.ConfigType == "FILE" {
			files = append(files, c)
		} else {
			envs = append(envs, c)
		}
	}

	return envs, files, nil
}
func (d *DeploymentOrder) fetchDeploymentParams(applicationId int64, workspace string) (*apistructs.DeploymentOrderParam, error) {
	app, err := d.bdl.GetApp(uint64(applicationId))
	if err != nil {
		return nil, fmt.Errorf("failed to get application %d, err: %v", applicationId, err)
	}

	var cfgNamespace string
	for _, ws := range app.Workspaces {
		if strings.ToUpper(ws.Workspace) != strings.ToUpper(workspace) {
			continue
		}
		cfgNamespace = ws.ConfigNamespace
	}

	if cfgNamespace == "" {
		return nil, fmt.Errorf("failed to get config namespace, application %d, workspace %s", applicationId, workspace)
	}

	deployConfig, fileConfig, err := d.FetchDeploymentConfigDetail(cfgNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment config, err: %v", err)
	}

	params := make(apistructs.DeploymentOrderParam, 0)

	for _, c := range deployConfig {
		params = append(params, &apistructs.DeploymentOrderParamData{
			Key:     c.Key,
			Value:   c.Value,
			Type:    "ENV",
			Encrypt: c.Encrypt,
			Comment: c.Comment,
		})
	}

	for _, c := range fileConfig {
		params = append(params, &apistructs.DeploymentOrderParamData{
			Key:     c.Key,
			Value:   c.Value,
			Type:    "FILE",
			Encrypt: c.Encrypt,
			Comment: c.Comment,
		})
	}

	return &params, nil
}

func (d *DeploymentOrder) composeRtReqWithDeployList(order *dbclient.DeploymentOrder,
	deployList [][]*pb.ApplicationReleaseSummary, source apistructs.RuntimeSource) ([]*apistructs.RuntimeCreateRequest, error) {
	if order == nil || deployList == nil {
		return nil, fmt.Errorf("deployment order or deploy list is nil")
	}

	if order.CurrentBatch == 0 {
		return nil, fmt.Errorf("current batch is 0")
	}

	if len(deployList) == 0 {
		return nil, errors.Errorf("invalid deploy list")
	}

	projectInfo, err := d.bdl.GetProject(order.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project info, id: %d, err: %v", order.ProjectId, err)
	}

	// get cluster name with workspace
	clusterName, ok := projectInfo.ClusterConfig[order.Workspace]
	if !ok {
		return nil, fmt.Errorf("cluster not found at workspace: %s", order.Workspace)
	}

	extraParamsStr := ""
	if order.Modes != "" {
		extraParams := map[string]string{
			DeployModeEnvName: order.Modes,
		}
		data, err := json.Marshal(extraParams)
		if err != nil {
			return nil, errors.Errorf("failed to marshal extraParams, %v", err)
		}
		extraParamsStr = string(data)
	}

	// parse params
	var orderParams map[string]*apistructs.DeploymentOrderParam
	if err := json.Unmarshal([]byte(order.Params), &orderParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params, err: %v", err)
	}

	ret := make([]*apistructs.RuntimeCreateRequest, 0, len(deployList[order.CurrentBatch-1]))

	for _, r := range deployList[order.CurrentBatch-1] {
		rtCreateReq := &apistructs.RuntimeCreateRequest{
			Name:              r.GetApplicationName(),
			DeploymentOrderId: order.ID,
			ReleaseVersion:    r.GetVersion(),
			ReleaseID:         r.GetReleaseID(),
			Source:            source,
			Operator:          order.Operator.String(),
			ClusterName:       clusterName,
			Extra: apistructs.RuntimeCreateRequestExtra{
				OrgID:           projectInfo.OrgID,
				ProjectID:       order.ProjectId,
				ApplicationName: r.GetApplicationName(),
				ApplicationID:   uint64(r.GetApplicationID()),
				DeployType:      release,
				Workspace:       order.Workspace,
				BuildID:         0, // Deprecated
			},
			SkipPushByOrch: false,
			ExtraParams:    extraParamsStr,
		}

		paramJson, err := json.Marshal(orderParams[r.GetApplicationName()])
		if err != nil {
			return nil, err
		}
		rtCreateReq.Param = string(paramJson)

		ret = append(ret, rtCreateReq)
	}
	return ret, nil
}

func (d *DeploymentOrder) composeRtReqWithRelease(order *dbclient.DeploymentOrder, r *pb.ReleaseGetResponseData,
	source apistructs.RuntimeSource, isRuntimeNameFromBranch bool) ([]*apistructs.RuntimeCreateRequest, error) {
	projectInfo, err := d.bdl.GetProject(order.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project info, id: %d, err: %v", order.ProjectId, err)
	}

	// get cluster name with workspace
	clusterName, ok := projectInfo.ClusterConfig[order.Workspace]
	if !ok {
		return nil, fmt.Errorf("cluster not found at workspace: %s", order.Workspace)
	}

	rtCreateReq := &apistructs.RuntimeCreateRequest{
		Name:              order.ApplicationName,
		DeploymentOrderId: order.ID,
		ReleaseVersion:    r.Version,
		ReleaseID:         r.ReleaseID,
		Source:            source,
		Operator:          order.Operator.String(),
		ClusterName:       clusterName,
		Extra: apistructs.RuntimeCreateRequestExtra{
			OrgID:           projectInfo.OrgID,
			ProjectID:       projectInfo.ID,
			ApplicationID:   uint64(r.ApplicationID),
			ApplicationName: r.ApplicationName,
			DeployType:      release,
			Workspace:       order.Workspace,
			BuildID:         0, // Deprecated
		},
		SkipPushByOrch: false,
	}

	// parse params
	var orderParams map[string]*apistructs.DeploymentOrderParam
	if err := json.Unmarshal([]byte(order.Params), &orderParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params, err: %v", err)
	}

	paramJson, err := json.Marshal(orderParams[r.ApplicationName])
	if err != nil {
		return nil, err
	}
	rtCreateReq.Param = string(paramJson)

	if isRuntimeNameFromBranch {
		branch, ok := r.Labels[gitBranchLabel]
		if !ok {
			return nil, fmt.Errorf("failed to get release branch in release %s", r.ReleaseID)
		}
		rtCreateReq.Name = branch
		rtCreateReq.Extra.DeployType = ""
	}
	return []*apistructs.RuntimeCreateRequest{rtCreateReq}, nil
}

func (d *DeploymentOrder) parseAppsInfoWithDeployList(deployList [][]*pb.ApplicationReleaseSummary) map[int64]string {
	ret := make(map[int64]string)
	for i := 0; i < len(deployList); i++ {
		if deployList[i] == nil {
			continue
		}
		for _, r := range deployList[i] {
			ret[r.GetApplicationID()] = r.GetApplicationName()
		}
	}

	return ret
}

func (d *DeploymentOrder) getReleaseIdFromReq(req *apistructs.DeploymentOrderCreateRequest) (string, error) {
	if req.ReleaseId != "" {
		return req.ReleaseId, nil
	}

	if req.ReleaseName == "" {
		return "", fmt.Errorf("please specified release name or release id")
	}

	switch req.Type {
	case apistructs.TypeProjectRelease:
		if req.ProjectId == 0 {
			return "", fmt.Errorf("deploy project release must provide effective project id")
		}
		r, err := d.db.GetProjectReleaseByVersion(req.ReleaseName, req.ProjectId)
		if err != nil {
			return "", err
		}
		return r.ReleaseId, nil
	case apistructs.TypeApplicationRelease:
		if req.ReleaseName == "" {
			return "", fmt.Errorf("deploy application relealse must provide effective application name")
		}
		r, err := d.db.GetApplicationReleaseByVersion(req.ReleaseName, req.ApplicationName)
		if err != nil {
			return "", err
		}
		return r.ReleaseId, nil
	default:
		return "", fmt.Errorf("deploy type %s is not support", req.Type)
	}
}

func (d *DeploymentOrder) getWorkspaceFromBranch(projectId uint64, branch string) (string, error) {
	rules, err := d.bdl.GetProjectBranchRules(projectId)
	if err != nil {
		return "", err
	}
	workspace, err := diceworkspace.GetByGitReference(branch, rules)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace by branch %s: %v", branch, err)
	}
	return workspace.String(), nil
}

// renderDeployListWithCrossProject: render deploy list with cross project
// - deploy target project is equal to project which release belong -> non-cross project
// - deploy target project is not equal to project which release belong -> cross project
func (d *DeploymentOrder) renderDeployListWithCrossProject(selectedModes []string, projectId uint64, userId string,
	releaseResp *pb.ReleaseGetResponseData) ([][]*pb.ApplicationReleaseSummary, error) {
	deployList := renderDeployList(selectedModes, releaseResp.Modes)
	// non-cross project
	if int64(projectId) == releaseResp.ProjectID {
		return deployList, nil
	}

	appNames := make([]string, 0)

	for _, subList := range deployList {
		for _, deploy := range subList {
			appNames = append(appNames, deploy.ApplicationName)
		}
	}

	resp, err := d.bdl.GetAppIDByNames(projectId, userId, appNames)
	if err != nil {
		return nil, err
	}

	// cross project, overwrite application id.
	for _, subList := range deployList {
		for _, deploy := range subList {
			appId, ok := resp.AppNameToID[deploy.ApplicationName]
			if !ok {
				return nil, fmt.Errorf("failed to find application %s in project %d", deploy.ApplicationName, projectId)
			}
			deploy.ApplicationID = appId
		}
	}

	return deployList, nil
}

func parseRuntimeNameFromBranch(r *apistructs.DeploymentOrderCreateRequest) bool {
	return r.Source == apistructs.SourceDeployPipeline && r.ReleaseId != "" && !r.DeployWithoutBranch
}

func parseOrderType(isProjectRelease bool) string {
	if isProjectRelease {
		return apistructs.TypeProjectRelease
	}
	return apistructs.TypeApplicationRelease
}

func covertParamsType(param *apistructs.DeploymentOrderParam) *apistructs.DeploymentOrderParam {
	if param == nil {
		return param
	}
	for _, data := range *param {
		data.Type = convertConfigType(data.Type)
	}
	return param
}

func renderDeployList(selectedModes []string, modes map[string]*pb.ModeSummary) [][]*pb.ApplicationReleaseSummary {
	var (
		deployList [][]*pb.ApplicationReleaseSummary
		dfs        func(string) int
		visited    = make(map[string]int)
	)

	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	dfs = func(u string) int {
		height := -1
		for _, v := range modes[u].DependOn {
			if h, ok := visited[v]; ok {
				height = max(height, h)
				continue
			}
			height = max(height, dfs(v))
		}
		for i := range modes[u].ApplicationReleaseList {
			height++
			if len(deployList) <= height {
				deployList = append(deployList, []*pb.ApplicationReleaseSummary{})
			}
			deployList[height] = append(deployList[height], modes[u].ApplicationReleaseList[i].List...)
		}
		visited[u] = height
		return height
	}

	for _, modeName := range selectedModes {
		if _, ok := visited[modeName]; !ok {
			dfs(modeName)
		}
	}
	return deployList
}

func marshalDeployList(deployList [][]string) (string, error) {
	data, err := json.Marshal(deployList)
	return string(data), err
}

func unmarshalDeployList(deployListData string) ([][]string, error) {
	if deployListData == "" {
		return nil, nil
	}
	var deployList [][]string
	err := json.Unmarshal([]byte(deployListData), &deployList)
	return deployList, err
}
