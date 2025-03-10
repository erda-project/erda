package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"gopkg.in/yaml.v3"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/pkg/transport"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/diceworkspace"
	"github.com/erda-project/erda/internal/pkg/gitflowutil"
	"github.com/erda-project/erda/internal/pkg/user"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/internal/tools/orchestrator/utils"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

// DeployContext 部署上下文
type DeployContext struct {
	Runtime        *dbclient.Runtime
	App            *apistructs.ApplicationDTO
	LastDeployment *dbclient.Deployment
	// ReleaseId to deploy
	ReleaseID  string
	Operator   string
	DeployType string

	// Extras:
	// used for pipeline
	BuildID uint64
	// used for ability
	AddonActions map[string]interface{}
	// used for runtime-addon
	InstanceID string
	// used for
	Scale0 bool

	// 不由 orchestrator 来推进部署
	SkipPushByOrch bool

	// deployment order
	DeploymentOrderId string
	Param             string
}

func (r *Service) Create(operator user.ID, req *apistructs.RuntimeCreateRequest) (*apistructs.DeploymentCreateResponseDTO, error) {
	// TODO: 需要等 pipeline action 调用走内网后，再从 header 中取 User-ID (operator)
	// TODO: should not assign like this
	//req.Operator = operator.String()
	if err := checkRuntimeCreateReq(req); err != nil {
		return nil, apierrors.ErrCreateRuntime.InvalidParameter(err)
	}
	var appID uint64
	if req.Source == apistructs.ABILITY {
		return nil, apierrors.ErrCreateRuntime.InvalidParameter("end support: ABILITY")
	} else {
		// appID already bean checked
		appID = req.Extra.ApplicationID
	}
	app, err := r.bundle.GetApp(appID)
	if err != nil {
		// TODO: shall minimize unknown error
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	// TODO 暂时封闭
	//if app.Mode == string(apistructs.ApplicationModeLibrary) {
	//	return nil, apierrors.ErrCreateRuntime.InvalidParameter("Non-business applications cannot be published.")
	//}

	resource := apistructs.NormalBranchResource
	rules, err := r.bundle.GetProjectBranchRules(app.ProjectID)
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	if diceworkspace.GetValidBranchByGitReference(req.Name, rules).IsProtect {
		resource = apistructs.ProtectedBranchResource
	}
	perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   operator.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  app.ID,
		Resource: resource,
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrCreateRuntime.AccessDenied()
	}
	ctx := transport.WithHeader(context.Background(), metadata.New(map[string]string{httputil.InternalHeader: "cmp"}))
	resp, err := r.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{IdOrName: req.ClusterName})
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InvalidState(fmt.Sprintf("cluster: %v not found", req.ClusterName))
	}
	cluster := resp.Data

	// build runtimeUniqueId
	uniqueID := spec.RuntimeUniqueId{ApplicationId: appID, Workspace: req.Extra.Workspace, Name: req.Name}

	// prepare runtime
	// TODO: we do not need RepoAbbrev
	runtime, created, err := r.db.FindRuntimeOrCreate(uniqueID, req.Operator, req.Source, req.ClusterName,
		uint64(cluster.Id), app.GitRepoAbbrev, req.Extra.ProjectID, app.OrgID, req.DeploymentOrderId, req.ReleaseVersion, req.ExtraParams)
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	if created {
		// emit runtime add event
		event := events.RuntimeEvent{
			EventName: events.RuntimeCreated,
			Operator:  req.Operator,
			Runtime:   dbclient.ConvertRuntimeDTO(runtime, app),
		}
		r.evMgr.EmitEvent(&event)
	}

	// find last deployment
	last, err := r.db.FindLastDeployment(runtime.ID)
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	if last != nil && IsDeploying(last.Status) {
		return nil, apierrors.ErrCreateRuntime.InvalidState("正在部署中，请不要重复部署")
	}
	deploytype := "BUILD"
	if req.Extra.DeployType == "RELEASE" {
		deploytype = "RELEASE"
	}
	deployContext := DeployContext{
		Runtime:           runtime,
		App:               app,
		LastDeployment:    last,
		ReleaseID:         req.ReleaseID,
		Operator:          req.Operator,
		BuildID:           req.Extra.BuildID,
		DeployType:        deploytype,
		AddonActions:      req.Extra.AddonActions,
		InstanceID:        req.Extra.InstanceID.String(),
		SkipPushByOrch:    req.SkipPushByOrch,
		Param:             req.Param,
		DeploymentOrderId: req.DeploymentOrderId,
	}

	return r.DoDeployRuntime(&deployContext)
}

