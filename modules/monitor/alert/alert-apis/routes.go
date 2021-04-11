package apis

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// Custom alarm
	routes.GET("/api/customize/alerts/metrics", p.queryCustomizeMetric)
	routes.GET("/api/customize/alerts/notifies/targets", p.queryCustomizeNotifyTarget)
	routes.GET("/api/customize/alerts", p.queryCustomizeAlert)
	routes.GET("/api/customize/alerts/:id", p.getCustomizeAlert)
	routes.GET("/api/customize/alerts/:id/detail", p.getCustomizeAlertDetail)
	routes.POST("/api/customize/alerts", p.createCustomizeAlert)
	routes.PUT("/api/customize/alerts/:id", p.updateCustomizeAlert)
	routes.PUT("/api/customize/alerts/:id/switch", p.updateCustomizeAlertEnable)
	routes.DELETE("/api/customize/alerts/:id", p.deleteCustomizeAlert)

	// Enterprise custom alarms
	checkListOrgAlertPermission := permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	)
	routes.GET("/api/orgs/customize/alerts/metrics", p.queryOrgCustomizeMetric, checkListOrgAlertPermission)
	routes.GET("/api/orgs/customize/alerts/notifies/targets", p.queryCustomizeNotifyTarget, checkListOrgAlertPermission)
	routes.GET("/api/orgs/customize/alerts", p.queryOrgCustomizeAlerts, checkListOrgAlertPermission)
	routes.GET("/api/orgs/customize/alerts/:id", p.getOrgCustomizeAlertDetail, checkListOrgAlertPermission)
	routes.POST("/api/orgs/customize/alerts", p.createOrgCustomizeAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionCreate,
	))
	routes.PUT("/api/orgs/customize/alerts/:id", p.updateOrgCustomizeAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionUpdate,
	))
	routes.PUT("/api/orgs/customize/alerts/:id/switch", p.updateOrgCustomizeAlertEnable, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionUpdate,
	))
	routes.DELETE("/api/orgs/customize/alerts/:id", p.deleteOrgCustomizeAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionDelete,
	))

	// Alarm rule generation preview dashboard
	routes.POST("/api/customize/alerts/dash-preview/query", p.queryDashboardByAlert)
	routes.POST("/api/orgs/customize/alerts/dash-preview/query", p.queryOrgDashboardByAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionCreate,
	))

	// alert
	routes.GET("/api/alerts/rules", p.queryAlertRule)
	routes.GET("/api/alerts", p.queryAlert)
	routes.GET("/api/alerts/:id", p.getAlert)
	routes.GET("/api/alerts/:id/detail", p.getAlertDetail)
	routes.POST("/api/alerts", p.createAlert)
	routes.PUT("/api/alerts/:id", p.updateAlert)
	routes.PUT("/api/alerts/:id/switch", p.updateAlertEnable)
	routes.DELETE("/api/alerts/:id", p.deleteAlert)

	// Enterprise Alert
	routes.GET("/api/orgs/alerts/rules", p.queryOrgAlertRule, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.GET("/api/orgs/alerts", p.queryOrgAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.GET("/api/orgs/alerts/:id", p.getOrgAlertDetail, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionGet,
	))
	routes.POST("/api/orgs/alerts", p.createOrgAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionCreate,
	))
	routes.PUT("/api/orgs/alerts/:id", p.updateOrgAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionUpdate,
	))
	routes.PUT("/api/orgs/alerts/:id/switch", p.updateOrgAlertEnable, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionUpdate,
	))
	routes.DELETE("/api/orgs/alerts/:id", p.deleteOrgAlert, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionDelete,
	))

	// Alarm record
	routes.GET("/api/alert-record-attrs", p.getAlertRecordAttr)
	routes.GET("/api/alert-records", p.queryAlertRecord)
	routes.GET("/api/alert-records/:groupId", p.getAlertRecord)
	routes.GET("/api/alert-records/:groupId/histories", p.queryAlertHistory)
	routes.POST("/api/alert-records/:groupId/issues", p.createAlertIssue)
	routes.PUT("/api/alert-records/:groupId/issues/:issueId", p.updateAlertIssue)

	// Enterprise Alarm Record
	routes.GET("/api/org-alert-record-attrs", p.getOrgAlertRecordAttr, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.GET("/api/org-alert-records", p.queryOrgAlertRecord, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.POST("/api/org-hosts-alert-records", p.queryOrgHostsAlertRecord, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.GET("/api/org-alert-records/:groupId", p.getOrgAlertRecord, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionGet,
	))
	routes.GET("/api/org-alert-records/:groupId/histories", p.queryOrgAlertHistory, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionList,
	))
	routes.POST("/api/org-alert-records/:groupId/issues", p.createOrgAlertIssue, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionCreate,
	))
	routes.PUT("/api/org-alert-records/:groupId/issues", p.updateOrgAlertIssue, permission.Intercepter(
		permission.ScopeOrg, permission.OrgIDFromHeader(),
		common.ResourceOrgAlert, permission.ActionUpdate,
	))

	return nil
}
