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

// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gorilla/schema"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/services/activity"
	"github.com/erda-project/erda/modules/cmdb/services/appcertificate"
	"github.com/erda-project/erda/modules/cmdb/services/application"
	"github.com/erda-project/erda/modules/cmdb/services/approve"
	"github.com/erda-project/erda/modules/cmdb/services/audit"
	"github.com/erda-project/erda/modules/cmdb/services/branchrule"
	"github.com/erda-project/erda/modules/cmdb/services/certificate"
	"github.com/erda-project/erda/modules/cmdb/services/cloudaccount"
	"github.com/erda-project/erda/modules/cmdb/services/cluster"
	"github.com/erda-project/erda/modules/cmdb/services/comment"
	"github.com/erda-project/erda/modules/cmdb/services/container"
	"github.com/erda-project/erda/modules/cmdb/services/environment"
	"github.com/erda-project/erda/modules/cmdb/services/errorbox"
	"github.com/erda-project/erda/modules/cmdb/services/filesvc"
	"github.com/erda-project/erda/modules/cmdb/services/filetree"
	"github.com/erda-project/erda/modules/cmdb/services/host"
	"github.com/erda-project/erda/modules/cmdb/services/issue"
	"github.com/erda-project/erda/modules/cmdb/services/issuepanel"
	"github.com/erda-project/erda/modules/cmdb/services/issueproperty"
	"github.com/erda-project/erda/modules/cmdb/services/issuerelated"
	"github.com/erda-project/erda/modules/cmdb/services/issuestate"
	"github.com/erda-project/erda/modules/cmdb/services/issuestream"
	"github.com/erda-project/erda/modules/cmdb/services/iteration"
	"github.com/erda-project/erda/modules/cmdb/services/label"
	"github.com/erda-project/erda/modules/cmdb/services/libreference"
	"github.com/erda-project/erda/modules/cmdb/services/manual_review"
	"github.com/erda-project/erda/modules/cmdb/services/mbox"
	"github.com/erda-project/erda/modules/cmdb/services/member"
	"github.com/erda-project/erda/modules/cmdb/services/namespace"
	"github.com/erda-project/erda/modules/cmdb/services/notice"
	"github.com/erda-project/erda/modules/cmdb/services/notify"
	"github.com/erda-project/erda/modules/cmdb/services/org"
	"github.com/erda-project/erda/modules/cmdb/services/permission"
	"github.com/erda-project/erda/modules/cmdb/services/project"
	"github.com/erda-project/erda/modules/cmdb/services/publisher"
	"github.com/erda-project/erda/modules/cmdb/services/ticket"
	"github.com/erda-project/erda/modules/cmdb/utils"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/license"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	store              jsonstore.JsonStore
	etcdStore          *etcd.Store
	ossClient          *oss.Client
	db                 *dao.DBClient
	uc                 *utils.UCClient
	bdl                *bundle.Bundle
	org                *org.Org
	cloudaccount       *cloudaccount.CloudAccount
	project            *project.Project
	publisher          *publisher.Publisher
	certificate        *certificate.Certificate
	approve            *approve.Approve
	appCertificate     *appcertificate.AppCertificate
	app                *application.Application
	member             *member.Member
	ManualReview       *manual_review.ManualReview
	ticket             *ticket.Ticket
	comment            *comment.Comment
	activity           *activity.Activity
	permission         *permission.Permission
	host               *host.Host
	container          *container.Container
	cluster            *cluster.Cluster
	namespace          *namespace.Namespace
	envConfig          *environment.EnvConfig
	license            *license.License
	notifyGroup        *notify.NotifyGroup
	mbox               *mbox.MBox
	label              *label.Label
	branchRule         *branchrule.BranchRule
	iteration          *iteration.Iteration
	issue              *issue.Issue
	issueStream        *issuestream.IssueStream
	issueRelated       *issuerelated.IssueRelated
	issueProperty      *issueproperty.IssueProperty
	issueState         *issuestate.IssueState
	issuePanel         *issuepanel.IssuePanel
	notice             *notice.Notice
	libReference       *libreference.LibReference
	fileSvc            *filesvc.FileService
	queryStringDecoder *schema.Decoder
	audit              *audit.Audit
	errorbox           *errorbox.ErrorBox
	fileTree           *filetree.FileTree
}

type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

// WithOSSClient 配置OSS Client
func WithOSSClient(client *oss.Client) Option {
	return func(e *Endpoints) {
		e.ossClient = client
	}
}

