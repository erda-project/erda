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

// Package runtime 应用实例相关操作
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
	"math"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/pkg/transport"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/user"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Runtime 应用实例对象封装
type Runtime struct {
	db               *dbclient.DBClient
	evMgr            *events.EventManager
	bdl              *bundle.Bundle
	addon            *addon.Addon
	releaseSvc       pb.ReleaseServiceServer
	serviceGroupImpl servicegroup.ServiceGroup
	clusterinfoImpl  clusterinfo.ClusterInfo
	clusterSvc       clusterpb.ClusterServiceServer
	pipelineSvc      pipelinepb.PipelineServiceServer
	org              org.ClientInterface
}

// Option 应用实例对象配置选项
type Option func(*Runtime)

// New 新建应用实例 service
func New(options ...Option) *Runtime {
	r := &Runtime{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(r *Runtime) {
		r.db = db
	}
}

// WithEventManager 配置 EventManager
func WithEventManager(evMgr *events.EventManager) Option {
	return func(r *Runtime) {
		r.evMgr = evMgr
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(r *Runtime) {
		r.bdl = bdl
	}
}

// WithAddon 配置 addon service
func WithAddon(a *addon.Addon) Option {
	return func(r *Runtime) {
		r.addon = a
	}
}

// WithReleaseSvc 配置 dicehub release service
func WithReleaseSvc(svc pb.ReleaseServiceServer) Option {
	return func(r *Runtime) {
		r.releaseSvc = svc
	}
}

// WithServiceGroup 配置 serviceGroupImpl
func WithServiceGroup(serviceGroupImpl servicegroup.ServiceGroup) Option {
	return func(r *Runtime) {
		r.serviceGroupImpl = serviceGroupImpl
	}
}

func WithClusterInfo(clusterinfo clusterinfo.ClusterInfo) Option {
	return func(r *Runtime) {
		r.clusterinfoImpl = clusterinfo
	}
}

func WithClusterSvc(clusterSvc clusterpb.ClusterServiceServer) Option {
	return func(r *Runtime) {
		r.clusterSvc = clusterSvc
	}
}

func WithOrg(org org.ClientInterface) Option {
	return func(e *Runtime) {
		e.org = org
	}
}

func WithPipelineSvc(svc pipelinepb.PipelineServiceServer) Option {
	return func(r *Runtime) {
		r.pipelineSvc = svc
	}
}

func (r *Runtime) CreateByReleaseIDPipeline(ctx context.Context, orgid uint64, operator user.ID, releaseReq *apistructs.RuntimeReleaseCreateRequest) (*apistructs.RuntimeDeployDTO, error) {
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := r.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: releaseReq.ReleaseID})
	if err != nil {
		return nil, err
	}
	app, err := r.bdl.GetApp(releaseReq.ApplicationID)
	if err != nil {
		return nil, err
	}
	workspaces := strutil.Split(releaseReq.Workspace, ",", true)
	commitid := releaseResp.Data.Labels["gitCommitId"]
	branch := releaseResp.Data.Labels["gitBranch"]

	// check if there is a runtime already being created by release
	pipelines, err := utils.FindCRBRRunningPipeline(releaseReq.ApplicationID, workspaces[0],
		fmt.Sprintf("dice-deploy-release-%s", branch), r.bdl)
	if err != nil {
		return nil, err
	}
	if len(pipelines) != 0 {
		return nil,
			errors.Errorf("There is already a runtime created by releaseID %s, please do not repeat deployment", releaseReq.ReleaseID)
	}

	yml := utils.GenCreateByReleasePipelineYaml(releaseReq.ReleaseID, workspaces)
	b, err := yaml.Marshal(yml)
	if err != nil {
		errstr := fmt.Sprintf("failed to marshal pipelineyml: %v", err)
		logrus.Errorf(errstr)
		return nil, err
	}
	detail := apistructs.CommitDetail{
		CommitID: "",
		Repo:     app.GitRepo,
		RepoAbbr: app.GitRepoAbbrev,
		Author:   "",
		Email:    "",
		Time:     nil,
		Comment:  "",
	}
	if commitid != "" {
		commit, err := r.bdl.GetGittarCommit(app.GitRepoAbbrev, commitid, string(operator))
		if err != nil {
			return nil, err
		}
		detail = apistructs.CommitDetail{
			CommitID: commitid,
			Repo:     app.GitRepo,
			RepoAbbr: app.GitRepoAbbrev,
			Author:   commit.Committer.Name,
			Email:    commit.Committer.Email,
			Time:     &commit.Committer.When,
			Comment:  commit.CommitMessage,
		}
	}
	commitdetail, err := json.Marshal(detail)
	if err != nil {
		return nil, err
	}
	dto, err := r.pipelineSvc.PipelineCreateV2(apis.WithInternalClientContext(context.Background(), discover.Orchestrator()), &pipelinepb.PipelineCreateRequestV2{
		UserID:      operator.String(),
		PipelineYml: string(b),
		Labels: map[string]string{
			apistructs.LabelBranch:        releaseResp.Data.ReleaseName,
			apistructs.LabelOrgID:         strconv.FormatUint(orgid, 10),
			apistructs.LabelProjectID:     strconv.FormatUint(releaseReq.ProjectID, 10),
			apistructs.LabelAppID:         strconv.FormatUint(releaseReq.ApplicationID, 10),
			apistructs.LabelDiceWorkspace: releaseReq.Workspace,
			apistructs.LabelCommitDetail:  string(commitdetail),
			apistructs.LabelAppName:       app.Name,
			apistructs.LabelProjectName:   app.ProjectName,
		},
		PipelineYmlName: fmt.Sprintf("dice-deploy-release-%s-%d-%s", releaseReq.Workspace,
			releaseReq.ApplicationID, branch),
		ClusterName:    releaseResp.Data.ClusterName,
		PipelineSource: apistructs.PipelineSourceDice.String(),
		AutoRunAtOnce:  true,
	})
	if err != nil {
		return nil, err
	}
	return convertRuntimeDeployDto(app, releaseResp.Data, dto.Data)
}

