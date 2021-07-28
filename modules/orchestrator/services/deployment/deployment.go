// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package deployment 部署相关操作
package deployment

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/events"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/services/migration"
	"github.com/erda-project/erda/modules/orchestrator/services/resource"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Deployment 部署对象封装
type Deployment struct {
	db        *dbclient.DBClient
	evMgr     *events.EventManager
	bdl       *bundle.Bundle
	addon     *addon.Addon
	resource  *resource.Resource
	migration *migration.Migration
	encrypt   *encryption.EnvEncrypt
}

// Option 部署对象配置选项
type Option func(*Deployment)

// New 新建部署对象实例
func New(options ...Option) *Deployment {
	d := &Deployment{}
	for _, op := range options {
		op(d)
	}
	return d
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(d *Deployment) {
		d.db = db
	}
}

// WithEventManager 配置 EventManager
func WithEventManager(evMgr *events.EventManager) Option {
	return func(d *Deployment) {
		d.evMgr = evMgr
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(d *Deployment) {
		d.bdl = bdl
	}
}

// WithAddon 配置 addon service
func WithAddon(a *addon.Addon) Option {
	return func(d *Deployment) {
		d.addon = a
	}
}

// WithMigration 配置 Migration service
func WithMigration(m *migration.Migration) Option {
	return func(d *Deployment) {
		d.migration = m
	}
}

// WithEncrypt 配置 encrypt service
func WithEncrypt(encrypt *encryption.EnvEncrypt) Option {
	return func(d *Deployment) {
		d.encrypt = encrypt
	}
}

// WithResource 配置 Runtime service
func WithResource(resource *resource.Resource) Option {
	return func(d *Deployment) {
		d.resource = resource
	}
}

func (d *Deployment) ContinueDeploy(deploymentID uint64) error {
	// prepare the context
	fsm := NewFSMContext(deploymentID, d.db, d.evMgr, d.bdl, d.addon, d.migration, d.encrypt, d.resource)
	if err := fsm.Load(); err != nil {
		return errors.Wrapf(err, "failed to load fsm, deployment: %d, (%v)", deploymentID, err)
	}
	if fsm.Deployment.SkipPushByOrch {
		return nil
	}
	if end, err := fsm.timeout(); end || err != nil {
		return err
	}
	if err := fsm.precheck(); err != nil {
		return fsm.failDeploy(err)
	}
	switch fsm.Deployment.Status {
	case apistructs.DeploymentStatusWaiting:
		return fsm.continueWaiting()
	case apistructs.DeploymentStatusDeploying:
		return fsm.continueDeploying()
	case apistructs.DeploymentStatusCanceling:
		return fsm.continueCanceling()
	default:
		return nil
	}
}

func (d *Deployment) CancelLastDeploy(runtimeID uint64, operator string, force bool) error {
	deployment, err := d.db.FindLastDeployment(runtimeID)
	if err != nil {
		return apierrors.ErrCancelDeployment.InternalError(err)
	}
	if deployment == nil {
		return apierrors.ErrCancelDeployment.NotFound()
	}
	fsm := NewFSMContext(deployment.ID, d.db, d.evMgr, d.bdl, d.addon, d.migration, d.encrypt, d.resource)
	if err := fsm.Load(); err != nil {
		return err
	}
	perm, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   operator,
		Scope:    apistructs.AppScope,
		ScopeID:  fsm.Runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(fsm.Runtime.Workspace),
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return apierrors.ErrCancelDeployment.InternalError(err)
	}
	if !perm.Access {
		return apierrors.ErrCancelDeployment.AccessDenied()
	}
	return fsm.doCancelDeploy(operator, force)
}