func (r *Service) GetOrg(orgID uint64) (*orgpb.Org, error) {
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

func (r *Service) checkOrgDeployBlocked(orgID uint64, runtime *dbclient.Runtime) (bool, error) {
	org, err := r.GetOrg(orgID)
	if err != nil {
		return false, err
	}
	blocked := false
	switch runtime.Workspace {
	case "DEV":
		blocked = org.BlockoutConfig.BlockDev
	case "TEST":
		blocked = org.BlockoutConfig.BlockTest
	case "STAGING":
		blocked = org.BlockoutConfig.BlockStage
	case "PROD":
		blocked = org.BlockoutConfig.BlockProd
	}
	if blocked {
		app, err := r.bundle.GetApp(runtime.ApplicationID)
		if err != nil {
			return false, err
		}
		if app.UnBlockStart == nil || app.UnBlockEnd == nil {
			return true, nil
		}
		now := time.Now()
		if app.UnBlockStart.Before(now) && app.UnBlockEnd.After(now) {
			return false, nil
		}
	}
	return blocked, nil
}

func (r *Service) PreCheck(dice *diceyml.DiceYaml, workspace string) error {
	defaultGroup := 10

	addonLi := make([]*diceyml.AddOn, 0, len(dice.Obj().AddOns))
	for _, a := range dice.Obj().AddOns {
		addonLi = append(addonLi, a)
	}

	errCh := make(chan error, len(addonLi))

	for i := 0; i < len(addonLi); i += defaultGroup {
		group := make([]*diceyml.AddOn, 0, defaultGroup)
		if len(addonLi)-i < defaultGroup {
			group = addonLi[i:]
		} else {
			group = addonLi[i : i+defaultGroup]
		}

		logrus.Debugf("current to check addon group: %+v, count: %d", group, len(group))

		wg := sync.WaitGroup{}
		wg.Add(len(group))
		for _, addOn := range group {
			go func(addOn *diceyml.AddOn) {
				defer wg.Done()
				addonName, addonPlan, err := r.Addon.ParseAddonFullPlan(addOn.Plan)
				if err != nil {
					errCh <- errors.Errorf("addon %s: %v", addonName, err)
					return
				}
				ok, err := r.Addon.CheckDeployCondition(addonName, addonPlan, workspace)
				if err != nil {
					errCh <- errors.Errorf("addon %s: %s", addonName, err)
					return
				}
				if !ok {
					errCh <- errors.Errorf("addon %s: basic plan addon cannot be used in production environment", addonName)
					return
				}
			}(addOn)
		}
		wg.Wait()
	}

	close(errCh)

	errs := make([]string, 0)
	for err := range errCh {
		errs = append(errs, err.Error())
	}

	if len(errs) != 0 {
		logrus.Errorf("do deploy runtime static precheck errors: %+v", errs)
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (r *Service) syncRuntimeServices(runtimeID uint64, dice *diceyml.DiceYaml) error {
	for name, service := range dice.Obj().Services {
		var envs string
		envsStr, err := json.Marshal(service.Envs)
		if err == nil {
			envs = string(envsStr)
		}
		var ports string
		portsStr, err := json.Marshal(service.Ports)
		if err == nil {
			ports = string(portsStr)
		}
		err = r.db.CreateOrUpdateRuntimeService(&dbclient.RuntimeService{
			RuntimeId:   runtimeID,
			ServiceName: name,
			Replica:     service.Deployments.Replicas,
			Status:      apistructs.ServiceStatusUnHealthy,
			Cpu:         fmt.Sprintf("%f", service.Resources.CPU),
			Mem:         service.Resources.Mem,
			Environment: envs,
			Ports:       ports,
		}, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Service) DoDeployRuntime(ctx *DeployContext) (*apistructs.DeploymentCreateResponseDTO, error) {
	// fetch & parse diceYml
	dice, err := r.bundle.GetDiceYAML(ctx.ReleaseID, ctx.Runtime.Workspace)
	if err != nil {
		return nil, err
	}

	// pre check
	if err := r.PreCheck(dice, ctx.Runtime.Workspace); err != nil {
		logrus.Errorf("deploy runtime pre check failed, error: %v", err)
		return nil, err
	}

	// build runtimeUniqueId
	uniqueID := spec.RuntimeUniqueId{
		ApplicationId: ctx.Runtime.ApplicationID,
		Workspace:     ctx.Runtime.Workspace,
		Name:          ctx.Runtime.Name,
	}

	// prepare pre deployments
	pre, err := r.db.FindPreDeploymentOrCreate(uniqueID, dice)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	diceYmlObj := dice.Obj()
	if pre.DiceOverlay != "" {
		var overlay diceyml.Object
		err = json.Unmarshal([]byte(pre.DiceOverlay), &overlay)
		if err != nil {
			return nil, apierrors.ErrDeployRuntime.InternalError(err)
		}
		utils.ApplyOverlay(diceYmlObj, &overlay)
	}
	if ctx.Scale0 {
		var scaleValue = 0
		for _, v := range diceYmlObj.Services {
			v.Deployments.Replicas = scaleValue
		}
	}

	// do sync RuntimeService table after dice.yml changed
	err = r.syncRuntimeServices(ctx.Runtime.ID, dice)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}

	// double check last deployment not active
	if ctx.LastDeployment != nil {
		switch ctx.LastDeployment.Status {
		case apistructs.DeploymentStatusWaitApprove, apistructs.DeploymentStatusInit, apistructs.DeploymentStatusWaiting, apistructs.DeploymentStatusDeploying:
			return nil, apierrors.ErrDeployRuntime.InvalidState("正在部署中，请不要重复部署")
		}
	}

	images := make(map[string]string)
	for name, s := range dice.Obj().Services {
		images[name] = s.Image
	}
	// check all services has it's image
	for name := range diceYmlObj.Services {
		if images[name] == "" {
			errMsg := fmt.Sprintf("bad release(%s), no image exist for service: %s",
				ctx.ReleaseID, name)
			return nil, apierrors.ErrDeployRuntime.InvalidState(errMsg)
		}
	}
	imageJSONByte, err := json.Marshal(images)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	imageJSON := string(imageJSONByte)
	diceJSONByte, err := json.Marshal(diceYmlObj)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	diceJSON := string(diceJSONByte)

	// 检查是否处于封网状态
	blocked, err := r.checkOrgDeployBlocked(ctx.Runtime.OrgID, ctx.Runtime)
	if err != nil {
		return nil, apierrors.ErrRollbackRuntime.InternalError(err)
	}
	status := apistructs.DeploymentStatusWaiting
	reason := ""
	needApproval := false
	branchrules, err := r.bundle.GetProjectBranchRules(ctx.Runtime.ProjectID)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	branch := diceworkspace.GetValidBranchByGitReference(ctx.Runtime.Name, branchrules)
	if blocked {
		status = apistructs.DeploymentStatusFailed
		reason = "企业封网中,无法部署"
	} else {
		// 检查 branchrule 来判断是否需要审批
		if branch.NeedApproval {
			status = apistructs.DeploymentStatusWaitApprove
			needApproval = true
		}
	}
	deployment := dbclient.Deployment{
		RuntimeId:         ctx.Runtime.ID,
		Status:            status,
		Phase:             "INIT",
		FailCause:         reason,
		Operator:          ctx.Operator,
		ReleaseId:         ctx.ReleaseID,
		BuildId:           ctx.BuildID,
		Dice:              diceJSON,
		Type:              ctx.DeployType,
		DiceType:          1,
		BuiltDockerImages: imageJSON,
		NeedApproval:      needApproval,
		ApprovalStatus:    map[bool]string{true: "WaitApprove", false: ""}[needApproval],
		SkipPushByOrch:    ctx.SkipPushByOrch,
		Param:             ctx.Param,
		DeploymentOrderId: ctx.DeploymentOrderId,
	}
	if err := r.db.CreateDeployment(&deployment); err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}

	// 发送 审批站内信
	if !blocked && branch.NeedApproval {
		for range []int{0} {
			approvers, err := r.bundle.ListMembers(apistructs.MemberListRequest{
				ScopeType:        "project",
				ScopeID:          int64(ctx.Runtime.ProjectID),
				Roles:            []string{"Owner", "Lead"},
				PageSize:         99,
				DesensitizeEmail: false,
			})
			if err != nil {
				logrus.Errorf("failed to listmembers: %v", err)
				break
			}
			approverIDs := []string{}
			emails := []string{}
			for _, approver := range approvers {
				approverIDs = append(approverIDs, approver.UserID)
				emails = append(emails, approver.Email)
			}

			memberlist, err := r.bundle.ListUsers(apistructs.UserListRequest{
				UserIDs: []string{deployment.Operator},
			})
			if err != nil {
				logrus.Errorf("failed to listuser(%s): %v", deployment.Operator, err)
				break
			}
			if len(memberlist.Users) == 0 {
				break
			}
			member := memberlist.Users[0].Name
			proj, err := r.bundle.GetProject(ctx.Runtime.ProjectID)
			if err != nil {
				logrus.Errorf("failed to get project(%d): %v", ctx.Runtime.ProjectID, err)
				break
			}
			app, err := r.bundle.GetApp(ctx.Runtime.ApplicationID)
			if err != nil {
				logrus.Errorf("failed to get app(%d): %v", ctx.Runtime.ApplicationID, err)
				break
			}
			d, err := r.clusterinfoImpl.Info(ctx.Runtime.ClusterName)
			if err != nil {
				logrus.Errorf("failed to QueryClusterInfo: %v", err)
			}
			protocols := strutil.Split(d.Get(apistructs.DICE_PROTOCOL), ",")
			protocol := "https"
			if len(protocols) > 0 {
				protocol = protocols[0]
			}
			domain := d.Get(apistructs.DICE_ROOT_DOMAIN)
			org, err := r.GetOrg(ctx.Runtime.OrgID)
			if err != nil {
				logrus.Errorf("failed to getorg(%v):%v", ctx.Runtime.OrgID, err)
				break
			}

			url := fmt.Sprintf("%s://%s-org.%s/workBench/approval/my-approve/pending?id=%d",
				protocol, org.Name, domain, deployment.ID)
			if err := r.bundle.CreateMboxNotify("notify.deployapproval.launch.markdown_template",
				map[string]string{
					"title":       fmt.Sprintf("【重要】请及时审核%s项目%s应用部署合规性", proj.Name, app.Name),
					"member":      member,
					"projectname": proj.Name,
					"appName":     app.Name,
					"url":         url,
				},
				"zh-CN", ctx.Runtime.OrgID, approverIDs); err != nil {
				logrus.Errorf("failed to CreateMboxNotify: %v", err)
			}
			if err := r.bundle.CreateEmailNotify("notify.deployapproval.launch.markdown_template",
				map[string]string{
					"title":       fmt.Sprintf("【重要】请及时审核%s项目%s应用部署合规性", proj.Name, app.Name),
					"member":      member,
					"projectname": proj.Name,
					"appName":     app.Name,
					"url":         url,
				},
				"zh-CN", ctx.Runtime.OrgID, emails); err != nil {
				logrus.Errorf("failed to CreateEmailNotify: %v", err)
			}
		}
	}

	// emit runtime deploy start event
	event := events.RuntimeEvent{
		EventName:  events.RuntimeDeployStart,
		Operator:   ctx.Operator,
		Runtime:    dbclient.ConvertRuntimeDTO(ctx.Runtime, ctx.App),
		Deployment: deployment.Convert(),
	}
	r.evMgr.EmitEvent(&event)

	// TODO: the response should be apistructs.RuntimeDTO
	return &apistructs.DeploymentCreateResponseDTO{
		DeploymentID:  deployment.ID,
		ApplicationID: ctx.Runtime.ApplicationID, // TODO: will all runtime has applicationId ?
		RuntimeID:     ctx.Runtime.ID,
	}, nil
}

func checkRuntimeCreateReq(req *apistructs.RuntimeCreateRequest) error {
	if req.Name == "" {
		return errors.New("runtime name is not specified")
	}
	if req.ReleaseID == "" {
		return errors.New("releaseId is not specified")
	}
	if req.Operator == "" {
		return errors.New("operator is not specified")
	}
	if req.ClusterName == "" {
		return errors.New("clusterName is not specified")
	}
	switch req.Source {
	case apistructs.PIPELINE:
	case apistructs.ABILITY:
	case apistructs.RUNTIMEADDON:
	case apistructs.RELEASE:
	default:
		return errors.New("source is unknown")
	}
	if req.Extra.OrgID == 0 {
		return errors.New("extra.orgId is not specified")
	}
	if req.Source == apistructs.PIPELINE || req.Source == apistructs.RUNTIMEADDON || req.Source == apistructs.RELEASE {
		if req.Extra.ProjectID == 0 {
			return errors.New("extra.projectId is not specified, for pipeline")
		}
		if req.Extra.ApplicationID == 0 {
			return errors.New("extra.applicationId is not specified, for pipeline")
		}
	}
	if req.Source == apistructs.RUNTIMEADDON {
		if len(req.Extra.InstanceID) == 0 {
			return errors.New("extra.instanceId is not specified, for runtimeaddon")
		}
	}
	if req.Source == apistructs.ABILITY {
		if req.Extra.ApplicationName == "" {
			return errors.New("extra.applicationName are not specified, for ability")
		}
	}
	if req.Extra.Workspace == "" {
		return errors.New("extra.workspace is not specified")
	}
	return nil
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

// Create 创建应用实例
func (r *Service) CreateByReleaseID(ctx context.Context, operator user.ID, releaseReq *apistructs.RuntimeReleaseCreateRequest) (*apistructs.DeploymentCreateResponseDTO, error) {
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
	releaseResp, err := r.releaseSvc.GetRelease(ctx, &pb.ReleaseGetRequest{ReleaseID: releaseReq.ReleaseID})
	if err != nil {
		return nil, err
	}
	if releaseReq == nil {
		return nil, errors.Errorf("releaseId does not exist")
	}
	if releaseReq.ProjectID != uint64(releaseResp.Data.ProjectID) {
		return nil, errors.Errorf("release does not correspond to the project")
	}
	if releaseReq.ApplicationID != uint64(releaseResp.Data.ApplicationID) {
		return nil, errors.Errorf("release does not correspond to the application")
	}
	branchWorkspaces, err := r.bundle.GetAllValidBranchWorkspace(releaseReq.ApplicationID, string(operator))
	if err != nil {
		return nil, apierrors.ErrCreateRuntime.InternalError(err)
	}
	_, validArtifactWorkspace := gitflowutil.IsValidBranchWorkspace(branchWorkspaces, apistructs.DiceWorkspace(releaseReq.Workspace))
	if !validArtifactWorkspace {
		return nil, errors.Errorf("release does not correspond to the workspace")
	}

	projectInfo, err := r.bundle.GetProject(releaseReq.ProjectID)
	if err != nil {
		return nil, err
	}
	if projectInfo == nil {
		return nil, errors.Errorf("The project is illegal")
	}

	wsCluster, ok := projectInfo.ClusterConfig[releaseReq.Workspace]
	if !ok {
		return nil, fmt.Errorf("workspace corresponding cluster is empty")
	}
	var targetClusterName string
	// 跨集群部署
	// 部署的目标集群，默认情况下为 release 所属的集群，跨集群部署时，部署到项目环境对应的集群
	if releaseResp.Data.CrossCluster {
		targetClusterName = wsCluster
	} else {
		// 在制品所属集群部署
		// 校验制品所属集群和环境对应集群是否相同
		if releaseResp.Data.ClusterName != wsCluster {
			return nil, fmt.Errorf("release does not correspond to the cluster")
		}
		targetClusterName = releaseResp.Data.ClusterName
	}

	var req apistructs.RuntimeCreateRequest
	req.ClusterName = targetClusterName
	req.Name = releaseResp.Data.ReleaseName
	req.Operator = operator.String()
	req.Source = "RELEASE"
	req.ReleaseID = releaseReq.ReleaseID
	req.SkipPushByOrch = false

	var extra apistructs.RuntimeCreateRequestExtra
	extra.OrgID = uint64(releaseResp.Data.OrgID)
	extra.ProjectID = uint64(releaseResp.Data.ProjectID)
	extra.ApplicationID = uint64(releaseResp.Data.ApplicationID)
	extra.ApplicationName = releaseResp.Data.ApplicationName
	extra.Workspace = releaseReq.Workspace
	extra.DeployType = "RELEASE"
	req.Extra = extra

	return r.Create(operator, &req)
}

// List 查询应用实例列表
func (r *Service) List(userID user.ID, orgID uint64, appID uint64, workspace, name string) ([]apistructs.RuntimeSummaryDTO, error) {
	var l = logrus.WithField("func", "*Runtime.List")
	var runtimes []dbclient.Runtime
	if len(workspace) > 0 && len(name) > 0 {
		r, err := r.db.FindRuntime(spec.RuntimeUniqueId{ApplicationId: appID, Workspace: workspace, Name: name})
		if err != nil {
			return nil, apierrors.ErrListRuntime.InternalError(err)
		}
		if r != nil {
			runtimes = append(runtimes, *r)
		}
	} else {
		v, err := r.db.FindRuntimesByAppId(appID)
		if err != nil {
			return nil, apierrors.ErrListRuntime.InternalError(err)
		}
		runtimes = v
	}
	app, err := r.bundle.GetApp(appID)
	if err != nil {
		return nil, err
	}

	// check four env perm
	rtEnvPermBranchMark := make(map[string][]string)
	anyPerm := false
	for _, env := range []string{"dev", "test", "staging", "prod"} {
		perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.AppScope,
			ScopeID:  app.ID,
			Resource: "runtime-" + env,
			Action:   apistructs.GetAction,
		})
		if err != nil {
			return nil, apierrors.ErrGetRuntime.InternalError(err)
		}
		if perm.Access {
			rtEnvPermBranchMark[env] = []string{}
			anyPerm = true
		}
	}
	if !anyPerm {
		return nil, apierrors.ErrGetRuntime.AccessDenied()
	}

	// TODO: apistructs.RuntimeSummaryDTO should be combine into apistructs.Runtime
	var data []apistructs.RuntimeSummaryDTO
	for _, runtime := range runtimes {
		if runtime.OrgID != orgID {
			continue
		}
		env := strutil.ToLower(runtime.Workspace)
		// If the user does not have the permission of this environment,
		// the runtime data in this environment will not be returned
		if _, exists := rtEnvPermBranchMark[env]; !exists {
			continue
		}
		// record all runtime's branchs in each environment
		rtEnvPermBranchMark[env] = append(rtEnvPermBranchMark[env], runtime.GitBranch)
		deployment, err := r.db.FindLastDeployment(runtime.ID)
		if err != nil {
			l.WithError(err).WithField("runtime.ID", runtime.ID).
				Warnln("failed to build summary item, failed to get last deployment")
			return nil, err
		}
		var d apistructs.RuntimeSummaryDTO
		if err := r.convertRuntimeSummaryDTOFromRuntimeModel(&d, runtime, deployment); err != nil {
			l.WithError(err).WithField("runtime.ID", runtime.ID).
				Warnln("failed to convertRuntimeSummaryDTOFromRuntimeModel")
			continue
		}
		data = append(data, d)
	}

	return data, nil
}

func (r *Service) convertRuntimeSummaryDTOFromRuntimeModel(d *apistructs.RuntimeSummaryDTO, runtime dbclient.Runtime, deployment *dbclient.Deployment) error {
	var l = logrus.WithField("func", "Runtime.convertRuntimeInspectDTOFromRuntimeModel")

	if d == nil {
		err := errors.New("the DTO is nil")
		l.WithError(err).Warnln()
		return err
	}

	isFakeRuntime := false
	// TODO: Deprecated, instead from runtime deployment_status filed
	if deployment == nil {
		isFakeRuntime = true
		// make a fake deployment
		deployment = &dbclient.Deployment{
			RuntimeId: runtime.ID,
			Status:    apistructs.DeploymentStatusInit,
			BaseModel: dbengine.BaseModel{
				UpdatedAt: runtime.UpdatedAt,
			},
		}
	}

	d.ID = runtime.ID
	d.Name = runtime.Name
	d.ServiceGroupNamespace = runtime.ScheduleName.Namespace
	d.ServiceGroupName = runtime.ScheduleName.Name
	d.Source = runtime.Source
	d.Status = runtime.Status
	d.Services = make(map[string]*apistructs.RuntimeInspectServiceDTO)
	d.DeployStatus = deployment.Status
	// 如果还 deployment 的状态不是终态, runtime 的状态返回为 init(前端显示为部署中效果),
	// 不然开始部署直接变为不健康不合理
	if deployment.Status == apistructs.DeploymentStatusDeploying ||
		deployment.Status == apistructs.DeploymentStatusWaiting ||
		deployment.Status == apistructs.DeploymentStatusInit ||
		deployment.Status == apistructs.DeploymentStatusWaitApprove {
		d.Status = apistructs.RuntimeStatusInit
	}
	if runtime.LegacyStatus == dbclient.LegacyStatusDeleting {
		d.DeleteStatus = dbclient.LegacyStatusDeleting
	}
	d.DeploymentOrderId = runtime.DeploymentOrderId
	d.DeploymentOrderName = utils.ParseOrderName(runtime.DeploymentOrderId)
	d.ReleaseVersion = runtime.ReleaseVersion
	d.ReleaseID = deployment.ReleaseId
	d.ClusterID = runtime.ClusterId
	d.ClusterName = runtime.ClusterName
	d.ReleaseVersion = runtime.ReleaseVersion
	d.Creator = runtime.Creator
	d.ApplicationID = runtime.ApplicationID
	d.CreatedAt = runtime.CreatedAt
	d.UpdatedAt = runtime.UpdatedAt
	d.RawStatus = runtime.Status
	d.RawDeploymentStatus = string(deployment.Status)
	d.TimeCreated = runtime.CreatedAt
	d.Extra = map[string]interface{}{
		"applicationId": runtime.ApplicationID,
		"workspace":     runtime.Workspace,
		"buildId":       deployment.BuildId,
		"fakeRuntime":   isFakeRuntime,
	}
	d.ProjectID = runtime.ProjectID
	UpdateStatusToDisplay(&d.RuntimeInspectDTO)
	if deployment.Status == apistructs.DeploymentStatusDeploying {
		UpdateStatusWhenDeploying(&d.RuntimeInspectDTO)
	}
	d.LastOperator = deployment.Operator
	d.LastOperatorId = deployment.ID
	d.LastOperateTime = deployment.UpdatedAt // TODO: use a standalone OperateTime
	if d.LastOperator == "" {
		d.LastOperateTime = runtime.UpdatedAt
	}
	return nil
}

// UpdateStatusToDisplay 修改内部状态为展示状态
func UpdateStatusToDisplay(runtime *apistructs.RuntimeInspectDTO) {
	if runtime == nil {
		return
	}
	runtime.Status = isStatusForDisplay(runtime.Status)
	for key := range runtime.Services {
		runtime.Services[key].Status = isStatusForDisplay(runtime.Services[key].Status)
	}
}

// UpdateStatusWhenDeploying 将 UnHealthy 修改为 Progressing（部署中）
func UpdateStatusWhenDeploying(runtime *apistructs.RuntimeInspectDTO) {
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

// Redeploy 重新部署
func (r *Service) Redeploy(operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.DeploymentCreateResponseDTO, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	app, err := r.bundle.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   operator.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  app.ID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrDeployRuntime.AccessDenied()
	}
	deployment, err := r.db.FindLastSuccessDeployment(runtimeID)
	if err != nil {
		return nil, apierrors.ErrDeployRuntime.InternalError(err)
	}
	if deployment == nil {
		// it will happen, but it often implicit some errors
		return nil, apierrors.ErrDeployRuntime.InvalidState("last success deployment not found")
	}
	if deployment.ReleaseId == "" {
		return nil, apierrors.ErrDeployRuntime.InvalidState("抱歉，检测到不兼容的部署任务，请去重新构建")
	}
	switch deployment.Status {
	case apistructs.DeploymentStatusWaitApprove, apistructs.DeploymentStatusInit, apistructs.DeploymentStatusWaiting, apistructs.DeploymentStatusDeploying:
		// we do not cancel, just report error
		return nil, apierrors.ErrDeployRuntime.InvalidState("正在部署中，请不要重复部署")
	}

	deployContext := DeployContext{
		Runtime:        runtime,
		App:            app,
		LastDeployment: deployment,
		DeployType:     "REDEPLOY",
		ReleaseID:      deployment.ReleaseId,
		Operator:       operator.String(),
		SkipPushByOrch: false,
	}
	return r.DoDeployRuntime(&deployContext)
}

// FullGCService 定时全量 GC 过期的部署单
func (r *Service) FullGCService() {
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
func (r *Service) getRollbackConfig() (map[uint64]map[string]int, error) {
	result := make(map[uint64]map[string]int, 0)
	// TODO: use cache to get project info
	projects, err := r.bundle.GetAllProjects()
	if err != nil {
		return nil, err
	}
	for _, prj := range projects {
		result[prj.ID] = prj.RollbackConfig
	}
	return result, nil
}

// ReferClusterService 查看 runtime & addon 是否有使用集群
func (r *Service) ReferClusterService(clusterName string, orgID uint64) bool {
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

// RuntimeDeployLogs deploy发布日志接口
func (r *Service) RuntimeDeployLogs(userID user.ID, orgID uint64, orgName string, deploymentID uint64, paramValues url.Values) (*apistructs.DashboardSpotLogData, error) {
	deployment, err := r.db.GetDeployment(deploymentID)
	if err != nil {
		return nil, apierrors.ErrGetRuntime.InternalError(err)
	}
	if deployment == nil {
		return nil, apierrors.ErrGetRuntime.InvalidParameter(errors.Errorf("deployment not found, id: %d", deploymentID))
	}
	if err := r.checkRuntimeScopePermission(userID, deployment.RuntimeId); err != nil {
		return nil, err
	}
	return r.requestMonitorLog(strconv.FormatUint(deploymentID, 10), orgName, paramValues, apistructs.DashboardSpotLogSourceDeploy)
}

// checkRuntimeScopePermission 检测runtime级别的权限
func (r *Service) checkRuntimeScopePermission(userID user.ID, runtimeID uint64) error {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return err
	}
	perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
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

// requestMonitorLog 调用bundle monitor log接口获取数据
func (r *Service) requestMonitorLog(requestID string, orgName string, paramValues url.Values, source apistructs.DashboardSpotLogSource) (*apistructs.DashboardSpotLogData, error) {
	// 获取日志
	var logReq apistructs.DashboardSpotLogRequest
	if err := queryStringDecoder.Decode(&logReq, paramValues); err != nil {
		return nil, err
	}
	logReq.ID = requestID
	logReq.Source = source

	logResult, err := r.bundle.GetLog(orgName, logReq)
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

// ListGroupByApps lists all runtimes for given apps.
// The key in the returned result map is appID.
func (r *Service) ListGroupByApps(appIDs []uint64, env string) (map[uint64][]*apistructs.RuntimeSummaryDTO, error) {
	var l = logrus.WithField("func", "*Runtime.ListGroupByApps")
	runtimes, ids, err := r.db.FindRuntimesInApps(appIDs, env)
	if err != nil {
		l.WithError(err).Errorln("failed to FindRuntimesInApps")
		return nil, err
	}
	deploymentIDs, err := r.db.FindLastDeploymentIDsByRutimeIDs(ids)
	if err != nil {
		l.WithError(err).Errorf("failed to list runtimes: %+v\n", err)
		return nil, err
	}
	deployments, err := r.db.FindDeploymentsByIDs(deploymentIDs)
	// note: internal API, do not check the permission
	var result = struct {
		sync.RWMutex
		m map[uint64][]*apistructs.RuntimeSummaryDTO
	}{m: make(map[uint64][]*apistructs.RuntimeSummaryDTO)}
	var wg sync.WaitGroup
	for appID, runtimeList := range runtimes {
		for _, runtime := range runtimeList {
			wg.Add(1)
			go func(runtime *dbclient.Runtime, deployment dbclient.Deployment) {
				r.generateListGroupAppResult(&result, appID, runtime, deployment, &wg)
			}(runtime, deployments[runtime.ID])
		}
	}
	wg.Wait()
	return result.m, nil
}

func (r *Service) generateListGroupAppResult(result *struct {
	sync.RWMutex
	m map[uint64][]*apistructs.RuntimeSummaryDTO
}, appID uint64,
	runtime *dbclient.Runtime, deployment dbclient.Deployment, wg *sync.WaitGroup) {
	var l = logrus.WithField("func", "*Runtime.ListGroupByApps")
	var d apistructs.RuntimeSummaryDTO
	if err := r.convertRuntimeSummaryDTOFromRuntimeModel(&d, *runtime, &deployment); err != nil {
		l.WithError(err).WithField("runtime.ID", runtime.ID).
			Warnln("failed to convertRuntimeSummaryDTOFromRuntimeModel")
	}
	result.Lock()
	result.m[appID] = append(result.m[appID], &d)
	result.Unlock()
	wg.Done()
}

// CountARByWorkspace count appliaction runtimes by workspace .
func (r *Service) CountARByWorkspace(appId uint64, env string) (uint64, error) {
	var l = logrus.WithField("func", "*Runtime.CountPRByWorkspace")
	cnt, err := r.db.GetAppRuntimeNumberByWorkspace(appId, env)
	if err != nil {
		l.WithError(err).Errorln("failed to CountPRByWorkspace")
		return 0, err
	}
	return cnt, nil
}

func (r *Service) GetServiceByRuntime(runtimeIDs []uint64) (map[uint64]*apistructs.RuntimeSummaryDTO, error) {
	logrus.Debug("get services started")
	var l = logrus.WithField("func", "*Runtime.GetServiceByRuntime")
	runtimes, err := r.db.FindRuntimesByIds(runtimeIDs)
	if err != nil {
		return nil, err
	}
	var servicesMap = struct {
		sync.RWMutex
		m map[uint64]*apistructs.RuntimeSummaryDTO
	}{m: make(map[uint64]*apistructs.RuntimeSummaryDTO)}
	wg := sync.WaitGroup{}
	for i := 0; i < len(runtimes); i++ {
		runtime := runtimes[i]
		deployment, err := r.db.FindLastDeployment(runtime.ID)
		if err != nil {
			l.WithError(err).WithField("runtime.ID", runtime.ID).
				Warnln("failed to build summary item, failed to get last deployment")
			continue
		}
		runtimeHPARules, err := r.db.GetRuntimeHPARulesByRuntimeId(runtime.ID)
		if err != nil {
			l.WithError(err).WithField("runtime.ID", runtime.ID).
				Warnln("failed to build summary item, failed to get runtime hpa rules")
		}
		runtimeVPARules, err := r.db.GetRuntimeVPARulesByRuntimeId(runtime.ID)
		if err != nil {
			l.WithError(err).WithField("runtime.ID", runtime.ID).
				Warnln("failed to build summary item, failed to get runtime vpa rules")
		}
		if deployment == nil {
			// make a fake deployment
			deployment = &dbclient.Deployment{
				RuntimeId: runtime.ID,
				Status:    apistructs.DeploymentStatusInit,
				BaseModel: dbengine.BaseModel{
					UpdatedAt: runtime.UpdatedAt,
				},
			}
		}
		if runtime.ScheduleName.Namespace != "" && runtime.ScheduleName.Name != "" {
			wg.Add(1)
			go func(rt dbclient.Runtime, wg *sync.WaitGroup, servicesMap *struct {
				sync.RWMutex
				m map[uint64]*apistructs.RuntimeSummaryDTO
			}, deployment *dbclient.Deployment, runtimeHPARules []dbclient.RuntimeHPA, runtimeVPARules []dbclient.RuntimeVPA) {
				d := apistructs.RuntimeSummaryDTO{}
				sg, err := r.serviceGroupImpl.InspectServiceGroupWithTimeout(rt.ScheduleName.Namespace, rt.ScheduleName.Name)
				if err != nil {
					l.WithError(err).Warnf("failed to inspect servicegroup: %s/%s",
						rt.ScheduleName.Namespace, rt.ScheduleName.Name)
				} else if sg.Status == "Ready" || sg.Status == "Healthy" {
					d.Status = apistructs.RuntimeStatusHealthy
				}
				var dice diceyml.Object
				if err = json.Unmarshal([]byte(deployment.Dice), &dice); err != nil {
					logrus.Error(apierrors.ErrGetRuntime.InvalidState(strutil.Concat("dice.json invalid: ", err.Error())))
					return
				}
				d.Services = make(map[string]*apistructs.RuntimeInspectServiceDTO)
				fillRuntimeDataWithServiceGroup(&d.RuntimeInspectDTO, dice.Services, dice.Jobs, sg, nil, string(deployment.Status))
				UpdatePARuleEnabledStatusToDisplay(runtimeHPARules, runtimeVPARules, &d.RuntimeInspectDTO)
				servicesMap.Lock()
				servicesMap.m[rt.ID] = &d
				servicesMap.Unlock()
				wg.Done()
			}(runtime, &wg, &servicesMap, deployment, runtimeHPARules, runtimeVPARules)
		}
	}
	wg.Wait()
	logrus.Debug("get services finished")
	return servicesMap.m, nil

}

// fillRuntimeDataWithServiceGroup use serviceGroup's data to fill RuntimeInspectDTO
func fillRuntimeDataWithServiceGroup(data *apistructs.RuntimeInspectDTO, targetService diceyml.Services, targetJob diceyml.Jobs,
	sg *apistructs.ServiceGroup, domainMap map[string][]string, status string) {
	statusServiceMap := map[string]string{}
	replicaMap := map[string]int{}
	resourceMap := map[string]apistructs.RuntimeServiceResourceDTO{}
	statusMap := map[string]map[string]string{}
	if sg != nil {
		if sg.Status != apistructs.StatusReady && sg.Status != apistructs.StatusHealthy {
			for _, serviceItem := range sg.Services {
				statusMap[serviceItem.Name] = map[string]string{
					"Msg":    serviceItem.LastMessage,
					"Reason": serviceItem.Reason,
				}
			}
		}
		data.ModuleErrMsg = statusMap

		for _, v := range sg.Services {
			statusServiceMap[v.Name] = string(v.StatusDesc.Status)
			if statusServiceMap[v.Name] == "Ready" || statusServiceMap[v.Name] == "Healthy" {
				statusServiceMap[v.Name] = apistructs.RuntimeStatusHealthy
			}
			replicaMap[v.Name] = v.Scale
			resourceMap[v.Name] = apistructs.RuntimeServiceResourceDTO{
				CPU:  v.Resources.Cpu,
				Mem:  int(v.Resources.Mem),
				Disk: int(v.Resources.Disk),
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
		if (len(v.Expose) != 0 || svcPortExpose) && domainMap != nil {
			expose = domainMap[k]
		}

		runtimeInspectService := &apistructs.RuntimeInspectServiceDTO{
			Resources: apistructs.RuntimeServiceResourceDTO{
				CPU:  v.Resources.CPU,
				Mem:  int(v.Resources.Mem),
				Disk: int(v.Resources.Disk),
			},
			Envs:        v.Envs,
			Addrs:       convertInternalAddrs(sg, k),
			Expose:      expose,
			Status:      status,
			Deployments: apistructs.RuntimeServiceDeploymentsDTO{Replicas: 0},
		}
		if sgStatus, ok := statusServiceMap[k]; ok {
			runtimeInspectService.Status = sgStatus
		}
		if sgReplicas, ok := replicaMap[k]; ok {
			runtimeInspectService.Deployments.Replicas = sgReplicas
		}
		if sgResources, ok := resourceMap[k]; ok {
			runtimeInspectService.Resources = sgResources
		}

		data.Services[k] = runtimeInspectService
	}
	for k, v := range targetJob {
		runtimeInspectService := &apistructs.RuntimeInspectServiceDTO{
			Resources: apistructs.RuntimeServiceResourceDTO{
				CPU:  v.Resources.CPU,
				Mem:  int(v.Resources.Mem),
				Disk: int(v.Resources.Disk),
			},
			Envs:        v.Envs,
			Type:        "job",
			Status:      statusServiceMap[k],
			Deployments: apistructs.RuntimeServiceDeploymentsDTO{Replicas: 1},
		}
		data.Services[k] = runtimeInspectService
	}
	data.Resources = apistructs.RuntimeServiceResourceDTO{CPU: 0, Mem: 0, Disk: 0}
	for _, v := range data.Services {
		if v.Type == "job" {
			continue
		}
		data.Resources.CPU += v.Resources.CPU * float64(v.Deployments.Replicas)
		data.Resources.Mem += v.Resources.Mem * v.Deployments.Replicas
		data.Resources.Disk += v.Resources.Disk * v.Deployments.Replicas
	}
}

// UpdatePARuleEnabledStatusToDisplay 显示 service 对应是否开启 HPA, VPA
func UpdatePARuleEnabledStatusToDisplay(hpaRules []dbclient.RuntimeHPA, vpaRules []dbclient.RuntimeVPA, runtime *apistructs.RuntimeInspectDTO) {
	if runtime == nil {
		return
	}

	for svc := range runtime.Services {
		runtime.Services[svc].HPAEnabled = pstypes.RuntimePARuleCanceled
		runtime.Services[svc].VPAEnabled = pstypes.RuntimePARuleCanceled
	}

	for _, rule := range hpaRules {
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			runtime.Services[rule.ServiceName].HPAEnabled = pstypes.RuntimePARuleApplied
		}
	}

	for _, rule := range vpaRules {
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			runtime.Services[rule.ServiceName].VPAEnabled = pstypes.RuntimePARuleApplied
		}
	}
}

func (r *Service) fullGCForSingleRuntime(runtimeID uint64, keep int) {
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

func (r *Service) markOutdated(deployment *dbclient.Deployment) {
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

func (r *Service) RedeployPipeline(ctx context.Context, operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDeployDTO, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, err
	}
	yml := utils.GenRedeployPipelineYaml(runtimeID)
	app, err := r.bundle.GetApp(runtime.ApplicationID)
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
		commit, err := r.bundle.GetGittarCommit(app.GitRepoAbbrev, commitid, string(operator))
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

func (r *Service) setClusterName(rt *dbclient.Runtime) error {
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

// DeleteRuntime 标记应用实例删除
func (r *Service) DeleteRuntime(operator user.ID, orgID uint64, runtimeID uint64) (*apistructs.RuntimeDTO, error) {
	runtime, err := r.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrDeleteRuntime.InternalError(err)
	}
	// TODO: do not query app
	app, err := r.bundle.GetApp(runtime.ApplicationID)
	if err != nil {
		return nil, err
	}
	perm, err := r.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
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
		if _, err = r.bundle.InvalidateOAuth2Token(apistructs.OAuth2TokenInvalidateRequest{AccessToken: runtime.FileToken}); err != nil {
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

// AppliedScaledObjects get pod autoscaler rules in k8s, include hpa and vpa
func (r *Service) AppliedScaledObjects(uniqueID spec.RuntimeUniqueId) (map[string]string, map[string]string, error) {
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