func (r *Runtime) RedeployPipeline(ctx context.Context, operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDeployDTO, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, err
	}
	yml := utils.GenRedeployPipelineYaml(runtimeID)
	app, err := r.bdl.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, err
	}
	deployment, err := r.db.FindLastDeployment(runtimeID)
	if err != nil {
		return nil, err
	}
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := r.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: deployment.ReleaseId})
	if err != nil {
		return nil, err
	}
	commitid := releaseResp.Data.Labels["gitCommitId"]
	detail := apistructs.CommitDetail{
		CommitID: "",
		Repo:     app.GitRepo,
		RepoAbbr: app.GitRepoAbbrev,
		Author:   "",
		Email:    "",
		Time:     nil,
		Comment:  "",
	}
	if commitid != "" {
		commit, err := r.bdl.GetGittarCommit(app.GitRepoAbbrev, commitid, string(operator))
		if err != nil {
			return nil, err
		}
		detail = apistructs.CommitDetail{
			CommitID: commitid,
			Repo:     app.GitRepo,
			RepoAbbr: app.GitRepoAbbrev,
			Author:   commit.Committer.Name,
			Email:    commit.Committer.Email,
			Time:     &commit.Committer.When,
			Comment:  commit.CommitMessage,
		}
	}
	commitdetail, err := json.Marshal(detail)
	if err != nil {
		return nil, err
	}
	b, err := yaml.Marshal(yml)
	if err != nil {
		errstr := fmt.Sprintf("failed to marshal pipelineyml: %v", err)
		logrus.Errorf(errstr)
		return nil, err
	}
	if err := r.setClusterName(runtime); err != nil {
		logrus.Errorf("get cluster info failed, cluster name: %s, error: %v", runtime.ClusterName, err)
	}
	dto, err := r.pipelineSvc.PipelineCreateV2(apis.WithInternalClientContext(context.Background(), discover.Orchestrator()), &pipelinepb.PipelineCreateRequestV2{
		UserID:      operator.String(),
		PipelineYml: string(b),
		Labels: map[string]string{
			apistructs.LabelBranch:        runtime.Name,
			apistructs.LabelOrgID:         strconv.FormatUint(orgID, 10),
			apistructs.LabelProjectID:     strconv.FormatUint(runtime.ProjectID, 10),
			apistructs.LabelAppID:         strconv.FormatUint(runtime.ApplicationID, 10),
			apistructs.LabelDiceWorkspace: runtime.Workspace,
			apistructs.LabelCommitDetail:  string(commitdetail),
			apistructs.LabelAppName:       app.Name,
			apistructs.LabelProjectName:   app.ProjectName,
		},
		PipelineYmlName: getRedeployPipelineYmlName(*runtime),
		ClusterName:     runtime.ClusterName,
		PipelineSource:  apistructs.PipelineSourceDice.String(),
		AutoRunAtOnce:   true,
	})
	if err != nil {
		return nil, err
	}

	return convertRuntimeDeployDto(app, releaseResp.Data, dto.Data)
}