// ListOrg 查询部署记录(列出orgid下所有有权限的deployments)
func (d *Deployment) ListOrg(userID user.ID, orgID uint64, needFilterProjectRole bool,
	needApproval *bool, approvedBy *user.ID, operateUsers []string, approved *bool,
	approvalStatus *string, types []string, ids []uint64, page apistructs.PageInfo) (
	*apistructs.DeploymentDetailListData, error) {
	myapps, err := d.bdl.GetMyApps(userID.String(), orgID)
	if err != nil {
		return nil, apierrors.ErrListDeployment.InternalError(err)
	}
	myapps_ := []apistructs.ApplicationDTO{}
	if needFilterProjectRole {
		userrolelist, err := d.bdl.ListMemberRolesByUser(apistructs.ListMemberRolesByUserRequest{
			UserID:    userID.String(),
			ScopeType: "project",
			ParentID:  int64(orgID),
			PageSize:  9999,
		})
		if err != nil {
			return nil, apierrors.ErrListDeployment.InternalError(err)
		}
		projectlist := []int64{}
		for _, ur := range userrolelist.List {
			if strutil.Contains(strutil.Concat(ur.Roles...), "Owner", "Lead", "owner", "lead") {
				projectlist = append(projectlist, ur.ScopeID)
			}
		}
		for _, app := range myapps.List {
			found := false
			for _, proj := range projectlist {
				if proj == int64(app.ProjectID) {
					found = true
					break
				}
			}
			if found {
				myapps_ = append(myapps_, app)
			}
		}
	} else {
		myapps_ = myapps.List
	}
	allruntimeids := []uint64{}
	for _, app := range myapps_ {
		runtimes, err := d.db.FindRuntimesByAppId(app.ID)
		if err != nil {
			return nil, apierrors.ErrListDeployment.InternalError(err)
		}
		for _, r := range runtimes {
			allruntimeids = append(allruntimeids, r.ID)
		}
	}
	filter := dbclient.DeploymentFilter{
		StatusIn:     nil,
		NeedApproved: needApproval,
		Approved:     approved,
	}

	if approvedBy != nil {
		approvedBy_aux := string(*approvedBy)
		filter.ApprovedByUser = &approvedBy_aux
	}
	if len(operateUsers) > 0 {
		filter.OperateUsers = operateUsers
	}
	if len(types) > 0 {
		filter.Types = types
	}
	if len(ids) > 0 {
		filter.IDs = ids
	}
	if approvalStatus != nil {
		filter.ApprovalStatus = approvalStatus
	}
	deployments, total, err := d.db.FindMultiRuntimesDeployments(allruntimeids, filter, page.GetOffset(), page.GetLimit())
	if err != nil {
		return nil, apierrors.ErrListDeployment.InternalError(err)
	}

	list := make([]*apistructs.DeploymentWithDetail, 0, len(deployments))
	for i := range deployments {
		dd := apistructs.DeploymentWithDetail{}
		deploy := deployments[i].Convert()
		dd.Deployment = *deploy

		runtime, err := d.db.GetRuntime(deployments[i].RuntimeId)
		if err != nil {
			return nil, apierrors.ErrListDeployment.InternalError(err)
		}
		dd.RuntimeName = runtime.Name
		dd.BranchName = runtime.GitBranch
		app, err := d.bdl.GetApp(runtime.ApplicationID)
		if err != nil {
			return nil, apierrors.ErrListDeployment.InternalError(err)
		}
		dd.ApplicationID = app.ID
		dd.ApplicationName = app.Name
		proj, err := d.bdl.GetProject(app.ProjectID)
		if err != nil {
			return nil, apierrors.ErrListDeployment.InternalError(err)
		}
		dd.ProjectID = proj.ID
		dd.ProjectName = proj.Name

		release, err := d.bdl.GetRelease(deployments[i].ReleaseId)
		if err != nil {
			// release 可能已经过期清理，忽略该条
			continue
		}
		dd.CommitID = release.Labels["gitCommitId"]
		dd.CommitMessage = release.Labels["gitCommitMessage"]
		list = append(list, &dd)
	}
	data := apistructs.DeploymentDetailListData{
		Total: total,
		List:  list,
	}
	return &data, nil
}

