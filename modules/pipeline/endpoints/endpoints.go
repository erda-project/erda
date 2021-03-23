// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/buildartifactsvc"
	"github.com/erda-project/erda/modules/pipeline/services/buildcachesvc"
	"github.com/erda-project/erda/modules/pipeline/services/cmsvc"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinecronsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/modules/pipeline/services/reportsvc"
	"github.com/erda-project/erda/modules/pipeline/services/snippetsvc"
	"github.com/erda-project/erda/pkg/httpserver"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	appSvc           *appsvc.AppSvc
	permissionSvc    *permissionsvc.PermissionSvc
	pipelineCronSvc  *pipelinecronsvc.PipelineCronSvc
	pipelineSvc      *pipelinesvc.PipelineSvc
	crondSvc         *crondsvc.CrondSvc
	cmSvc            *cmsvc.CMSvc
	buildArtifactSvc *buildartifactsvc.BuildArtifactSvc
	buildCacheSvc    *buildcachesvc.BuildCacheSvc
	actionAgentSvc   *actionagentsvc.ActionAgentSvc
	extMarketSvc     *extmarketsvc.ExtMarketSvc
	snippetSvc       *snippetsvc.SnippetSvc
	reportSvc        *reportsvc.ReportSvc

	dbClient           *dbclient.Client
	queryStringDecoder *schema.Decoder

	reconciler *reconciler.Reconciler
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

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(e *Endpoints) {
		e.dbClient = dbClient
	}
}

func WithAppSvc(svc *appsvc.AppSvc) Option {
	return func(e *Endpoints) {
		e.appSvc = svc
	}
}

func WithCMSvc(svc *cmsvc.CMSvc) Option {
	return func(e *Endpoints) {
		e.cmSvc = svc
	}
}

func WithBuildArtifactSvc(svc *buildartifactsvc.BuildArtifactSvc) Option {
	return func(e *Endpoints) {
		e.buildArtifactSvc = svc
	}
}

func WithBuildCacheSvc(svc *buildcachesvc.BuildCacheSvc) Option {
	return func(e *Endpoints) {
		e.buildCacheSvc = svc
	}
}

func WithPermissionSvc(svc *permissionsvc.PermissionSvc) Option {
	return func(e *Endpoints) {
		e.permissionSvc = svc
	}
}

func WithCrondSvc(svc *crondsvc.CrondSvc) Option {
	return func(e *Endpoints) {
		e.crondSvc = svc
	}
}

func WithActionAgentSvc(svc *actionagentsvc.ActionAgentSvc) Option {
	return func(e *Endpoints) {
		e.actionAgentSvc = svc
	}
}

func WithExtMarketSvc(svc *extmarketsvc.ExtMarketSvc) Option {
	return func(e *Endpoints) {
		e.extMarketSvc = svc
	}
}

func WithPipelineCronSvc(svc *pipelinecronsvc.PipelineCronSvc) Option {
	return func(e *Endpoints) {
		e.pipelineCronSvc = svc
	}
}

func WithPipelineSvc(svc *pipelinesvc.PipelineSvc) Option {
	return func(e *Endpoints) {
		e.pipelineSvc = svc
	}
}

func WithSnippetSvc(svc *snippetsvc.SnippetSvc) Option {
	return func(e *Endpoints) {
		e.snippetSvc = svc
	}
}

func WithReportSvc(svc *reportsvc.ReportSvc) Option {
	return func(e *Endpoints) {
		e.reportSvc = svc
	}
}

func WithQueryStringDecoder(decoder *schema.Decoder) Option {
	return func(e *Endpoints) {
		e.queryStringDecoder = decoder
	}
}