func (r *Runtime) setClusterName(rt *dbclient.Runtime) error {
	clusterInfo, err := r.clusterinfoImpl.Info(rt.ClusterName)
	if err != nil {
		logrus.Errorf("get cluster info failed, cluster name: %s, error: %v", rt.ClusterName, err)
		return err
	}
	jobCluster := clusterInfo.Get(apistructs.JOB_CLUSTER)
	if jobCluster != "" {
		rt.ClusterName = jobCluster
	}
	return nil
}

// Delete 标记应用实例删除
func (r *Runtime) Delete(operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDTO, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrDeleteRuntime.InternalError(err)
	}
	// TODO: do not query app
	app, err := r.bdl.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, err
	}
	perm, err := r.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   operator.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  app.ID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.DeleteAction,
	})
	if err != nil {
		return nil, apierrors.ErrDeleteRuntime.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrDeleteRuntime.AccessDenied()
	}
	if runtime.LegacyStatus == dbclient.LegacyStatusDeleting {
		// already marked
		return dbclient.ConvertRuntimeDTO(runtime, app), nil
	}
	if runtime.FileToken != "" {
		if _, err = r.bdl.InvalidateOAuth2Token(apistructs.OAuth2TokenInvalidateRequest{AccessToken: runtime.FileToken}); err != nil {
			logrus.Errorf("failed to invalidate openapi oauth2 token of runtime %v, token: %v, err: %v",
				runtime.ID, runtime.FileToken, err)
		}
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
	return dbclient.ConvertRuntimeDTO(runtime, app), nil
}