// List 查询部署记录列表
func (d *Deployment) List(userID user.ID, orgID uint64, runtimeID uint64, statuses []string, page apistructs.PageInfo) (
	*apistructs.DeploymentListData, error) {
	runtime, err := d.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrListDeployment.InternalError(err)
	}
	perm, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, apierrors.ErrListDeployment.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrListDeployment.AccessDenied()
	}

	filter := dbclient.DeploymentFilter{StatusIn: statuses}
	deployments, _, err := d.db.FindDeployments(runtimeID, filter, page.GetOffset(), page.GetLimit())
	if err != nil {
		return nil, apierrors.ErrListDeployment.InternalError(err)
	}
	list := make([]*apistructs.Deployment, 0, len(deployments))
	for i := range deployments {
		list = append(list, deployments[i].Convert())
	}
	data := apistructs.DeploymentListData{
		Total: len(list),
		List:  list,
	}
	return &data, nil
}

// GetStatus 查询部署状态
func (d *Deployment) GetStatus(deploymentID uint64) (*apistructs.DeploymentStatusDTO, error) {
	deployment, err := d.db.GetDeployment(deploymentID)
	if err != nil {
		return nil, apierrors.ErrGetDeployment.InternalError(err)
	}
	// TODO: not returns runtime info later
	runtime, err := d.db.GetRuntime(deployment.RuntimeId)
	if err != nil {
		return nil, apierrors.ErrGetDeployment.InternalError(err)
	}
	statusMap := map[string]string{}

	if deployment.NeedApproval {
		statusMap["approval_status"] = deployment.ApprovalStatus
		statusMap["approval_reason"] = deployment.ApprovalReason
		switch deployment.Status {
		case apistructs.DeploymentStatusCanceled:
			statusMap["approval_status"] = "Canceled"
		}
	}

	sg, err := d.bdl.InspectServiceGroupWithTimeout(runtime.ScheduleName.Args())

	var rt *apistructs.DeploymentStatusRuntimeDTO
	if deployment.Status == apistructs.DeploymentStatusOK {
		if err != nil {
			// TODO: we should not treat NOT_FOUND err as DEPLOYING
			// TODO: because scheduler return NOT_FOUND as an err
			return &apistructs.DeploymentStatusDTO{
				DeploymentID: deploymentID,
				Status:       apistructs.DeploymentStatusOK,
			}, nil
		}
		rt = convertDeploymentRuntimeDTO(sg)
	} else {
		if sg != nil {
			if sg.Status != apistructs.StatusReady && sg.Status != apistructs.StatusHealthy {
				for _, serviceItem := range sg.Services {
					statusMap[fmt.Sprintf("Error.Msg.%s", serviceItem.Name)] = serviceItem.LastMessage
					statusMap[fmt.Sprintf("Error.Reason.%s", serviceItem.Name)] = serviceItem.Reason
				}
				cc, _ := json.Marshal(statusMap)
				logrus.Errorf("InspectServiceGroupWithTimeout statusMap: %v", string(cc))
			}
		}
	}
	return &apistructs.DeploymentStatusDTO{
		DeploymentID: deploymentID,
		Status:       deployment.Status,
		Phase:        deployment.Phase,
		FailCause:    deployment.FailCause,
		ModuleErrMsg: statusMap,
		Runtime:      rt,
	}, nil
}

