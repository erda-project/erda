// Package endpoints 定义所有的 route handle.
package endpoints

import (
	"net/http"

	"github.com/erda-project/erda/modules/ops/services/kubernetes"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/jsonstore"

	"github.com/erda-project/erda/modules/ops/dbclient"
	"github.com/erda-project/erda/modules/ops/impl/addons"
	cloud_account "github.com/erda-project/erda/modules/ops/impl/cloud-account"
	"github.com/erda-project/erda/modules/ops/impl/clusters"
	"github.com/erda-project/erda/modules/ops/impl/edge"
	"github.com/erda-project/erda/modules/ops/impl/ess"
	"github.com/erda-project/erda/modules/ops/impl/labels"
	"github.com/erda-project/erda/modules/ops/impl/mns"
	"github.com/erda-project/erda/modules/ops/impl/nodes"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	bdl      *bundle.Bundle
	dbclient *dbclient.DBClient

	nodes        *nodes.Nodes
	labels       *labels.Labels
	clusters     *clusters.Clusters
	Mns          *mns.Mns
	Ess          *ess.Ess
	CloudAccount *cloud_account.CloudAccount
	Addons       *addons.Addons
	JS           jsonstore.JsonStore
	CachedJS     jsonstore.JsonStore
	edge         *edge.Edge
}