// Destroy 摧毁应用实例
func (r *Runtime) Destroy(runtimeID uint64) error {
	// 调用hepa完成runtime删除
	logrus.Infof("delete runtime and request hepa, runtimeId is %d", runtimeID)
	if err := r.bdl.DeleteRuntimeService(runtimeID); err != nil {
		logrus.Errorf("failed to request hepa, (%v)", err)
	}
	logrus.Debugf("do delete runtime %d", runtimeID)
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return err
	}
	app, err := r.bdl.GetApp(runtime.ApplicationID)
	if err != nil {
		return err
	}
	if err := r.db.DeleteDomainsByRuntimeId(runtimeID); err != nil {
		return err
	}
	if err := r.addon.RuntimeAddonRemove(strconv.FormatUint(runtimeID, 10), runtime.Workspace, runtime.Creator, app.ProjectID); err != nil {
		// TODO: need addon-platform to fix the 405 error
		//return err
		logrus.Infof("failed to delete addons when delete runtimes: %v", err)
	}
	// TODO: delete alert rules
	uniqueID := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}
	if err := r.db.ResetPreDice(uniqueID); err != nil {
		return err
	}
	// clear reference
	r.MarkOutdatedForDelete(runtimeID)
	if runtime.Source == apistructs.ABILITY {
		// TODO: delete ability info
	}
	if runtime.ScheduleName.Name != "" {
		// delete scheduler group
		var req apistructs.ServiceGroupDeleteRequest
		ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.InternalHeader: "cmp"}))
		resp, err := r.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{IdOrName: runtime.ClusterName})
		if err != nil {
			logrus.Errorf("get cluster info failed, cluster name: %s, error: %v", runtime.ClusterName, err)
			return err
		}
		cInfo := resp.Data
		if cInfo != nil && cInfo.OpsConfig != nil && cInfo.OpsConfig.Status == apistructs.ClusterStatusOffline {
			req.Force = true
		}
		req.Namespace = runtime.ScheduleName.Namespace
		req.Name = runtime.ScheduleName.Name
		forceDelete := req.Force
		delHPAObjects, delVPAObjects, err := r.AppliedScaledObjects(uniqueID)
		if err != nil {
			logrus.Warnf("[alert] failed delete group, error in get runtime hpa and vpa rules: %v, (%v)",
				runtime.ScheduleName, err)
		}
		extra := make(map[string]string)
		for svcName, ruleJson := range delHPAObjects {
			extra[pstypes.ErdaHPAPrefix+svcName] = ruleJson
		}
		for svcName, ruleJson := range delVPAObjects {
			extra[pstypes.ErdaVPAPrefix+svcName] = ruleJson
		}
		if err := r.serviceGroupImpl.Delete(req.Namespace, req.Name, forceDelete, extra); err != nil && err != servicegroup.DeleteNotFound {
			// TODO: we should return err if delete failed (even if err is group not exist?)
			logrus.Errorf("[alert] failed delete group in scheduler: %v, (%v)",
				runtime.ScheduleName, err)
			return err
		}

		// delete runtime hpa/vpa rule
		if err := r.deleteRuntimePA(uniqueID, runtimeID); err != nil {
			logrus.Errorf("[alert] failed delete group, error in delete runtime hpa rules: %v, (%v)",
				runtime.ScheduleName, err)
		}
	}
	// TODO: delete services & instances table

	// do delete runtime table
	if err := r.db.DeleteRuntime(runtimeID); err != nil {
		return err
	}
	event := events.RuntimeEvent{
		EventName: events.RuntimeDeleted,
		Runtime:   dbclient.ConvertRuntimeDTO(runtime, app),
		// TODO: no activity yet, so operator useless currently
		//Operator:  operator,
	}
	r.evMgr.EmitEvent(&event)
	return nil
}

// GetSpec 查询应用实例规格
func (r *Runtime) GetSpec(userID user.ID, orgID uint64, runtimeID uint64) (*apistructs.ServiceGroup, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InternalError(err)
	}
	perm, err := r.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  runtime.OrgID,
		Resource: "runtime-config",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrGetRuntime.AccessDenied()
	}
	return r.serviceGroupImpl.InspectServiceGroupWithTimeout(runtime.ScheduleName.Namespace, runtime.ScheduleName.Name)
}

func (r *Runtime) KillPod(runtimeID uint64, podname string) error {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return err
	}

	if runtime.ScheduleName.Namespace == "" || runtime.ScheduleName.Name == "" || podname == "" {
		return errors.New("empty namespace or name or podname")
	}
	return r.serviceGroupImpl.KillPod(context.Background(), runtime.ScheduleName.Namespace, runtime.ScheduleName.Name, podname)
}