func (d *Deployment) Approve(userID user.ID, orgID uint64, deploymentID uint64, reject bool, reason string, referer string) error {
	deployment, err := d.db.GetDeployment(deploymentID)
	if err != nil {
		return apierrors.ErrApproveDeployment.InternalError(err)
	}
	if deployment.Status == apistructs.DeploymentStatusCanceled {
		return apierrors.ErrApproveDeployment.InvalidParameter(fmt.Errorf("该部署(%d)已被取消", deploymentID))
	}
	if deployment.Status != apistructs.DeploymentStatusWaitApprove {
		return apierrors.ErrApproveDeployment.InvalidParameter(fmt.Errorf("该部署(%d)已被审批过", deploymentID))
	}
	if !deployment.NeedApproval {
		return apierrors.ErrApproveDeployment.InvalidParameter(fmt.Errorf("deployment(%d) NeedApproval = false ", deploymentID))
	}
	if deployment.ApprovedAt != nil {
		return apierrors.ErrApproveDeployment.InvalidParameter(fmt.Errorf("该部署(%d)已被审批过", deploymentID))
	}
	if utf8.RuneCountInString(reason) > 100 {
		return apierrors.ErrApproveDeployment.InvalidParameter(fmt.Errorf("拒绝理由过长"))
	}

	deployment.Status = apistructs.DeploymentStatusWaiting
	now := time.Now()
	deployment.ApprovedAt = &now
	deployment.ApprovedByUser = userID.String()
	deployment.ApprovalReason = reason
	deployment.ApprovalStatus = "Accept"
	if reject {
		deployment.ApprovalStatus = "Reject"
		deployment.Status = apistructs.DeploymentStatusFailed
	}
	if err := d.db.UpdateDeployment(deployment); err != nil {
		return apierrors.ErrApproveDeployment.InternalError(err)
	}
	for range []int{0} {
		runtime, err := d.db.GetRuntime(deployment.RuntimeId)
		if err != nil {
			logrus.Errorf("failed to get runtime(%d): %v", deployment.RuntimeId, err)
			break
		}
		proj, err := d.bdl.GetProject(runtime.ProjectID)
		if err != nil {
			logrus.Errorf("failed to get project(%d): %v", proj.ID, err)
			break
		}
		app, err := d.bdl.GetApp(runtime.ApplicationID)
		if err != nil {
			logrus.Errorf("failed to get app(%d): %v", app.ID, err)
			break
		}
		user, err := d.bdl.ListUsers(apistructs.UserListRequest{
			UserIDs: []string{deployment.Operator},
		})
		if err != nil || len(user.Users) != 1 {
			logrus.Errorf("failed to get user(%s): %v", deployment.Operator, err)
			break
		}
		u, err := url.Parse(referer)
		if err != nil {
			logrus.Errorf("failed to parse referer(%s): %v", referer, err)
			break
		}

		url := fmt.Sprintf("%s://%s/workBench/approval/my-initiate/%s?id=%d",
			u.Scheme, u.Host, deployment.ApprovalStatus, deployment.ID)
		if err := d.bdl.CreateMboxNotify("notify.deployapproval.done.markdown_template",
			map[string]string{
				"title":       fmt.Sprintf("【通知】%s项目%s应用部署审核完成", proj.Name, app.Name),
				"projectName": proj.Name,
				"appName":     app.Name,
				"url":         url,
			}, "zh-CN", proj.OrgID, []string{user.Users[0].ID}); err != nil {
			logrus.Errorf("failed to CreateMboxNotify: %v", err)
		}
		if err := d.bdl.CreateEmailNotify("notify.deployapproval.done.markdown_template",
			map[string]string{
				"title":       fmt.Sprintf("【通知】%s项目%s应用部署审核完成", proj.Name, app.Name),
				"projectName": proj.Name,
				"appName":     app.Name,
				"url":         url,
			}, "zh-CN", proj.OrgID, []string{user.Users[0].Email}); err != nil {
			logrus.Errorf("failed to CreateEmailNotify: %v", err)
		}
	}
	return nil
}

// Deprecated
func convertDeploymentRuntimeDTO(group *apistructs.ServiceGroup) *apistructs.DeploymentStatusRuntimeDTO {
	var runtimeDTO apistructs.DeploymentStatusRuntimeDTO
	runtimeDTO.Endpoints = make(map[string]*apistructs.DeploymentStatusRuntimeServiceDTO)
	runtimeDTO.Services = make(map[string]*apistructs.DeploymentStatusRuntimeServiceDTO)
	for _, service := range group.Services {
		serviceDTO := apistructs.DeploymentStatusRuntimeServiceDTO{
			Host:  service.Vip,
			Ports: diceyml.ComposeIntPortsFromServicePorts(service.Ports),
		}
		vHosts := service.Labels["HAPROXY_0_VHOST"]
		if len(vHosts) > 0 {
			serviceDTO.PublicHosts = strings.Split(vHosts, ",")
			runtimeDTO.Endpoints[service.Name] = &serviceDTO
		} else {
			runtimeDTO.Services[service.Name] = &serviceDTO
		}
	}
	return &runtimeDTO
}
