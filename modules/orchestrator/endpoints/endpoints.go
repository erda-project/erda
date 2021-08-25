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

// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/events"
	"github.com/erda-project/erda/modules/orchestrator/queue"
	"github.com/erda-project/erda/modules/orchestrator/services/addon"
	"github.com/erda-project/erda/modules/orchestrator/services/deployment"
	"github.com/erda-project/erda/modules/orchestrator/services/domain"
	"github.com/erda-project/erda/modules/orchestrator/services/instance"
	"github.com/erda-project/erda/modules/orchestrator/services/migration"
	"github.com/erda-project/erda/modules/orchestrator/services/resource"
	"github.com/erda-project/erda/modules/orchestrator/services/runtime"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/goroutinepool"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	db         *dbclient.DBClient // TODO: Endpoints should not take db
	queue      *queue.PusherQueue
	bdl        *bundle.Bundle
	pool       *goroutinepool.GoroutinePool
	evMgr      *events.EventManager
	runtime    *runtime.Runtime
	deployment *deployment.Deployment
	domain     *domain.Domain
	addon      *addon.Addon
	resource   *resource.Resource
	encrypt    *encryption.EnvEncrypt
	instance   *instance.Instance
	migration  *migration.Migration
}

// Option Endpoints 配置选项
type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

// WithDBClient 配置 db
func WithDBClient(db *dbclient.DBClient) Option {
	return func(e *Endpoints) {
		e.db = db
	}
}

// WithPool 配置 pool
func WithPool(pool *goroutinepool.GoroutinePool) Option {
	return func(e *Endpoints) {
		e.pool = pool
	}
}

// WithQueue 配置 queue
func WithQueue(queue *queue.PusherQueue) Option {
	return func(e *Endpoints) {
		e.queue = queue
	}
}