func (r *Runtime) GetRuntimeServiceCurrentPods(runtimeID uint64, serviceName string) (*apistructs.ServiceGroup, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, err
	}

	if runtime.ScheduleName.Namespace == "" || runtime.ScheduleName.Name == "" || serviceName == "" {
		return nil, errors.New("empty namespace or name or serviceName")
	}

	return r.serviceGroupImpl.InspectRuntimeServicePods(runtime.ScheduleName.Namespace, runtime.ScheduleName.Name, serviceName, fmt.Sprintf("%d", runtimeID))
}

// FullGC 定时全量 GC 过期的部署单
func (r *Runtime) FullGC() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logrus.Errorf("[alert] failed to fullGC, panic: %v", err)
		}
	}()

	rollbackCfg, err := r.getRollbackConfig()
	if err != nil {
		logrus.Errorf("[alert] failed to get all rollback config: %v", err)
		return
	}

	bulk := 100
	lastRuntimeID := uint64(0)
	for {
		runtimes, err := r.db.FindRuntimesNewerThan(lastRuntimeID, bulk)
		if err != nil {
			logrus.Errorf("[alert] failed to find runtimes after: %v, (%v)", lastRuntimeID, err)
			break
		}
		for i := range runtimes {
			keep, ok := rollbackCfg[runtimes[i].ProjectID][strings.ToUpper(runtimes[i].Workspace)]
			if !ok || keep <= 0 || keep > 100 {
				keep = 5
			}
			r.fullGCForSingleRuntime(runtimes[i].ID, keep)
		}
		if len(runtimes) < bulk {
			// ended
			break
		}
		lastRuntimeID = runtimes[len(runtimes)-1].ID
	}
}

// getRollbackConfig return the number of rollback record for each project and workspace
// key1: project_id, key2: workspace, value: the limit of rollback record
func (r *Runtime) getRollbackConfig() (map[uint64]map[string]int, error) {
	result := make(map[uint64]map[string]int, 0)
	// TODO: use cache to get project info
	projects, err := r.bdl.GetAllProjects()
	if err != nil {
		return nil, err
	}
	for _, prj := range projects {
		result[prj.ID] = prj.RollbackConfig
	}
	return result, nil
}

func (r *Runtime) fullGCForSingleRuntime(runtimeID uint64, keep int) {
	top, err := r.db.FindTopDeployments(runtimeID, keep)
	if err != nil {
		logrus.Errorf("[alert] failed to find top %d deployments for gc, (%v)", keep, err)
	}
	if len(top) < keep {
		// all of deployments should keep
		return
	}
	oldestID := top[len(top)-1].ID
	var hasSuccess bool
	for _, deployment := range top {
		if deployment.Status == apistructs.DeploymentStatusOK {
			hasSuccess = true
			break
		}
	}
	if deployments, err := r.db.FindNotOutdatedOlderThan(runtimeID, oldestID); err != nil {
		logrus.Errorf("[alert] failed to set outdated, runtimeID: %v, maxID: %v, (%v)",
			runtimeID, oldestID, err)
	} else {
		for i := range deployments {
			if !hasSuccess && deployments[i].Status == apistructs.DeploymentStatusOK {
				hasSuccess = true
				continue
			}
			r.markOutdated(&deployments[i])
		}
	}
}

// Deprecated
// ReferCluster 查看 runtime & addon 是否有使用集群
func (r *Runtime) ReferCluster(clusterName string, orgID uint64) bool {
	runtimes, err := r.db.ListRuntimeByOrgCluster(clusterName, orgID)
	if err != nil {
		logrus.Warnf("failed to list runtime, %v", err)
		return true
	}
	if len(runtimes) > 0 {
		return true
	}

	routingInstances, err := r.db.ListRoutingInstanceByOrgCluster(clusterName, orgID)
	if err != nil {
		logrus.Warnf("failed to list addon, %v", err)
		return true
	}
	if len(routingInstances) > 0 {
		return true
	}

	return false
}