// WithDBClient 配置 db
func WithDBClient(db *dao.DBClient) Option {
	return func(e *Endpoints) {
		e.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// WithUCClient 配置 UC Client
func WithUCClient(uc *utils.UCClient) Option {
	return func(e *Endpoints) {
		e.uc = uc
	}
}

// WithJSONStore 配置 jsonstore
func WithJSONStore(store jsonstore.JsonStore) Option {
	return func(e *Endpoints) {
		e.store = store
	}
}

// WithEtcdStore 配置 etcdStore
func WithEtcdStore(etcdStore *etcd.Store) Option {
	return func(e *Endpoints) {
		e.etcdStore = etcdStore
	}
}

// WithOrg 配置 org service
func WithOrg(org *org.Org) Option {
	return func(e *Endpoints) {
		e.org = org
	}
}

// WithCloudAccount 配置 cloudaccount service
func WithCloudAccount(account *cloudaccount.CloudAccount) Option {
	return func(e *Endpoints) {
		e.cloudaccount = account
	}
}

// WithProject 配置 project service
func WithProject(project *project.Project) Option {
	return func(e *Endpoints) {
		e.project = project
	}
}

// WithPublisher 配置 publisher service
func WithPublisher(pub *publisher.Publisher) Option {
	return func(e *Endpoints) {
		e.publisher = pub
	}
}

// WithCertificate 配置证书 service
func WithCertificate(cer *certificate.Certificate) Option {
	return func(e *Endpoints) {
		e.certificate = cer
	}
}

// WithAppCertificate 配置证书 service
func WithAppCertificate(cer *appcertificate.AppCertificate) Option {
	return func(e *Endpoints) {
		e.appCertificate = cer
	}
}

// WithApp 配置 app service
func WithApp(app *application.Application) Option {
	return func(e *Endpoints) {
		e.app = app
	}
}

// WithMember 配置 member service
func WithMember(member *member.Member) Option {
	return func(e *Endpoints) {
		e.member = member
	}
}

// WithManualReview 配置 ManualReview service
func WithManualReview(ManualReview *manual_review.ManualReview) Option {
	return func(e *Endpoints) {
		e.ManualReview = ManualReview
	}
}

// WithTicket 配置 ticket service
func WithTicket(ticket *ticket.Ticket) Option {
	return func(e *Endpoints) {
		e.ticket = ticket
	}
}

// WithComment 配置 comment service
func WithComment(comment *comment.Comment) Option {
	return func(e *Endpoints) {
		e.comment = comment
	}
}

// WithActivity 配置 activity service
func WithActivity(activity *activity.Activity) Option {
	return func(e *Endpoints) {
		e.activity = activity
	}
}

// WithPermission 配置 permission service
func WithPermission(permission *permission.Permission) Option {
	return func(e *Endpoints) {
		e.permission = permission
	}
}

// WithHost 配置 host service
func WithHost(host *host.Host) Option {
	return func(e *Endpoints) {
		e.host = host
	}
}

// WithContainer 配置 container service
func WithContainer(container *container.Container) Option {
	return func(e *Endpoints) {
		e.container = container
	}
}

// WithCluster 配置 cluster service
func WithCluster(cluster *cluster.Cluster) Option {
	return func(e *Endpoints) {
		e.cluster = cluster
	}
}

// WithNamespace 配置 namespace service
func WithNamespace(namespace *namespace.Namespace) Option {
	return func(e *Endpoints) {
		e.namespace = namespace
	}
}

// WithNotify 配置 notify group service
func WithNotify(notifyGroup *notify.NotifyGroup) Option {
	return func(e *Endpoints) {
		e.notifyGroup = notifyGroup
	}
}

// WithMBox 配置 mbox service
func WithMBox(mbox *mbox.MBox) Option {
	return func(e *Endpoints) {
		e.mbox = mbox
	}
}

// WithEnvConfig 配置 env config
func WithEnvConfig(envConfig *environment.EnvConfig) Option {
	return func(e *Endpoints) {
		e.envConfig = envConfig
	}
}

// WithLicense 配置 license
func WithLicense(license *license.License) Option {
	return func(e *Endpoints) {
		e.license = license
	}
}

// WithLabel 配置 label
func WithLabel(label *label.Label) Option {
	return func(e *Endpoints) {
		e.label = label
	}
}

// WithIteration 配置 iteration
func WithIteration(itr *iteration.Iteration) Option {
	return func(e *Endpoints) {
		e.iteration = itr
	}
}

// WithIssue 配置 issue
func WithIssue(issue *issue.Issue) Option {
	return func(e *Endpoints) {
		e.issue = issue
	}
}

func WithIssueRelated(ir *issuerelated.IssueRelated) Option {
	return func(e *Endpoints) {
		e.issueRelated = ir
	}
}

func WithIssueState(state *issuestate.IssueState) Option {
	return func(e *Endpoints) {
		e.issueState = state
	}
}

func WithIssuePanel(panel *issuepanel.IssuePanel) Option {
	return func(e *Endpoints) {
		e.issuePanel = panel
	}
}

// WithIssue 配置 issueStream
func WithIssueStream(stream *issuestream.IssueStream) Option {
	return func(e *Endpoints) {
		e.issueStream = stream
	}
}

// WithIssue 配置 issueStream
func WithIssueProperty(property *issueproperty.IssueProperty) Option {
	return func(e *Endpoints) {
		e.issueProperty = property
	}
}

// WithNotice 设置 notice service
func WithNotice(notice *notice.Notice) Option {
	return func(e *Endpoints) {
		e.notice = notice
	}
}

// WithLibReference 设置 libReference service
func WithLibReference(libReference *libreference.LibReference) Option {
	return func(e *Endpoints) {
		e.libReference = libReference
	}
}

func WithFileSvc(svc *filesvc.FileService) Option {
	return func(e *Endpoints) {
		e.fileSvc = svc
	}
}

func WithApprove(approve *approve.Approve) Option {
	return func(e *Endpoints) {
		e.approve = approve
	}
}

// WithQueryStringDecoder 配置 queryStringDecoder
func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

// WithBranchRule 配置 branchRule
func WithBranchRule(branchRule *branchrule.BranchRule) Option {
	return func(e *Endpoints) {
		e.branchRule = branchRule
	}
}

// WithAudit 配置 Audit
func WithAudit(audit *audit.Audit) Option {
	return func(e *Endpoints) {
		e.audit = audit
	}
}

func WithErrorBox(errorbox *errorbox.ErrorBox) Option {
	return func(e *Endpoints) {
		e.errorbox = errorbox
	}
}

func WithGittarFileTree(fileTree *filetree.FileTree) Option {
	return func(e *Endpoints) {
		e.fileTree = fileTree
	}
}

// DBClient 获取db client
func (e *Endpoints) DBClient() *dao.DBClient {
	return e.db
}

// GetLocale 获取本地化资源
func (e *Endpoints) GetLocale(request *http.Request) *i18n.LocaleResource {
	return e.bdl.GetLocaleByRequest(request)
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/info", Method: http.MethodGet, Handler: e.Info},

		// hosts
		{Path: "/api/hosts/{host}", Method: http.MethodGet, Handler: e.GetHost},

		// 仅供监控使用，不在 openapi 暴露
		{Path: "/api/containers/actions/list-edas", Method: http.MethodGet, Handler: e.ListEdasContainers},

		// 集群相关
		{Path: "/api/clusters", Method: http.MethodPost, Handler: e.CreateCluster},
		{Path: "/api/clusters", Method: http.MethodPut, Handler: e.UpdateCluster},
		{Path: "/api/clusters/{idOrName}", Method: http.MethodGet, Handler: e.GetCluster},
		{Path: "/api/clusters", Method: http.MethodGet, Handler: e.ListCluster},
		{Path: "/api/clusters/{clusterName}", Method: http.MethodDelete, Handler: e.DeleteCluster},
		{Path: "/api/clusters/actions/dereference", Method: http.MethodPut, Handler: e.DereferenceCluster},

		{Path: "/api/org/actions/list-running-tasks", Method: http.MethodGet, Handler: e.ListOrgRunningTasks},
		{Path: "/api/tasks", Method: http.MethodPost, Handler: e.DealTaskEvent},

		// webhook
		{Path: "/api/events/instance-status", Method: http.MethodPost, Handler: e.UpdateInstanceBySchedulerEvent},

		// 企业相关
		{Path: "/api/orgs", Method: http.MethodPost, Handler: e.CreateOrg},
		{Path: "/api/orgs/{orgID}", Method: http.MethodPut, Handler: e.UpdateOrg},
		{Path: "/api/orgs/{idOrName}", Method: http.MethodGet, Handler: e.GetOrg},
		{Path: "/api/orgs/{idOrName}", Method: http.MethodDelete, Handler: e.DeleteOrg},
		{Path: "/api/orgs", Method: http.MethodGet, Handler: e.ListOrg},
		{Path: "/api/orgs/actions/list-public", Method: http.MethodGet, Handler: e.ListPublicOrg},
		{Path: "/api/orgs/ingress/{orgID}/actions/update-ingress", Method: http.MethodGet, Handler: e.UpdateOrgIngress},
		{Path: "/api/orgs/actions/get-by-domain", Method: http.MethodGet, Handler: e.GetOrgByDomain},
		{Path: "/api/orgs/actions/switch", Method: http.MethodPost, Handler: e.ChangeCurrentOrg},
		{Path: "/api/orgs/actions/relate-cluster", Method: http.MethodPost, Handler: e.CreateOrgClusterRelation},
		{Path: "/api/orgs/clusters/relations", Method: http.MethodGet, Handler: e.ListAllOrgClusterRelation},
		{Path: "/api/orgs/{orgID}/nexus", Method: http.MethodGet, Handler: e.GetOrgNexus},
		{Path: "/api/orgs/{orgID}/actions/show-nexus-password", Method: http.MethodGet, Handler: e.ShowOrgNexusPassword},
		{Path: "/api/orgs/{orgID}/actions/create-publisher", Method: http.MethodPost, Handler: e.CreateOrgPublisher},
		{Path: "/api/orgs/{orgID}/actions/create-publisher", Method: http.MethodGet, Handler: e.CreateOrgPublisher},
		{Path: "/api/orgs/{orgID}/actions/set-release-cross-cluster", Method: http.MethodPost, Handler: e.SetReleaseCrossCluster},
		{Path: "/api/orgs/{orgID}/actions/get-nexus-docker-credential-by-image", Method: http.MethodGet, Handler: e.GetNexusOrgDockerCredentialByImage},
		{Path: "/api/orgs/actions/gen-verify-code", Method: http.MethodPost, Handler: e.GenVerifiCode},
		{Path: "/api/orgs/{orgID}/actions/set-notify-config", Method: http.MethodPost, Handler: e.SetNotifyConfig},
		{Path: "/api/orgs/{orgID}/actions/get-notify-config", Method: http.MethodGet, Handler: e.GetNotifyConfig},

		// 获取企业可用资源
		{Path: "/api/orgs/actions/fetch-resources", Method: http.MethodGet, Handler: e.FetchOrgResources},

		// 云账号相关
		{Path: "/api/cloud-accounts", Method: http.MethodPost, Handler: e.CreateCloudAccount},
		{Path: "/api/cloud-accounts", Method: http.MethodGet, Handler: e.ListCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodGet, Handler: e.GetCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodPut, Handler: e.UpdateCloudAccount},
		{Path: "/api/cloud-accounts/{accountID}", Method: http.MethodDelete, Handler: e.DeleteCloudAccount},

		// 平台公告
		{Path: "/api/notices", Method: http.MethodPost, Handler: e.CreateNotice},
		{Path: "/api/notices/{id}", Method: http.MethodPut, Handler: e.UpdateNotice},
		{Path: "/api/notices/{id}/actions/publish", Method: http.MethodPut, Handler: e.PublishNotice},
		{Path: "/api/notices/{id}/actions/unpublish", Method: http.MethodPut, Handler: e.UnpublishNotice},
		{Path: "/api/notices/{id}", Method: http.MethodDelete, Handler: e.DeleteNotice},
		{Path: "/api/notices", Method: http.MethodGet, Handler: e.ListNotice},

		// 项目相关
		{Path: "/api/projects", Method: http.MethodPost, Handler: e.CreateProject},
		{Path: "/api/projects/{projectID}", Method: http.MethodPut, Handler: e.UpdateProject},
		{Path: "/api/projects/{projectID}", Method: http.MethodGet, Handler: e.GetProject},
		{Path: "/api/projects/{projectID}", Method: http.MethodDelete, Handler: e.DeleteProject},
		{Path: "/api/projects", Method: http.MethodGet, Handler: e.ListProject},
		{Path: "/api/projects/resource/{resourceType}/actions/list-usage-histogram", Method: http.MethodGet, Handler: e.ListProjectResourceUsage},
		{Path: "/api/projects/actions/list-my-projects", Method: http.MethodGet, Handler: e.ListMyProject},
		{Path: "/api/projects/actions/list-public-projects", Method: http.MethodGet, Handler: e.ListPublicProject},
		{Path: "/api/projects/actions/refer-cluster", Method: http.MethodGet, Handler: e.ReferCluster},
		{Path: "/api/projects/actions/fill-branch-rule", Method: http.MethodGet, Handler: e.FillBranchRule},
		{Path: "/api/projects/actions/get-project-functions", Method: http.MethodGet, Handler: e.GetFunctions},
		{Path: "/api/projects/actions/set-project-functions", Method: http.MethodPost, Handler: e.SetFunctions},
		{Path: "/api/projects/actions/update-active-time", Method: http.MethodPut, Handler: e.UpdateProjectActiveTime},

		{Path: "/api/branch-rules", Method: http.MethodPost, Handler: e.CreateBranchRule},
		{Path: "/api/branch-rules", Method: http.MethodGet, Handler: e.QueryBranchRules},
		{Path: "/api/branch-rules/{id}", Method: http.MethodPut, Handler: e.UpdateBranchRule},
		{Path: "/api/branch-rules/{id}", Method: http.MethodDelete, Handler: e.DeleteBranchRule},
		{Path: "/api/branch-rules/actions/app-all-valid-branch-workspaces", Method: http.MethodGet, Handler: e.GetAllValidBranchWorkspaces},

		{Path: "/api/projects/{projectID}/actions/get-ns-info", Method: http.MethodGet, Handler: e.GetNSInfo},

		// 应用相关
		{Path: "/api/applications", Method: http.MethodPost, Handler: e.CreateApplication},
		{Path: "/api/applications/{applicationID}/actions/init", Method: http.MethodPut, Handler: e.InitApplication},
		{Path: "/api/applications/{applicationID}", Method: http.MethodPut, Handler: e.UpdateApplication},
		{Path: "/api/applications/{applicationID}", Method: http.MethodGet, Handler: e.GetApplication},
		{Path: "/api/applications/{applicationID}", Method: http.MethodDelete, Handler: e.DeleteApplication},
		{Path: "/api/applications", Method: http.MethodGet, Handler: e.ListApplication},
		{Path: "/api/applications/actions/list-my-applications", Method: http.MethodGet, Handler: e.ListMyApplication},
		{Path: "/api/applications/actions/remove-publish-item-relations", Method: http.MethodPost, Handler: e.RemoveApplicationPublishItemRelations},
		{Path: "/api/applications/{applicationID}/actions/get-publish-item-relations", Method: http.MethodGet, Handler: e.GetApplicationPublishItemRelationsGroupByENV},
		{Path: "/api/applications/actions/query-publish-item-relations", Method: http.MethodGet, Handler: e.QueryApplicationPublishItemRelations},
		{Path: "/api/applications/{applicationID}/actions/update-publish-item-relations", Method: http.MethodPost, Handler: e.UpdateApplicationPublishItemRelations},

		{Path: "/api/applications/{applicationID}/actions/pin", Method: http.MethodPut, Handler: e.PinApplication},
		{Path: "/api/applications/{applicationID}/actions/unpin", Method: http.MethodPut, Handler: e.UnPinApplication},
		{Path: "/api/applications/actions/prepare-ability-app", Method: http.MethodPost, Handler: e.PrepareAbilityApp},
		{Path: "/api/applications/actions/list-templates", Method: http.MethodGet, Handler: e.ListAppTemplates},

		// 成员相关
		{Path: "/api/members", Method: http.MethodPost, Handler: e.CreateOrUpdateMember},
		{Path: "/api/members/actions/remove", Method: http.MethodPost, Handler: e.DeleteMember},
		{Path: "/api/members/actions/destroy", Method: http.MethodPost, Handler: e.DestroyMember},
		{Path: "/api/members", Method: http.MethodGet, Handler: e.ListMember},
		{Path: "/api/members/actions/get-by-token", Method: http.MethodGet, Handler: e.GetMemberByToken},
		{Path: "/api/members/actions/list-roles", Method: http.MethodGet, Handler: e.ListMemberRoles},
		{Path: "/api/members/actions/list-user-roles", Method: http.MethodGet, Handler: e.ListMemberRolesByUser},
		{Path: "/api/members/actions/get-all-organizational", Method: http.MethodGet, Handler: e.GetAllOrganizational},
		{Path: "/api/members/actions/update-userinfo", Method: http.MethodPut, Handler: e.UpdateMemberUserInfo},
		{Path: "/api/members/actions/create-by-invitecode", Method: http.MethodPost, Handler: e.CreateMemberByInviteCode},
		// 成员标签
		{Path: "/api/members/actions/list-labels", Method: http.MethodGet, Handler: e.ListMeberLabels},

		// 鉴权相关
		{Path: "/api/permissions", Method: http.MethodGet, Handler: e.ListScopeRole},
		{Path: "/api/permissions/actions/access", Method: http.MethodPost, Handler: e.ScopeRoleAccess},
		{Path: "/api/permissions/actions/check", Method: http.MethodPost, Handler: e.CheckPermission},

		// 工单相关
		{Path: "/api/tickets", Method: http.MethodPost, Handler: e.CreateTicket},
		{Path: "/api/tickets/{ticketID}", Method: http.MethodPut, Handler: e.UpdateTicket},
		{Path: "/api/tickets/{ticketID}/actions/close", Method: http.MethodPut, Handler: e.CloseTicket},
		{Path: "/api/tickets/actions/close-by-key", Method: http.MethodPut, Handler: e.CloseTicketByKey},
		{Path: "/api/tickets/{ticketID}/actions/reopen", Method: http.MethodPut, Handler: e.ReopenTicket},
		{Path: "/api/tickets/{ticketID}", Method: http.MethodGet, Handler: e.GetTicket},
		{Path: "/api/tickets", Method: http.MethodGet, Handler: e.ListTicket},
		{Path: "/api/tickets/actions/batch-delete", Method: http.MethodDelete, Handler: e.DeleteTicket},

		// 工单评论相关
		{Path: "/api/comments", Method: http.MethodPost, Handler: e.CreateComment},
		{Path: "/api/comments/{commentID}", Method: http.MethodPut, Handler: e.UpdateComment},
		{Path: "/api/comments", Method: http.MethodGet, Handler: e.ListComments},
		{Path: "/api/comments/{commentID}", Method: http.MethodDelete, Handler: e.DeleteComment},

		// 用户相关
		{Path: "/api/users", Method: http.MethodGet, Handler: e.ListUser},
		{Path: "/api/users/current", Method: http.MethodGet, Handler: e.GetCurrentUser},

		// 活动相关
		{Path: "/api/activities", Method: http.MethodGet, Handler: e.ListActivity},

		// 配置管理相关
		{Path: "/api/config/namespace", Method: http.MethodPost, Handler: e.CreateNamespace},
		{Path: "/api/config/namespace", Method: http.MethodDelete, Handler: e.DeleteNamespace},
		{Path: "/api/config/namespace/relation", Method: http.MethodPost, Handler: e.CreateNamespaceRelation},
		{Path: "/api/config/namespace/relation", Method: http.MethodDelete, Handler: e.DeleteNamespaceRelation},
		{Path: "/api/config", Method: http.MethodPost, Handler: e.AddConfigs},
		{Path: "/api/config", Method: http.MethodGet, Handler: e.GetConfigs},
		{Path: "/api/config", Method: http.MethodPut, Handler: e.UpdateConfigs},
		{Path: "/api/config", Method: http.MethodDelete, Handler: e.DeleteConfig},
		{Path: "/api/config/actions/export", Method: http.MethodGet, Handler: e.ExportConfigs},
		{Path: "/api/config/actions/import", Method: http.MethodPost, Handler: e.ImportConfigs},
		{Path: "/api/config/deployment", Method: http.MethodGet, Handler: e.GetDeployConfigs},
		//{"/api/configmanage/configs/publish",Method:http.MethodPost,Handler: e.PublishConfig},
		//{"/api/configmanage/configs/publish/all",Method:http.MethodPost,Handler: e.PublishConfigs},
		{Path: "/api/config/actions/list-multinamespace-configs", Method: http.MethodPost, Handler: e.GetMultiNamespaceConfigs},
		// 以前的dice_config_namespace表数据不全，里面很多name没有了，导致check ns exist时报错，用这个接口修复
		{Path: "/api/config/namespace/fix-namespace-data-err", Method: http.MethodGet, Handler: e.FixDataErr},

		// 通知组相关
		{Path: "/api/notify-groups", Method: http.MethodPost, Handler: e.CreateNotifyGroup},
		{Path: "/api/notify-groups", Method: http.MethodGet, Handler: e.QueryNotifyGroup},
		{Path: "/api/notify-groups/{groupID}", Method: http.MethodGet, Handler: e.GetNotifyGroup},
		{Path: "/api/notify-groups/{groupID}", Method: http.MethodPut, Handler: e.UpdateNotifyGroup},
		{Path: "/api/notify-groups/{groupID}/detail", Method: http.MethodGet, Handler: e.GetNotifyGroupDetail},
		{Path: "/api/notify-groups/{groupID}", Method: http.MethodDelete, Handler: e.DeleteNotifyGroup},
		{Path: "/api/notifies", Method: http.MethodPost, Handler: e.CreateNotify},
		{Path: "/api/notifies", Method: http.MethodGet, Handler: e.QueryNotifies},
		{Path: "/api/notifies/{notifyID}", Method: http.MethodGet, Handler: e.GetNotify},
		{Path: "/api/notifies/{notifyID}", Method: http.MethodPut, Handler: e.UpdateNotify},
		{Path: "/api/notifies/{notifyID}", Method: http.MethodDelete, Handler: e.DeleteNotify},
		{Path: "/api/notifies/{notifyID}/actions/enable", Method: http.MethodPut, Handler: e.NotifyEnable},
		{Path: "/api/notifies/{notifyID}/actions/disable", Method: http.MethodPut, Handler: e.NotifyDisable},
		{Path: "/api/notify-sources", Method: http.MethodDelete, Handler: e.DeleteNotifySource},
		{Path: "/api/notify-items", Method: http.MethodGet, Handler: e.QueryNotifyItems},
		{Path: "/api/notify-items/{notifyItemID}", Method: http.MethodPut, Handler: e.UpdateNotifyItem},
		{Path: "/api/notify-histories", Method: http.MethodGet, Handler: e.QueryNotifyHistories},
		{Path: "/api/notify-histories", Method: http.MethodPost, Handler: e.CreateNotifyHistory},
		{Path: "/api/notifies/actions/search-by-source", Method: http.MethodGet, Handler: e.QueryNotifiesBySource},
		{Path: "/api/notifies/actions/fuzzy-query-by-source", Method: http.MethodGet, Handler: e.FuzzyQueryNotifiesBySource},
		{Path: "/api/notify-groups/actions/batch-get", Method: http.MethodGet, Handler: e.BatchGetNotifyGroup}, //内部接口

		// license
		{Path: "/api/license", Method: http.MethodGet, Handler: e.GetLicense},

		// 标签相关
		{Path: "/api/labels", Method: http.MethodPost, Handler: e.CreateLabel},
		{Path: "/api/labels/{id}", Method: http.MethodDelete, Handler: e.DeleteLabel},
		{Path: "/api/labels/{id}", Method: http.MethodPut, Handler: e.UpdateLabel},
		{Path: "/api/labels/{id}", Method: http.MethodGet, Handler: e.GetLabel},
		{Path: "/api/labels", Method: http.MethodGet, Handler: e.ListLabel},

		// 站内信相关
		{Path: "/api/mboxs", Method: http.MethodGet, Handler: e.QueryMBox},
		{Path: "/api/mboxs", Method: http.MethodPost, Handler: e.CreateMBox},
		{Path: "/api/mboxs/actions/stats", Method: http.MethodGet, Handler: e.GetMBoxStats},
		{Path: "/api/mboxs/actions/set-read", Method: http.MethodPost, Handler: e.SetMBoxReadStatus},
		{Path: "/api/mboxs/{mboxID}", Method: http.MethodGet, Handler: e.GetMBox},

		// 其他
		{Path: "/api/images/actions/upload", Method: http.MethodPost, Handler: e.UploadImage},

		// 迭代
		{Path: "/api/iterations", Method: http.MethodPost, Handler: e.CreateIteration},
		{Path: "/api/iterations/{id}", Method: http.MethodPut, Handler: e.UpdateIteration},
		{Path: "/api/iterations/{id}", Method: http.MethodDelete, Handler: e.DeleteIteration},
		{Path: "/api/iterations/{id}", Method: http.MethodGet, Handler: e.GetIteration},
		{Path: "/api/iterations", Method: http.MethodGet, Handler: e.PagingIterations},

		// issue 管理
		{Path: "/api/issues", Method: http.MethodPost, Handler: e.CreateIssue},
		{Path: "/api/issues", Method: http.MethodGet, Handler: e.PagingIssues},
		{Path: "/api/issues/{id}", Method: http.MethodGet, Handler: e.GetIssue},
		{Path: "/api/issues/{id}", Method: http.MethodPut, Handler: e.UpdateIssue},
		{Path: "/api/issues/{id}", Method: http.MethodDelete, Handler: e.DeleteIssue},
		{Path: "/api/issues/actions/batch-update", Method: http.MethodPut, Handler: e.BatchUpdateIssue},
		{Path: "/api/issues/actions/export-excel", Method: http.MethodGet, WriterHandler: e.ExportExcelIssue},
		{Path: "/api/issues/actions/import-excel", Method: http.MethodPost, Handler: e.ImportExcelIssue},
		{Path: "/api/issues/actions/man-hour", Method: http.MethodGet, Handler: e.GetIssueManHourSum},
		{Path: "/api/issues/actions/bug-percentage", Method: http.MethodGet, Handler: e.GetIssueBugPercentage},
		{Path: "/api/issues/actions/bug-status-percentage", Method: http.MethodGet, Handler: e.GetIssueBugStatusPercentage},
		{Path: "/api/issues/actions/bug-severity-percentage", Method: http.MethodGet, Handler: e.GetIssueBugSeverityPercentage},
		{Path: "/api/issues/{id}/streams", Method: http.MethodPost, Handler: e.CreateCommentIssueStream},
		{Path: "/api/issues/{id}/streams", Method: http.MethodGet, Handler: e.PagingIssueStreams},
		{Path: "/api/issues/{id}/relations", Method: http.MethodPost, Handler: e.AddIssueRelation},
		{Path: "/api/issues/{id}/relations/{relatedIssueID}", Method: http.MethodDelete, Handler: e.DeleteIssueRelation},
		{Path: "/api/issues/{id}/relations", Method: http.MethodGet, Handler: e.GetIssueRelations},
		{Path: "/api/issues/actions/update-issue-type", Method: http.MethodPut, Handler: e.UpdateIssueType},
		// issue state
		{Path: "/api/issues/actions/create-state", Method: http.MethodPost, Handler: e.CreateIssueState},
		{Path: "/api/issues/actions/delete-state", Method: http.MethodDelete, Handler: e.DeleteIssueState},
		{Path: "/api/issues/actions/update-state-relation", Method: http.MethodPut, Handler: e.UpdateIssueStateRelation},
		{Path: "/api/issues/actions/get-states", Method: http.MethodGet, Handler: e.GetIssueStates},
		{Path: "/api/issues/actions/get-state-relations", Method: http.MethodGet, Handler: e.GetIssueStateRelation},
		{Path: "/api/issues/actions/get-state-belong", Method: http.MethodGet, Handler: e.GetIssueStatesBelong},
		{Path: "/api/issues/actions/get-state-name", Method: http.MethodGet, Handler: e.GetIssueStatesByIDs},
		// issue property
		{Path: "/api/issues/actions/create-property", Method: http.MethodPost, Handler: e.CreateIssueProperty},
		{Path: "/api/issues/actions/delete-property", Method: http.MethodDelete, Handler: e.DeleteIssueProperty},
		{Path: "/api/issues/actions/update-property", Method: http.MethodPut, Handler: e.UpdateIssueProperty},
		{Path: "/api/issues/actions/get-properties", Method: http.MethodGet, Handler: e.GetIssueProperties},
		{Path: "/api/issues/actions/update-properties-index", Method: http.MethodPut, Handler: e.UpdateIssuePropertiesIndex},
		{Path: "/api/issues/actions/get-properties-time", Method: http.MethodGet, Handler: e.GetIssuePropertyUpdateTime},
		// issue panel
		{Path: "/api/issues/actions/create-panel", Method: http.MethodPost, Handler: e.CreateIssuePanel},
		{Path: "/api/issues/actions/delete-panel", Method: http.MethodDelete, Handler: e.DeleteIssuePanel},
		{Path: "/api/issues/actions/update-panel-issue", Method: http.MethodPut, Handler: e.UpdateIssuePanelIssue},
		{Path: "/api/issues/actions/update-panel", Method: http.MethodPut, Handler: e.UpdateIssuePanel},
		{Path: "/api/issues/actions/get-panel", Method: http.MethodGet, Handler: e.GetIssuePanel},
		{Path: "/api/issues/actions/get-panel-issue", Method: http.MethodGet, Handler: e.GetIssuePanelIssue},
		// issue instance
		{Path: "/api/issues/actions/create-property-instance", Method: http.MethodPost, Handler: e.CreateIssuePropertyInstance},
		{Path: "/api/issues/actions/update-property-instance", Method: http.MethodPut, Handler: e.UpdateIssuePropertyInstance},
		{Path: "/api/issues/actions/get-property-instance", Method: http.MethodGet, Handler: e.GetIssuePropertyInstance},
		// issue stage
		{Path: "/api/issues/action/update-stage", Method: http.MethodPut, Handler: e.CreateIssueStage},
		{Path: "/api/issues/action/get-stage", Method: http.MethodGet, Handler: e.GetIssueStage},
		//执行issue的历史数据推送到监控平台
		{Path: "/api/issues/monitor/history", Method: http.MethodGet, Handler: e.RunIssueHistory},
		{Path: "/api/issues/monitor/addOrRepairHistory", Method: http.MethodGet, Handler: e.RunIssueAddOrRepairHistory},

		// 文件服务
		{Path: "/api/files", Method: http.MethodPost, Handler: e.UploadFile},
		{Path: "/api/files", Method: http.MethodGet, WriterHandler: e.DownloadFile},
		{Path: "/api/files/{uuid}", Method: http.MethodGet, WriterHandler: e.DownloadFile},
		{Path: "/api/files/{uuid}", Method: http.MethodHead, WriterHandler: e.HeadFile},
		{Path: "/api/files/{uuid}", Method: http.MethodDelete, Handler: e.DeleteFile},

		// Publisher
		{Path: "/api/publishers", Method: http.MethodPost, Handler: e.CreatePublisher},
		{Path: "/api/publishers", Method: http.MethodPut, Handler: e.UpdatePublisher},
		{Path: "/api/publishers/{publisherID}", Method: http.MethodGet, Handler: e.GetPublisher},
		{Path: "/api/publishers/{publisherID}", Method: http.MethodDelete, Handler: e.DeletePublisher},
		{Path: "/api/publishers", Method: http.MethodGet, Handler: e.ListPublishers},
		{Path: "/api/publishers/actions/list-my-publishers", Method: http.MethodGet, Handler: e.ListMyPublishers},

		// Certificate
		{Path: "/api/certificates", Method: http.MethodPost, Handler: e.CreateCertificate},
		{Path: "/api/certificates/{certificateID}", Method: http.MethodPut, Handler: e.UpdateCertificate},
		{Path: "/api/certificates/{certificateID}", Method: http.MethodGet, Handler: e.GetCertificate},
		{Path: "/api/certificates/{certificateID}", Method: http.MethodDelete, Handler: e.DeleteCertificate},
		{Path: "/api/certificates/actions/list-certificates", Method: http.MethodGet, Handler: e.ListCertificates},

		// Application Certificate
		{Path: "/api/certificates/actions/application-quote", Method: http.MethodPost, Handler: e.QuoteCertificate},
		{Path: "/api/certificates/actions/application-cancel-quote", Method: http.MethodDelete, Handler: e.CancelQuoteCertificate},
		{Path: "/api/certificates/actions/list-application-quotes", Method: http.MethodGet, Handler: e.ListQuoteCertificates},

		// push certificate config
		{Path: "/api/certificates/actions/push-configs", Method: http.MethodPost, Handler: e.PushCertificateConfig},

		{Path: "/api/lib-references", Method: http.MethodPost, Handler: e.CreateLibReference},
		{Path: "/api/lib-references/{id}", Method: http.MethodDelete, Handler: e.DeleteLibReference},
		{Path: "/api/lib-references", Method: http.MethodGet, Handler: e.ListLibReference},
		{Path: "/api/lib-references/actions/fetch-versions", Method: http.MethodGet, Handler: e.ListLibReferenceVersion},

		// approval
		{Path: "/api/approves", Method: http.MethodPost, Handler: e.CreateApprove},
		{Path: "/api/approves/{approveId}", Method: http.MethodPut, Handler: e.UpdateApprove},
		{Path: "/api/approves/{approveId}", Method: http.MethodGet, Handler: e.GetApprove},
		{Path: "/api/approves/actions/list-approves", Method: http.MethodGet, Handler: e.ListApproves},
		{Path: "/api/approvals/actions/watch-status", Method: http.MethodPost, Handler: e.WatchApprovalStatusChanged},

		// 审计事件相关
		{Path: "/api/audits/actions/create", Method: http.MethodPost, Handler: e.CreateAudits},
		{Path: "/api/audits/actions/batch-create", Method: http.MethodPost, Handler: e.BatchCreateAudits},
		{Path: "/api/audits/actions/list", Method: http.MethodGet, Handler: e.ListAudits},
		{Path: "/api/audits/actions/setting", Method: http.MethodPut, Handler: e.PutAuditsSettings},
		{Path: "/api/audits/actions/setting", Method: http.MethodGet, Handler: e.GetAuditsSettings},
		{Path: "/api/audits/actions/export-excel", Method: http.MethodGet, WriterHandler: e.ExportExcelAudit},

		// 错误日志相关
		{Path: "/api/task-error/actions/create", Method: http.MethodPost, Handler: e.CreateOrUpdateErrorLog},
		{Path: "/api/task-error/actions/list", Method: http.MethodGet, Handler: e.ListErrorLog},

		// 流水线filetree查询
		{Path: "/api/project-pipeline/filetree/{inode}/actions/find-ancestors", Method: http.MethodGet, Handler: e.FindFileTreeNodeAncestors},
		{Path: "/api/project-pipeline/filetree", Method: http.MethodPost, Handler: e.CreateFileTreeNode},
		{Path: "/api/project-pipeline/filetree/{inode}", Method: http.MethodDelete, Handler: e.DeleteFileTreeNode},
		{Path: "/api/project-pipeline/filetree", Method: http.MethodGet, Handler: e.ListFileTreeNodes},
		{Path: "/api/project-pipeline/filetree/{inode}", Method: http.MethodGet, Handler: e.GetFileTreeNode},
		{Path: "/api/project-pipeline/filetree/actions/fuzzy-search", Method: http.MethodGet, Handler: e.FuzzySearchFileTreeNodes},

		//人工审核相关
		{Path: "/api/reviews/actions/list-launched-approval", Method: http.MethodGet, Handler: e.GetReviewsBySponsorId},
		{Path: "/api/reviews/actions/list-approved", Method: http.MethodGet, Handler: e.GetReviewsByUserId},
		{Path: "/api/reviews/actions/review/approve", Method: http.MethodPost, Handler: e.CreateReview},
		{Path: "/api/reviews/actions/authority", Method: http.MethodGet, Handler: e.GetAuthorityByUserId},
		{Path: "/api/reviews/actions/updateReview", Method: http.MethodPut, Handler: e.UpdateApproval},
		{Path: "/api/reviews/actions/user/create", Method: http.MethodPost, Handler: e.CreateReviewUser},
		{Path: "/api/reviews/actions/{taskId}", Method: http.MethodGet, Handler: e.GetReviewByTaskId},
	}
}