func WithReconciler(r *reconciler.Reconciler) Option {
	return func(e *Endpoints) {
		e.reconciler = r
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		// health check
		{Path: "/ping", Method: http.MethodGet, Handler: e.healthCheck},
		// version
		{Path: "/version", Method: http.MethodGet, Handler: e.version},

		// pipelines
		{Path: "/api/v2/pipelines", Method: http.MethodPost, Handler: e.pipelineCreateV2},
		{Path: "/api/pipelines", Method: http.MethodPost, Handler: e.pipelineCreate}, // TODO qa 和 adaptor 通过 bundle 调用 v1 create，需要调整后再下线
		{Path: "/api/pipelines", Method: http.MethodGet, Handler: e.pipelineList},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodGet, Handler: e.pipelineDetail},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodPut, Handler: e.pipelineOperate},
		{Path: "/api/pipelines/{pipelineID}", Method: http.MethodDelete, Handler: e.pipelineDelete},
		{Path: "/api/pipelines/{pipelineID}/actions/run", Method: http.MethodPost, Handler: e.pipelineRun},
		{Path: "/api/pipelines/{pipelineID}/actions/cancel", Method: http.MethodPost, Handler: e.pipelineCancel},
		{Path: "/api/pipelines/{pipelineID}/actions/rerun", Method: http.MethodPost, Handler: e.pipelineRerun},
		{Path: "/api/pipelines/{pipelineID}/actions/rerun-failed", Method: http.MethodPost, Handler: e.pipelineRerunFailed},

		// tasks
		{Path: "/api/pipelines/{pipelineID}/tasks/{taskID}", Method: http.MethodGet, Handler: e.pipelineTaskDetail},
		{Path: "/api/pipelines/{pipelineID}/tasks/{taskID}/actions/get-bootstrap-info", Method: http.MethodGet, Handler: e.taskBootstrapInfo},

		// cms
		{Path: "/api/pipelines/cms/ns", Method: http.MethodPost, Handler: e.createCmsNs},
		{Path: "/api/pipelines/cms/ns", Method: http.MethodGet, Handler: e.listCmsNs},
		{Path: "/api/pipelines/cms/ns/{ns}", Method: http.MethodPost, Handler: e.updateCmsNsConfigs},
		{Path: "/api/pipelines/cms/ns/{ns}", Method: http.MethodDelete, Handler: e.deleteCmsNsConfigs},
		{Path: "/api/pipelines/cms/ns/{ns}", Method: http.MethodGet, Handler: e.getCmsNsConfigs},

		// pipeline related actions
		{Path: "/api/pipelines/actions/batch-create", Method: http.MethodPost, Handler: e.pipelineBatchCreate},
		{Path: "/api/pipelines/actions/pipeline-yml-graph", Method: http.MethodPost, Handler: e.pipelineYmlGraph},
		{Path: "/api/pipelines/actions/statistics", Method: http.MethodGet, Handler: e.pipelineStatistic},
		{Path: "/api/pipelines/actions/task-view", Method: http.MethodGet, Handler: e.pipelineTaskView},

		// pipeline cron
		{Path: "/api/pipeline-crons", Method: http.MethodGet, Handler: e.pipelineCronPaging},
		{Path: "/api/pipeline-crons/{cronID}/actions/start", Method: http.MethodPut, Handler: e.pipelineCronStart},
		{Path: "/api/pipeline-crons/{cronID}/actions/stop", Method: http.MethodPut, Handler: e.pipelineCronStop},
		{Path: "/api/pipeline-crons", Method: http.MethodPost, Handler: e.pipelineCronCreate},
		{Path: "/api/pipeline-crons/{cronID}", Method: http.MethodDelete, Handler: e.pipelineCronDelete},
		{Path: "/api/pipeline-crons/{cronID}", Method: http.MethodGet, Handler: e.pipelineCronGet},

		// build artifact
		{Path: "/api/build-artifacts/{sha}", Method: http.MethodGet, Handler: e.queryBuildArtifact},
		{Path: "/api/build-artifacts", Method: http.MethodPost, Handler: e.registerBuildArtifact},

		// build cache
		{Path: "/api/build-caches", Method: http.MethodPost, Handler: e.reportBuildCache},

		// platform callback
		{Path: "/api/pipelines/actions/callback", Method: http.MethodPost, Handler: e.pipelineCallback},

		// daemon
		{Path: "/_daemon/reload-action-executor-config", Method: http.MethodGet, Handler: e.reloadActionExecutorConfig},
		{Path: "/_daemon/crond/actions/reload", Method: http.MethodGet, Handler: e.crondReload},
		{Path: "/_daemon/crond/actions/snapshot", Method: http.MethodGet, Handler: e.crondSnapshot},
		{Path: "/_daemon/reconciler/throttler/snapshot", Method: http.MethodGet, WriterHandler: e.throttlerSnapshot},

		{Path: "/api/pipeline-snippets/actions/query-details", Method: http.MethodPost, Handler: e.querySnippetDetails},

		// reports
		{Path: "/api/pipeline-reportsets/{pipelineID}", Method: http.MethodGet, Handler: e.queryPipelineReportSet},
		{Path: "/api/pipeline-reportsets", Method: http.MethodGet, Handler: e.pagingPipelineReportSets},
	}
}