// MarkOutdatedForDelete 将删除的应用实例，他们的所有部署单，标记为废弃
func (r *Runtime) MarkOutdatedForDelete(runtimeID uint64) {
	deployments, err := r.db.FindNotOutdatedOlderThan(runtimeID, math.MaxUint32)
	if err != nil {
		logrus.Errorf("[alert] failed to query all not outdated deployment before delete runtime: %v, (%v)",
			runtimeID, err)
	}
	for i := range deployments {
		r.markOutdated(&deployments[i])
	}
}

func (r *Runtime) markOutdated(deployment *dbclient.Deployment) {
	if deployment.Outdated {
		// already outdated
		return
	}
	deployment.Outdated = true
	if err := r.db.UpdateDeployment(deployment); err != nil {
		logrus.Errorf("[alert] failed to set deployment: %v outdated, (%v)", deployment.ID, err)
		return
	}
	if len(deployment.ReleaseId) > 0 {
		ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.InternalHeader: "true"}))
		if _, err := r.releaseSvc.UpdateReleaseReference(ctx, &pb.ReleaseReferenceUpdateRequest{
			ReleaseID: deployment.ReleaseId,
			Increase:  false,
		}); err != nil {
			logrus.Errorf("[alert] failed to decrease reference of release: %s, (%v)",
				deployment.ReleaseId, err)
		}
	}
}

// OrgJobLogs 数据中心--->任务列表 日志接口
func (r *Runtime) OrgJobLogs(ctx context.Context, userID user.ID, orgID uint64, orgName string, jobID, clusterName string, paramValues url.Values) (*apistructs.DashboardSpotLogData, error) {
	if clusterName == "" {
		logrus.Errorf("job instance infos without cluster, jobID is: %s", jobID)
		return nil, apierrors.ErrOrgLog.AccessDenied()
	}
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "cmp"}))
	resp, err := r.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{IdOrName: clusterName})
	if err != nil {
		return nil, err
	}
	clusterInfo := resp.Data
	if clusterInfo == nil {
		logrus.Errorf("can not find cluster info, clusterName: %s", clusterName)
		return nil, apierrors.ErrOrgLog.AccessDenied()
	}
	orgID = uint64(clusterInfo.OrgID)
	if err := r.checkOrgScopePermission(userID, orgID); err != nil {
		return nil, err
	}
	return r.requestMonitorLog(jobID, orgName, paramValues, apistructs.DashboardSpotLogSourceJob)
}

// checkRuntimeScopePermission 检测runtime级别的权限
func (r *Runtime) checkRuntimeScopePermission(userID user.ID, runtimeID uint64) error {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return err
	}
	perm, err := r.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return err
	}
	if !perm.Access {
		return apierrors.ErrGetRuntime.AccessDenied()
	}

	return nil
}

// checkOrgScopePermission 检测org级别的权限
func (r *Runtime) checkOrgScopePermission(userID user.ID, orgID uint64) error {
	perm, err := r.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: "monitor_org_alert",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return err
	}
	if !perm.Access {
		return apierrors.ErrGetRuntime.AccessDenied()
	}

	return nil
}

// requestMonitorLog 调用bundle monitor log接口获取数据
func (r *Runtime) requestMonitorLog(requestID string, orgName string, paramValues url.Values, source apistructs.DashboardSpotLogSource) (*apistructs.DashboardSpotLogData, error) {
	// 获取日志
	var logReq apistructs.DashboardSpotLogRequest
	if err := queryStringDecoder.Decode(&logReq, paramValues); err != nil {
		return nil, err
	}
	logReq.ID = requestID
	logReq.Source = source

	logResult, err := r.bdl.GetLog(orgName, logReq)
	if err != nil {
		return nil, err
	}
	return logResult, nil
}

var queryStringDecoder *schema.Decoder

func init() {
	queryStringDecoder = schema.NewDecoder()
	queryStringDecoder.IgnoreUnknownKeys(true)
}

func getRedeployPipelineYmlName(runtime dbclient.Runtime) string {
	return fmt.Sprintf("%d/%s/%s/pipeline.yml", runtime.ApplicationID, runtime.Workspace, runtime.Name)
}