type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(db *dbclient.DBClient, js jsonstore.JsonStore, cachedJS jsonstore.JsonStore, options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}
	e.dbclient = db
	e.labels = labels.New(db, e.bdl)
	e.nodes = nodes.New(db, e.bdl)
	e.clusters = clusters.New(db, e.bdl)
	e.Mns = mns.New(db, e.bdl, e.nodes, js)
	e.Ess = ess.New(e.bdl, e.Mns, e.nodes, e.labels)
	e.CloudAccount = cloud_account.New(db, cachedJS)
	e.Addons = addons.New(db, e.bdl)
	e.JS = js
	e.CachedJS = cachedJS
	e.edge = edge.New(
		edge.WithDBClient(db),
		edge.WithBundle(e.bdl),
		edge.WithKubernetes(
			kubernetes.New(kubernetes.WithBundle(e.bdl)),
		),
	)
	return e
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bdl = bdl
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/info", Method: http.MethodGet, Handler: e.Info},
		{Path: "/api/node-labels", Method: http.MethodGet, Handler: i18nPrinter(e.ListLabels)},
		{Path: "/api/node-labels", Method: http.MethodPost, Handler: auth(i18nPrinter(e.UpdateLabels))},
		{Path: "/api/nodes", Method: http.MethodPost, Handler: auth(i18nPrinter(e.AddNodes))},
		{Path: "/api/nodes", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.RmNodes))},
		{Path: "/api/records", Method: http.MethodGet, Handler: auth(i18nPrinter(e.Query))},
		{Path: "/api/recordtypes", Method: http.MethodGet, Handler: auth(i18nPrinter(e.RecordTypeList))},
		{Path: "/api/node-logs", Method: http.MethodGet, Handler: auth(i18nPrinter(e.Logs))},
		{Path: "/api/cluster/actions/upgrade", Method: http.MethodPost, Handler: auth(i18nPrinter(e.UpgradeEdgeCluster))},
		{Path: "/api/cluster/actions/batch-upgrade", Method: http.MethodPost, Handler: auth(i18nPrinter(e.BatchUpgradeEdgeCluster))},
		{Path: "/api/cluster", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.OfflineEdgeCluster))},
		{Path: "/api/cluster", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ClusterInfo))},
		{Path: "/api/org-cluster-info", Method: http.MethodGet, Handler: auth(i18nPrinter(e.OrgClusterInfo))},

		// officer apis
		{Path: "/api/clusters/{clusterName}/registry/readonly", Method: http.MethodGet, Handler: e.RegistryReadonly},
		{Path: "/api/clusters/{clusterName}/registry/layers", Method: http.MethodDelete, Handler: e.RegistryRemoveLayers},
		{Path: "/api/clusters", Method: http.MethodPut, Handler: auth(i18nPrinter(e.ClusterUpdate))},
		{Path: "/api/clusters/{clusterName}/registry/manifests/actions/remove", Method: http.MethodPost, Handler: e.RegistryRemoveManifests},
		{Path: "/api/script/info", Method: http.MethodGet, Handler: e.GetScriptInfo},
		{Path: "/api/script/{Name}", Method: http.MethodGet, WriterHandler: e.ServeScript},

		// cloud node apis
		{Path: "/api/ops/cloud-nodes", Method: http.MethodPost, Handler: auth(i18nPrinter(e.AddCloudNodes))},

		// cloud cluster apis
		{Path: "/api/cloud-clusters", Method: http.MethodPost, Handler: auth(i18nPrinter(e.AddCloudClusters))},
		{Path: "/api/cluster-preview", Method: http.MethodPost, Handler: auth(i18nPrinter(e.ClusterPreview))},
		{Path: "/api/ops/lock-cluster", Method: http.MethodPost, Handler: auth(i18nPrinter(e.LockCluster))},
		{Path: "/api/ops/unlock-cluster", Method: http.MethodPost, Handler: auth(i18nPrinter(e.UnlockCluster))},

		{Path: "/api/ops/cloud-resource-list", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListAliyunResources))},
		{Path: "/api/ops/cloud-resource", Method: http.MethodGet, Handler: auth(i18nPrinter(e.QueryCloudResourceDetail))},
		{Path: "/api/ops/cloud-resource-tag", Method: http.MethodPost, Handler: auth(i18nPrinter(e.TagResources))},
		{Path: "/api/cloud-resource/set-tag", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CloudResourceSetTag))},

		// ecs
		{Path: "/api/cloud-ecs", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListECS))},
		{Path: "/api/cloud-ecs/actions/stop", Method: http.MethodPost, Handler: auth(i18nPrinter(e.StopECS))},
		{Path: "/api/cloud-ecs/actions/start", Method: http.MethodPost, Handler: auth(i18nPrinter(e.StartECS))},
		{Path: "/api/cloud-ecs/actions/restart", Method: http.MethodPost, Handler: auth(i18nPrinter(e.RestartECS))},
		{Path: "/api/cloud-ecs/actions/config-renew-attribute", Method: http.MethodPost, Handler: auth(i18nPrinter(e.AutoRenewECS))},

		//vpc, vswitch
		{Path: "/api/cloud-vpc", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListVPC))},
		{Path: "/api/cloud-vpc", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateVPC))},
		{Path: "/api/cloud-vpc/actions/tag-cluster", Method: http.MethodPost, Handler: auth(i18nPrinter(e.VPCTagCluster))},
		{Path: "/api/cloud-vsw", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListVSW))},
		{Path: "/api/cloud-vsw", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateVSW))},

		// region, zone list
		{Path: "/api/cloud-region", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListRegion))},
		{Path: "/api/cloud-zone", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListZone))},

		// cloud-account
		{Path: "/api/cloud-account", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListAccount))},
		{Path: "/api/cloud-account", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateAccount))},
		{Path: "/api/cloud-account", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteAccount))},

		// cloud-mysql
		{Path: "/api/cloud-mysql", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListMysql))},
		{Path: "/api/cloud-mysql/{instanceID}", Method: http.MethodGet, Handler: auth(i18nPrinter(e.GetMysqlDetailInfo))},
		{Path: "/api/cloud-mysql", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateMysqlInstance))},
		{Path: "/api/cloud-mysql", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteMysql))},
		{Path: "/api/cloud-mysql/{instanceID}/databases", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListMysqlDatabase))},
		{Path: "/api/cloud-mysql/actions/delete-db", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteMysqlDatabase))},
		{Path: "/api/cloud-mysql/actions/create-db", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateMysqlDatabase))},
		{Path: "/api/cloud-mysql/{instanceID}/accounts", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListMysqlAccount))},
		{Path: "/api/cloud-mysql/actions/create-account", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateMysqlAccount))},
		{Path: "/api/cloud-mysql/actions/reset-password", Method: http.MethodPost, Handler: auth(i18nPrinter(e.ResetMysqlAccountPassword))},
		{Path: "/api/cloud-mysql/actions/grant-privilege", Method: http.MethodPost, Handler: auth(i18nPrinter(e.GrantMysqlAccountPrivilege))},

		// cloud-redis
		{Path: "/api/cloud-redis", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateRedis))},
		{Path: "/api/cloud-redis", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteRedisResource))},
		{Path: "/api/cloud-redis", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListRedis))},
		{Path: "/api/cloud-redis/{instanceID}", Method: http.MethodGet, Handler: auth(i18nPrinter(e.CetRedisDetailInfo))},

		// cloud-oss
		{Path: "/api/cloud-oss", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateOSS))},
		{Path: "/api/cloud-oss", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteOSSResource))},
		{Path: "/api/cloud-oss", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListOSS))},

		// cloud-ons
		{Path: "/api/cloud-ons", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateOnsInstance))},
		{Path: "/api/cloud-ons", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteOns))},
		{Path: "/api/cloud-ons/actions/delete-topic", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteOnsTopic))},
		{Path: "/api/cloud-ons/actions/list-topic", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListOnsTopic))},
		{Path: "/api/cloud-ons/actions/create-topic", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateOnsTopic))},
		{Path: "/api/cloud-ons/actions/list-group", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListOnsGroup))},
		{Path: "/api/cloud-ons/actions/create-group", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateOnsGroup))},
		{Path: "/api/cloud-ons", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListOns))},
		{Path: "/api/cloud-ons/{instanceID}", Method: http.MethodGet, Handler: auth(i18nPrinter(e.CetOnsDetailInfo))},

		// cloud-gateway
		{Path: "/api/cloud-gateway", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateGatewayVpcGrant))},
		{Path: "/api/cloud-gateway", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ListGatewayAndVpc))},
		{Path: "/api/cloud-gateway", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteGateway))},
		{Path: "/api/cloud-gateway/actions/create-vpc-grant", Method: http.MethodPost, Handler: auth(i18nPrinter(e.CreateGatewayVpcGrant))},
		{Path: "/api/cloud-gateway/actions/delete-vpc-grant", Method: http.MethodDelete, Handler: auth(i18nPrinter(e.DeleteGateway))},

		// cloud resource overview
		{Path: "/api/cloud-resource-overview", Method: http.MethodGet, Handler: auth(i18nPrinter(e.CloudResourceOverview))},
		{Path: "/api/ecs-trending", Method: http.MethodGet, Handler: auth(i18nPrinter(e.ECSTrending))},

		// addon management
		{Path: "/api/addons/actions/config", Method: http.MethodGet, Handler: auth(i18nPrinter(e.GetAddonConfig))},
		{Path: "/api/addons/actions/config", Method: http.MethodPost, Handler: auth(i18nPrinter(e.UpdateAddonConfig))},
		{Path: "/api/addons/actions/scale", Method: http.MethodPost, Handler: auth(i18nPrinter(e.AddonScale))},
		{Path: "/api/addons/status", Method: http.MethodGet, Handler: auth(i18nPrinter(e.GetAddonStatus))},

		{Path: "/api/aliyun-client", Method: http.MethodPost, WriterHandler: e.DoRemoteAction},

		{Path: "/api/internal-cloud-account", Method: http.MethodGet, Handler: i18nPrinter(e.GetCloudAccount)},

		// edge
		// TODO: auth and i18n
		{Path: "/api/edge/site", Method: http.MethodGet, Handler: e.ListEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodGet, Handler: e.GetEdgeSite},
		{Path: "/api/edge/site", Method: http.MethodPost, Handler: e.CreateEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeSite},
		{Path: "/api/edge/site/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeSite},
		{Path: "/api/edge/site/init/{ID}", Method: http.MethodGet, Handler: e.GetInitEdgeSiteShell},
		{Path: "/api/edge/site/offline/{ID}", Method: http.MethodDelete, Handler: e.OfflineEdgeHost},

		{Path: "/api/edge/configset", Method: http.MethodGet, Handler: e.ListEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodGet, Handler: e.GetEdgeConfigSet},
		{Path: "/api/edge/configset", Method: http.MethodPost, Handler: e.CreateEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeConfigSet},
		{Path: "/api/edge/configset/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeConfigSet},

		{Path: "/api/edge/configset-item", Method: http.MethodGet, Handler: e.ListEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodGet, Handler: e.GetEdgeConfigSetItem},
		{Path: "/api/edge/configset-item", Method: http.MethodPost, Handler: e.CreateEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeConfigSetItem},
		{Path: "/api/edge/configset-item/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeConfigSetItem},

		{Path: "/api/edge/app", Method: http.MethodGet, Handler: e.ListEdgeApp},
		{Path: "/api/edge/app", Method: http.MethodPost, Handler: e.CreateEdgeApp},
		{Path: "/api/edge/app/{ID}", Method: http.MethodGet, Handler: e.GetEdgeApp},
		{Path: "/api/edge/app/status/{ID}", Method: http.MethodGet, Handler: e.GetEdgeAppStatus},
		{Path: "/api/edge/app/{ID}", Method: http.MethodPut, Handler: e.UpdateEdgeApp},
		{Path: "/api/edge/app/{ID}", Method: http.MethodDelete, Handler: e.DeleteEdgeApp},

		{Path: "/api/edge/app/site/offline/{ID}", Method: http.MethodPost, Handler: e.OfflineAppSite},
		{Path: "/api/edge/app/site/restart/{ID}", Method: http.MethodPost, Handler: e.RestartAppSite},
	}
}