// WithEventManager 配置 EventManager
func WithEventManager(evMgr *events.EventManager) Option {
	return func(e *Endpoints) {
		e.evMgr = evMgr
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// WithRuntime 设置 runtime 对象.
func WithRuntime(runtime *runtime.Runtime) Option {
	return func(e *Endpoints) {
		e.runtime = runtime
	}
}

// WithDeployment 设置 deployment 对象.
func WithDeployment(deployment *deployment.Deployment) Option {
	return func(e *Endpoints) {
		e.deployment = deployment
	}
}

// WithDomain 设置 domain 对象.
func WithDomain(domain *domain.Domain) Option {
	return func(e *Endpoints) {
		e.domain = domain
	}
}

// WithAddon 设置 addon service
func WithAddon(addon *addon.Addon) Option {
	return func(e *Endpoints) {
		e.addon = addon
	}
}

// WithInstance 设置 instance 对象
func WithInstance(instance *instance.Instance) Option {
	return func(e *Endpoints) {
		e.instance = instance
	}
}

// WithEnvEncrypt 设置 encrypt service
func WithEnvEncrypt(encrypt *encryption.EnvEncrypt) Option {
	return func(e *Endpoints) {
		e.encrypt = encrypt
	}
}

// WithResource 设置 resource service
func WithResource(resource *resource.Resource) Option {
	return func(e *Endpoints) {
		e.resource = resource
	}
}

// WithMigration 设置 migration service
func WithMigration(migration *migration.Migration) Option {
	return func(e *Endpoints) {
		e.migration = migration
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		// system endpoints
		{Path: "/info", Method: http.MethodGet, Handler: e.Info},

		// runtime endpoints
		{Path: "/api/runtimes", Method: http.MethodPost, Handler: e.CreateRuntime},
		{Path: "/api/runtimes/actions/deploy-release", Method: http.MethodPost, Handler: e.CreateRuntimeByRelease},
		{Path: "/api/runtimes/actions/deploy-release-action", Method: http.MethodPost, Handler: e.CreateRuntimeByReleaseAction},
		{Path: "/api/runtimes", Method: http.MethodGet, Handler: e.ListRuntimes},
		{Path: "/api/runtimes/{idOrName}", Method: http.MethodGet, Handler: e.GetRuntime},
		{Path: "/api/runtimes/{runtimeID}", Method: http.MethodDelete, Handler: e.DeleteRuntime},
		// TODO: change configuration -> spec
		{Path: "/api/runtimes/{runtimeID}/configuration", Method: http.MethodGet, Handler: e.GetRuntimeSpec},
		{Path: "/api/runtimes/{runtimeID}/actions/stop", Method: http.MethodPost, Handler: e.StopRuntime},
		{Path: "/api/runtimes/{runtimeID}/actions/start", Method: http.MethodPost, Handler: e.StartRuntime},
		{Path: "/api/runtimes/{runtimeID}/actions/restart", Method: http.MethodPost, Handler: e.RestartRuntime},
		{Path: "/api/runtimes/{runtimeID}/actions/redeploy", Method: http.MethodPost, Handler: e.RedeployRuntime},
		{Path: "/api/runtimes/{runtimeID}/actions/redeploy-action", Method: http.MethodPost, Handler: e.RedeployRuntimeAction},
		{Path: "/api/runtimes/{runtimeID}/actions/rollback", Method: http.MethodPost, Handler: e.RollbackRuntime},
		{Path: "/api/runtimes/{runtimeID}/actions/rollback-action", Method: http.MethodPost, Handler: e.RollbackRuntimeAction},
		{Path: "/api/runtimes/actions/bulk-get-status", Method: http.MethodGet, Handler: e.epBulkGetRuntimeStatusDetail},
		{Path: "/api/runtimes/actions/update-pre-overlay", Method: http.MethodPut, Handler: e.epUpdateOverlay},
		{Path: "/api/runtimes/actions/full-gc", Method: http.MethodPost, Handler: e.FullGC},
		{Path: "/api/runtimes/actions/refer-cluster", Method: http.MethodGet, Handler: e.ReferCluster},
		{Path: "/api/runtimes/deploy/logs", Method: http.MethodGet, Handler: e.RuntimeLogs},
		{Path: "/api/runtimes/actions/get-app-workspace-releases", Method: http.MethodGet, Handler: e.GetAppWorkspaceReleases},
		// kill pod (only k8s)
		{Path: "/api/runtimes/actions/killpod", Method: http.MethodPost, Handler: e.KillPod},

		// deployment endpoints
		{Path: "/api/deployments", Method: http.MethodGet, Handler: e.ListDeployments},
		{Path: "/api/deployments/actions/list-launched-approval", Method: http.MethodGet, Handler: e.ListLaunchedApprovalDeployments},
		{Path: "/api/deployments/actions/list-pending-approval", Method: http.MethodGet, Handler: e.ListPendingApprovalDeployments},
		{Path: "/api/deployments/actions/list-approved", Method: http.MethodGet, Handler: e.ListApprovedDeployments},
		{Path: "/api/deployments/actions/approve", Method: http.MethodPost, Handler: e.DeploymentApprove},

		// TODO: do not returns runtime info, use /api/runtimes/{runtimeId} instead
		{Path: "/api/deployments/{deploymentID}/status", Method: http.MethodGet, Handler: e.GetDeploymentStatus},
		{Path: "/api/deployments/{deploymentID}/actions/cancel", Method: http.MethodPost, Handler: e.CancelDeployment},

		{Path: "/api/deployments/{deploymentID}/actions/deploy-addons", Method: http.MethodPost, Handler: e.DeployStagesAddons},
		{Path: "/api/deployments/{deploymentID}/actions/deploy-services", Method: http.MethodPost, Handler: e.DeployStagesServices},
		{Path: "/api/deployments/{deploymentID}/actions/deploy-domains", Method: http.MethodPost, Handler: e.DeployStagesDomains},

		// domain endpoints
		// TODO: api should be `/api/domains`
		{Path: "/api/runtimes/{runtimeID}/domains", Method: http.MethodGet, Handler: e.ListDomains},
		{Path: "/api/runtimes/{runtimeID}/domains", Method: http.MethodPut, Handler: e.UpdateDomains},

		// instance endpoints
		{Path: "/api/instances/actions/get-service", Method: http.MethodGet, Handler: e.ListServiceInstance},
		{Path: "/api/instances/actions/get-service-pods", Method: http.MethodGet, Handler: e.ListServicePod},
		// 实例统计
		{Path: "/api/clusters/{cluster}/instances-usage", Method: http.MethodGet, Handler: e.InstancesUsage},
		{Path: "/api/instances-usage", Method: http.MethodGet, Handler: e.InstancesUsage},

		// addon endpoints
		{Path: "/api/addons/actions/create-addon", Method: http.MethodPost, Handler: e.CreateAddonDirectly},
		{Path: "/api/addons/actions/create-tenant", Method: http.MethodPost, Handler: e.CreateAddonTenant},
		{Path: "/api/addons/actions/create-custom", Method: http.MethodPost, Handler: e.CreateCustomAddon},
		{Path: "/api/addons/{addonID}/actions/update-custom", Method: http.MethodPut, Handler: e.UpdateCustomAddon},
		{Path: "/api/addons/{addonID}", Method: http.MethodDelete, Handler: e.DeleteAddon},
		{Path: "/api/addons/{addonID}", Method: http.MethodGet, Handler: e.GetAddon},
		{Path: "/api/addons/{addonID}/actions/references", Method: http.MethodGet, Handler: e.GetAddonReferences},
		{Path: "/api/addons", Method: http.MethodGet, Handler: e.ListAddon},
		{Path: "/api/addons/actions/list-extension", Method: http.MethodGet, Handler: e.ListExtensionAddon},
		{Path: "/api/addons/types/{addonName}", Method: http.MethodGet, Handler: e.ListByAddonName},
		{Path: "/api/addons/actions/list-available", Method: http.MethodGet, Handler: e.ListAvailableAddon},
		{Path: "/api/addons/actions/menu", Method: http.MethodGet, Handler: e.ListAddonMenu},
		{Path: "/api/addon-platform/addons/{addonID}/action/provision", Method: http.MethodPost, Handler: e.AddonCreateCallback},
		{Path: "/api/addon-platform/addons/{addonID}/action/deprovision", Method: http.MethodPost, Handler: e.AddonDeleteCallback},
		{Path: "/api/addon-platform/addons/{addonID}/config", Method: http.MethodPost, Handler: e.AddonConfigCallback},
		{Path: "/api/addons/actions/list-customs", Method: http.MethodGet, Handler: e.ListCustomAddon},

		// middleware endpoints(real addon instance)
		{Path: "/api/middlewares", Method: http.MethodGet, Handler: e.ListMiddleware},
		{Path: "/api/middlewares/{middlewareID}", Method: http.MethodGet, Handler: e.GetMiddleware},
		{Path: "/api/middlewares/inner/{middlewareID}", Method: http.MethodGet, Handler: e.InnerGetMiddleware},
		{Path: "/api/middlewares/resource/classification", Method: http.MethodGet, Handler: e.GetMiddlewareAddonClassification},
		{Path: "/api/middlewares/resource/daily", Method: http.MethodGet, Handler: e.GetMiddlewareAddonDaily},
		{Path: "/api/middlewares/{middlewareID}/actions/get-resource", Method: http.MethodGet, Handler: e.GetMiddlewareResource},

		// microService
		{Path: "/api/microservice/projects", Method: http.MethodGet, Handler: e.ListMicroServiceProject},
		{Path: "/api/microservice/project/{projectID}/menus", Method: http.MethodGet, Handler: e.ListMicroServiceMenu},

		// addon metrics
		{Path: "/api/metrics/charts/{scope}/histogram", Method: http.MethodGet, Handler: e.AddonMetrics},
		{Path: "/api/addons/{instanceId}/logs", Method: http.MethodGet, Handler: e.AddonLogs},

		// migration log
		{Path: "/api/migration/{migrationId}/logs", Method: http.MethodGet, Handler: e.MigrationLog},

		// orgcenter jobs log
		{Path: "/api/orgCenter/job/logs", Method: http.MethodGet, Handler: e.OrgcenterJobLogs},

		// project resource info
		{Path: "/api/projects/resource", Method: http.MethodPost, Handler: e.GetProjectResource},

		// resource info
		{Path: "/api/resources/reference", Method: http.MethodGet, Handler: e.GetClusterResourceReference},

		// export, import addonyml
		{Path: "/api/addon/action/yml-export", Method: http.MethodPost, Handler: e.AddonYmlExport},
		{Path: "/api/addon/action/yml-import", Method: http.MethodPost, Handler: e.AddonYmlImport},
	}
}

// TODO: we should refactor the polling
func (e *Endpoints) PushOnDeploymentPolling() (abort bool, err0 error) {
	deployments, err := e.db.FindUnfinishedDeployments()
	if err != nil {
		logrus.Warnf("failed to find unfinished deployments to continue, (%v)", err)
	}
	for _, d := range deployments {
		if _, err := e.db.GetRuntime(d.RuntimeId); err != nil {
			continue
		}
		e.queue.Push(queue.DEPLOY_CONTINUING, strconv.Itoa(int(d.ID)))
	}
	return
}

func (e *Endpoints) PushOnDeployment() (bool, error) {
	item, err := e.queue.Pop(queue.DEPLOY_CONTINUING)
	if err != nil {
		logrus.Warn("failed to pop DEPLOY_CONTINUING task")
		return false, err
	}
	if item == "" {
		// tasks not found
		return false, errors.New("tasks not found")
	}
	deploymentID, err := strconv.Atoi(item)
	if err != nil {
		logrus.Warnf("failed to push on DEPLOY_CONTINUING task, item: %s not a number as deploymentID", item)
		return false, err
	}

	doContinueDeploy := func() {
		err := e.deployment.ContinueDeploy(uint64(deploymentID))
		if err != nil {
			logrus.Warnf("failed to continue deploy, deploymentID: %d, (%v)", deploymentID, err)
		}
		// unlock
		if _, err := e.queue.Unlock(queue.DEPLOY_CONTINUING, item); err != nil {
			logrus.Errorf("[alert] failed to unlock %v/%v, (%v)", queue.DEPLOY_CONTINUING, item, err)
		}
	}
	if err := e.pool.GoWithTimeout(doContinueDeploy, 5*time.Second); err != nil {
		logrus.Warnf("failed to continue deploy, deploymentID: %d, timeout after 5 second", deploymentID)
		return false, err
	}
	return false, nil
}

// TODO: we should refactor the polling
func (e *Endpoints) PushOnDeletingRuntimesPolling() (abort bool, err0 error) {
	runtimes, err := e.db.FindDeletingRuntimes()
	if err != nil {
		logrus.Warnf("failed to find deleting runtimes, (%v)", err)
	}
	logrus.Debugf("%d runtimes to deleting", len(runtimes))
	for _, r := range runtimes {
		e.queue.Push(queue.RUNTIME_DELETING, strconv.FormatUint(r.ID, 10))
	}
	return
}

func (e *Endpoints) PushOnDeletingRuntimes() (abort bool, err0 error) {
	item, err := e.queue.Pop(queue.RUNTIME_DELETING)
	if err != nil {
		logrus.Warn("failed to pop RUNTIME_DELETING task")
		return
	}
	if item == "" {
		// no tasks found
		return
	}
	runtimeID, err := strconv.ParseUint(item, 10, 64)
	if err != nil {
		logrus.Warnf("failed to execute RUNTIME_DELETING task, item: %s not a number as runtimeID", item)
		return
	}
	if err := e.runtime.Destroy(runtimeID); err != nil {
		logrus.Warnf("failed to execute RUNTIME_DELETING task, (%v)", err)
		return
	}
	if _, err := e.queue.Unlock(queue.RUNTIME_DELETING, item); err != nil {
		logrus.Warnf("failed to unlock %v/%v, (%v)", queue.RUNTIME_DELETING, item, err)
	}
	return
}

func (e *Endpoints) FullGCLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			logrus.Infof("start full GC at: %v", t)
			e.runtime.FullGC()
			logrus.Infof("end full GC at: %v, started at: %v", time.Now(), t)
		case <-ctx.Done():
			return
		}
	}
}