func convertRuntimeDeployDto(app *apistructs.ApplicationDTO, release *pb.ReleaseGetResponseData, dto *basepb.PipelineDTO) (*apistructs.RuntimeDeployDTO, error) {
	names, err := getServicesNames(release.Diceyml)
	if err != nil {
		return nil, err
	}
	return &apistructs.RuntimeDeployDTO{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		ProjectID:       app.ProjectID,
		ProjectName:     app.ProjectName,
		OrgID:           app.OrgID,
		OrgName:         app.OrgName,
		PipelineID:      dto.ID,
		ServicesNames:   names,
	}, nil
}

// getServicesNames get servicesNames by diceYml
func getServicesNames(diceYml string) ([]string, error) {
	yml, err := diceyml.New([]byte(diceYml), false)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0)
	for k := range yml.Obj().Services {
		names = append(names, k)
	}
	return names, nil
}

func IsDeploying(status apistructs.DeploymentStatus) bool {
	switch status {
	// report error, we no longer support auto-cancel
	case apistructs.DeploymentStatusWaitApprove, apistructs.DeploymentStatusInit, apistructs.DeploymentStatusWaiting, apistructs.DeploymentStatusDeploying:
		return true
	default:
		return false
	}
}

// get pod autoscaler rules in k8s, include hpa and vpa
func (r *Runtime) AppliedScaledObjects(uniqueID spec.RuntimeUniqueId) (map[string]string, map[string]string, error) {
	hpaRules, err := r.db.GetRuntimeHPAByServices(uniqueID, nil)
	if err != nil {
		return nil, nil, errors.Errorf("get runtime HPA rules by RuntimeUniqueId %#v failed: %v", uniqueID, err)
	}
	hpaScaledRules := make(map[string]string)
	for _, rule := range hpaRules {
		// only applied rules need to delete
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			hpaScaledRules[rule.ServiceName] = rule.Rules
		}
	}

	vpaRules, err := r.db.GetRuntimeVPAByServices(uniqueID, nil)
	if err != nil {
		return nil, nil, errors.Errorf("get runtime VPA rules by RuntimeUniqueId %#v failed: %v", uniqueID, err)
	}
	vpaScaledRules := make(map[string]string)
	for _, rule := range vpaRules {
		// only applied rules need to delete
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			vpaScaledRules[rule.ServiceName] = rule.Rules
		}
	}

	return hpaScaledRules, vpaScaledRules, nil
}

func (r *Runtime) deleteRuntimePA(uniqueID spec.RuntimeUniqueId, runtimeId uint64) error {
	hpaRules, err := r.db.GetRuntimeHPAByServices(uniqueID, nil)
	if err != nil {
		return err
	}
	for _, rule := range hpaRules {
		// only applied rules need to delete
		if err = r.db.DeleteRuntimeHPAByRuleId(rule.ID); err != nil {
			return errors.Errorf("delete runtime hpa rule by rule_id %s failed: %v", rule.ID, err)
		}
	}

	vpaRules, err := r.db.GetRuntimeVPAByServices(uniqueID, nil)
	if err != nil {
		return err
	}
	for _, rule := range vpaRules {
		// only applied rules need to delete
		if err = r.db.DeleteRuntimeVPAByRuleId(rule.ID); err != nil {
			return errors.Errorf("delete runtime hpa rule by rule_id %s failed: %v", rule.ID, err)
		}
	}

	err = r.db.DeleteRuntimeVPARecommendationsByRuntimeId(runtimeId)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) GetOrg(orgID uint64) (*orgpb.Org, error) {
	if orgID == 0 {
		return nil, fmt.Errorf("the orgID is 0")
	}
	orgResp, err := r.org.GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcOrchestrator), &orgpb.GetOrgRequest{
		IdOrName: strconv.FormatUint(orgID, 10),
	})
	if err != nil {
		return nil, err
	}
	return orgResp.Data, nil
}
