//generated file, DO NOT EDIT
package auto_register

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol"
	actionactionForm "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/action/components/actionForm"
	apppipelinetreefileTree "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/app-pipeline-tree/components/fileTree"
	apppipelinetreenodeFormModal "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/app-pipeline-tree/components/nodeFormModal"
	autotestplandetailaddScenesSetButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/addScenesSetButton"
	autotestplandetailcancelExecuteButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/cancelExecuteButton"
	autotestplandetailenvBaseInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envBaseInfo"
	autotestplandetailenvBaseInfoTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envBaseInfoTitle"
	autotestplandetailenvDrawer "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envDrawer"
	autotestplandetailenvGlobalInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envGlobalInfo"
	autotestplandetailenvGlobalTable "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envGlobalTable"
	autotestplandetailenvGlobalText "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envGlobalText"
	autotestplandetailenvGlobalTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envGlobalTitle"
	autotestplandetailenvHeaderInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envHeaderInfo"
	autotestplandetailenvHeaderTable "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envHeaderTable"
	autotestplandetailenvHeaderText "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envHeaderText"
	autotestplandetailenvHeaderTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/envHeaderTitle"
	autotestplandetailexecuteAlertInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeAlertInfo"
	autotestplandetailexecuteHead "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHead"
	autotestplandetailexecuteHistory "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHistory"
	autotestplandetailexecuteHistoryButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHistoryButton"
	autotestplandetailexecuteHistoryPop "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHistoryPop"
	autotestplandetailexecuteHistoryRefresh "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHistoryRefresh"
	autotestplandetailexecuteHistoryTable "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeHistoryTable"
	autotestplandetailexecuteInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeInfo"
	autotestplandetailexecuteInfoTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeInfoTitle"
	autotestplandetailexecuteTaskBreadcrumb "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeTaskBreadcrumb"
	autotestplandetailexecuteTaskTable "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeTaskTable"
	autotestplandetailexecuteTaskTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/executeTaskTitle"
	autotestplandetailfileConfig "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileConfig"
	autotestplandetailfileDetail "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileDetail"
	autotestplandetailfileExecute "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileExecute"
	autotestplandetailfileInfo "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileInfo"
	autotestplandetailfileInfoHead "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileInfoHead"
	autotestplandetailfileInfoTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/fileInfoTitle"
	autotestplandetailrefreshButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/refreshButton"
	autotestplandetailscenesSetDrawer "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/scenesSetDrawer"
	autotestplandetailscenesSetInParams "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/scenesSetInParams"
	autotestplandetailscenesSetSelect "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/scenesSetSelect"
	autotestplandetailstages "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/stages"
	autotestplandetailstagesOperations "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/stagesOperations"
	autotestplandetailstagesTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/stagesTitle"
	autotestplandetailtabExecuteButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/components/tabExecuteButton"
	edgeappsiteipappSiteBreadcrumb "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site-ip/components/appSiteBreadcrumb"
	edgeappsiteipsiteIpList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site-ip/components/siteIpList"
	edgeappsiteipstatusViewGroup "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site-ip/components/statusViewGroup"
	edgeappsiteappSiteBreadcrumb "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site/components/appSiteBreadcrumb"
	edgeappsiteappSiteManage "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site/components/appSiteManage"
	edgeappsitesiteNameFilter "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site/components/siteNameFilter"
	edgeappsitestatusViewGroup "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-app-site/components/statusViewGroup"
	edgeapplicationaddAppButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/addAppButton"
	edgeapplicationaddAppDrawer "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/addAppDrawer"
	edgeapplicationappConfigForm "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/appConfigForm"
	edgeapplicationapplicationList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/applicationList"
	edgeapplicationkeyValueList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/keyValueList"
	edgeapplicationkeyValueListTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-application/components/keyValueListTitle"
	edgeconfigSetitemclusterAddButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet-item/components/clusterAddButton"
	edgeconfigSetitemconfigItemFormModal "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet-item/components/configItemFormModal"
	edgeconfigSetitemconfigItemList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet-item/components/configItemList"
	edgeconfigSetitemconfigItemListFilter "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet-item/components/configItemListFilter"
	edgeconfigSetclusterAddButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet/components/clusterAddButton"
	edgeconfigSetconfigSetFormModal "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet/components/configSetFormModal"
	edgeconfigSetconfigSetList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-configSet/components/configSetList"
	edgesitesiteAddButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/siteAddButton"
	edgesitesiteAddDrawer "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/siteAddDrawer"
	edgesitesiteFormModal "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/siteFormModal"
	edgesitesiteList "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/siteList"
	edgesitesiteNameFilter "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/siteNameFilter"
	edgesitesitePreview "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/edge-site/components/sitePreview"
	notifyconfignotifyAddButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/notify-config/components/notifyAddButton"
	notifyconfignotifyConfigModal "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/notify-config/components/notifyConfigModal"
	notifyconfignotifyConfigTable "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/notify-config/components/notifyConfigTable"
	notifyconfignotifyTitle "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/notify-config/components/notifyTitle"
	orglistallfilter "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-all/components/filter"
	orglistalllist "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-all/components/list"
	orglistallpage "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-all/components/page"
	orglistmycreateButton "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/createButton"
	orglistmyemptyContainer "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/emptyContainer"
	orglistmyemptyText "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/emptyText"
	orglistmyfilter "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/filter"
	orglistmylist "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/list"
	orglistmypage "github.com/erda-project/erda/modules/tools/openapi/legacy/component-protocol/scenarios/org-list-my/components/page"
)

func RegisterAll() {
	specs := []*protocol.CompRenderSpec{
		{Scenario: "org-list-all", CompName: "filter", RenderC: orglistallfilter.RenderCreator},
		{Scenario: "org-list-all", CompName: "list", RenderC: orglistalllist.RenderCreator},
		{Scenario: "org-list-all", CompName: "page", RenderC: orglistallpage.RenderCreator},
		{Scenario: "edge-app-site", CompName: "appSiteBreadcrumb", RenderC: edgeappsiteappSiteBreadcrumb.RenderCreator},
		{Scenario: "edge-app-site", CompName: "appSiteManage", RenderC: edgeappsiteappSiteManage.RenderCreator},
		{Scenario: "edge-app-site", CompName: "siteNameFilter", RenderC: edgeappsitesiteNameFilter.RenderCreator},
		{Scenario: "edge-app-site", CompName: "statusViewGroup", RenderC: edgeappsitestatusViewGroup.RenderCreator},
		{Scenario: "edge-configSet", CompName: "clusterAddButton", RenderC: edgeconfigSetclusterAddButton.RenderCreator},
		{Scenario: "edge-configSet", CompName: "configSetFormModal", RenderC: edgeconfigSetconfigSetFormModal.RenderCreator},
		{Scenario: "edge-configSet", CompName: "configSetList", RenderC: edgeconfigSetconfigSetList.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "addScenesSetButton", RenderC: autotestplandetailaddScenesSetButton.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "cancelExecuteButton", RenderC: autotestplandetailcancelExecuteButton.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envBaseInfo", RenderC: autotestplandetailenvBaseInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envBaseInfoTitle", RenderC: autotestplandetailenvBaseInfoTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envDrawer", RenderC: autotestplandetailenvDrawer.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envGlobalInfo", RenderC: autotestplandetailenvGlobalInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envGlobalTable", RenderC: autotestplandetailenvGlobalTable.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envGlobalText", RenderC: autotestplandetailenvGlobalText.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envGlobalTitle", RenderC: autotestplandetailenvGlobalTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envHeaderInfo", RenderC: autotestplandetailenvHeaderInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envHeaderTable", RenderC: autotestplandetailenvHeaderTable.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envHeaderText", RenderC: autotestplandetailenvHeaderText.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "envHeaderTitle", RenderC: autotestplandetailenvHeaderTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeAlertInfo", RenderC: autotestplandetailexecuteAlertInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHead", RenderC: autotestplandetailexecuteHead.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHistory", RenderC: autotestplandetailexecuteHistory.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHistoryButton", RenderC: autotestplandetailexecuteHistoryButton.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHistoryPop", RenderC: autotestplandetailexecuteHistoryPop.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHistoryRefresh", RenderC: autotestplandetailexecuteHistoryRefresh.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeHistoryTable", RenderC: autotestplandetailexecuteHistoryTable.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeInfo", RenderC: autotestplandetailexecuteInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeInfoTitle", RenderC: autotestplandetailexecuteInfoTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeTaskBreadcrumb", RenderC: autotestplandetailexecuteTaskBreadcrumb.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeTaskTable", RenderC: autotestplandetailexecuteTaskTable.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "executeTaskTitle", RenderC: autotestplandetailexecuteTaskTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileConfig", RenderC: autotestplandetailfileConfig.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileDetail", RenderC: autotestplandetailfileDetail.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileExecute", RenderC: autotestplandetailfileExecute.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileInfo", RenderC: autotestplandetailfileInfo.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileInfoHead", RenderC: autotestplandetailfileInfoHead.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "fileInfoTitle", RenderC: autotestplandetailfileInfoTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "refreshButton", RenderC: autotestplandetailrefreshButton.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "scenesSetDrawer", RenderC: autotestplandetailscenesSetDrawer.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "scenesSetInParams", RenderC: autotestplandetailscenesSetInParams.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "scenesSetSelect", RenderC: autotestplandetailscenesSetSelect.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "stages", RenderC: autotestplandetailstages.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "stagesOperations", RenderC: autotestplandetailstagesOperations.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "stagesTitle", RenderC: autotestplandetailstagesTitle.RenderCreator},
		{Scenario: "auto-test-plan-detail", CompName: "tabExecuteButton", RenderC: autotestplandetailtabExecuteButton.RenderCreator},
		{Scenario: "edge-app-site-ip", CompName: "appSiteBreadcrumb", RenderC: edgeappsiteipappSiteBreadcrumb.RenderCreator},
		{Scenario: "edge-app-site-ip", CompName: "siteIpList", RenderC: edgeappsiteipsiteIpList.RenderCreator},
		{Scenario: "edge-app-site-ip", CompName: "statusViewGroup", RenderC: edgeappsiteipstatusViewGroup.RenderCreator},
		{Scenario: "edge-application", CompName: "addAppButton", RenderC: edgeapplicationaddAppButton.RenderCreator},
		{Scenario: "edge-application", CompName: "addAppDrawer", RenderC: edgeapplicationaddAppDrawer.RenderCreator},
		{Scenario: "edge-application", CompName: "appConfigForm", RenderC: edgeapplicationappConfigForm.RenderCreator},
		{Scenario: "edge-application", CompName: "applicationList", RenderC: edgeapplicationapplicationList.RenderCreator},
		{Scenario: "edge-application", CompName: "keyValueList", RenderC: edgeapplicationkeyValueList.RenderCreator},
		{Scenario: "edge-application", CompName: "keyValueListTitle", RenderC: edgeapplicationkeyValueListTitle.RenderCreator},
		{Scenario: "edge-configSet-item", CompName: "clusterAddButton", RenderC: edgeconfigSetitemclusterAddButton.RenderCreator},
		{Scenario: "edge-configSet-item", CompName: "configItemFormModal", RenderC: edgeconfigSetitemconfigItemFormModal.RenderCreator},
		{Scenario: "edge-configSet-item", CompName: "configItemList", RenderC: edgeconfigSetitemconfigItemList.RenderCreator},
		{Scenario: "edge-configSet-item", CompName: "configItemListFilter", RenderC: edgeconfigSetitemconfigItemListFilter.RenderCreator},
		{Scenario: "edge-site", CompName: "siteAddButton", RenderC: edgesitesiteAddButton.RenderCreator},
		{Scenario: "edge-site", CompName: "siteAddDrawer", RenderC: edgesitesiteAddDrawer.RenderCreator},
		{Scenario: "edge-site", CompName: "siteFormModal", RenderC: edgesitesiteFormModal.RenderCreator},
		{Scenario: "edge-site", CompName: "siteList", RenderC: edgesitesiteList.RenderCreator},
		{Scenario: "edge-site", CompName: "siteNameFilter", RenderC: edgesitesiteNameFilter.RenderCreator},
		{Scenario: "edge-site", CompName: "sitePreview", RenderC: edgesitesitePreview.RenderCreator},
		{Scenario: "notify-config", CompName: "notifyAddButton", RenderC: notifyconfignotifyAddButton.RenderCreator},
		{Scenario: "notify-config", CompName: "notifyConfigModal", RenderC: notifyconfignotifyConfigModal.RenderCreator},
		{Scenario: "notify-config", CompName: "notifyConfigTable", RenderC: notifyconfignotifyConfigTable.RenderCreator},
		{Scenario: "notify-config", CompName: "notifyTitle", RenderC: notifyconfignotifyTitle.RenderCreator},
		{Scenario: "action", CompName: "actionForm", RenderC: actionactionForm.RenderCreator},
		{Scenario: "app-pipeline-tree", CompName: "fileTree", RenderC: apppipelinetreefileTree.RenderCreator},
		{Scenario: "app-pipeline-tree", CompName: "nodeFormModal", RenderC: apppipelinetreenodeFormModal.RenderCreator},
		{Scenario: "org-list-my", CompName: "createButton", RenderC: orglistmycreateButton.RenderCreator},
		{Scenario: "org-list-my", CompName: "emptyContainer", RenderC: orglistmyemptyContainer.RenderCreator},
		{Scenario: "org-list-my", CompName: "emptyText", RenderC: orglistmyemptyText.RenderCreator},
		{Scenario: "org-list-my", CompName: "filter", RenderC: orglistmyfilter.RenderCreator},
		{Scenario: "org-list-my", CompName: "list", RenderC: orglistmylist.RenderCreator},
		{Scenario: "org-list-my", CompName: "page", RenderC: orglistmypage.RenderCreator},
	}

	for _, s := range specs {
		if err := protocol.Register(s); err != nil {
			logrus.Errorf("register render failed, scenario: %v, components: %v, err: %v", s.Scenario, s.CompName, err)
			panic(err)
		}
	}

	var protocols = map[string]string{
		"edge-app-site": `
scenario: edge-app-site

hierarchy:
  root: siteManage
  structure:
    siteManage:
      - head
      - appSiteManage
    head:
      left: appSiteBreadcrumb
      right: siteFilterGroup
    siteFilterGroup:
      - siteNameFilter
      - statusViewGroup

components:
  siteManage:
    type: Container
  head:
    type: LRContainer
  appSiteManage:
    type: Table
  statusViewGroup:
    type: Radio
  appSiteBreadcrumb:
    type: Breadcrumb
  siteFilterGroup:
    type: RowContainer
  siteNameFilter:
    type: ContractiveFilter

rendering:
  appSiteManage:
    - name: statusViewGroup
  statusViewGroup:
    - name: appSiteManage
      state:
        - name: viewGroupSelected
          value: "{{ statusViewGroup.viewGroupSelected }}"
  siteNameFilter:
    - name: statusViewGroup
      state:
        - name: searchCondition
          value: "{{ siteNameFilter.searchCondition }}"
    - name: appSiteManage
      state:
        - name: searchCondition
          value: "{{ siteNameFilter.searchCondition }}"`,
		"edge-configSet": `
scenario: edge-configSet

hierarchy:
  root: configSetManage
  structure:
    configSetManage:
      - topHead
      - configSetList
      - configSetFormModal
    topHead:
      - clusterAddButton

components:
  configSetManage:
    type: Container
  configSetList:
    type: Table
  topHead:
    type: RowContainer
    props:
      isTopHead: true
  configSetFormModal:
    type: FormModal
  clusterAddButton:
    type: Button

rendering:
  configSetFormModal:
    - name: configSetList`,
		"org-list-all": `
# 场景名
scenario: "org-list-all"

# 布局
hierarchy:
  root: page
  structure:
    page: 
      children:
        - myPage
    myPage:
      - filter
      - list

components:
  page:
    type: Tabs
  myPage:
    type: Container
  filter:
    type: ContractiveFilter
  list:
    type: List
  
rendering:
  filter:
    - name: list
      state:
        - name: "searchEntry"
          value: "{{ filter.searchEntry }}"
        - name: "searchRefresh"
          value: "{{ filter.searchRefresh }}"
`,
		"edge-app-site-ip": `
scenario: edge-app-site-ip

hierarchy:
  root: siteIpManage
  structure:
    siteIpManage:
      - head
      - siteIpList
    head:
      left: appSiteBreadcrumb
      right: statusViewGroup

components:
  siteIpManage:
    type: Container
  head:
    type: LRContainer
  siteIpList:
    type: Table
  statusViewGroup:
    type: Radio
  appSiteBreadcrumb:
    type: Breadcrumb

rendering:
  statusViewGroup:
    - name: siteIpList
      state:
        - name: viewGroupSelected
          value: "{{ statusViewGroup.viewGroupSelected }}"`,
		"edge-application": `
scenario: edge-application

hierarchy:
  root: appManage
  structure:
    appManage:
      - topHead
      - applicationList
      - addAppDrawer
    topHead:
      - addAppButton
    addAppDrawer:
      content:
        - appConfigForm
        - keyValueListTitle
        - keyValueList

components:
  appManage:
    type: Container
  topHead:
    type: RowContainer
    props:
      isTopHead: true
  applicationList:
    type: Table
  keyValueListTitle:
    type: Title
  keyValueList:
    type: Table
  addAppButton:
    type: Button
  addAppDrawer:
    type: Drawer
  appConfigForm:
    type: Form

rendering:
  appConfigForm:
    - name: applicationList
    - name: addAppDrawer
      state:
        - name: visible
          value: "{{ appConfigForm.addAppDrawerVisible }}"
  addAppButton:
    - name: keyValueListTitle
      state:
        - name: visible
          value: "{{ addAppButton.keyValueListTitleVisible }}"
    - name: keyValueList
      state:
        - name: visible
          value: "{{ addAppButton.keyValueListVisible }}"
    - name: addAppDrawer
      state:
        - name: visible
          value: "{{ addAppButton.addAppDrawerVisible }}"
        - name: operationType
          value: "{{ addAppButton.operationType }}"
    - name: appConfigForm
      state:
        - name: visible
          value: "{{ addAppButton.appConfigFormVisible }}"
        - name: formClear
          value: "{{ addAppButton.formClear }}"
        - name: operationType
          value: "{{ addAppButton.operationType }}"
  applicationList:
    - name: addAppDrawer
      state:
        - name: visible
          value: "{{ applicationList.addAppDrawerVisible }}"
        - name: operationType
          value: "{{ applicationList.operationType }}"
    - name: appConfigForm
      state:
        - name: operationType
          value: "{{ applicationList.operationType }}"
        - name: visible
          value: "{{ applicationList.appConfigFormVisible }}"
        - name: appID
          value: "{{ applicationList.appID }}"
        - name: formClear
          value: "{{ applicationList.formClear }}"
    - name: keyValueList
      state:
        - name: visible
          value: "{{ applicationList.keyValueListVisible }}"
        - name: appID
          value: "{{ applicationList.appID }}"
    - name: keyValueListTitle
      state:
        - name: visible
          value: "{{ applicationList.keyValueListTitleVisible }}"
`,
		"edge-configSet-item": `
scenario: edge-configSet-item

hierarchy:
  root: configItemManage
  structure:
    configItemManage:
      - topHead
      - configItemListFilter
      - configItemList
      - configItemFormModal
    topHead:
      - clusterAddButton

components:
  configItemManage:
    type: Container
  configItemList:
    type: Table
  topHead:
    type: RowContainer
    props:
      isTopHead: true
  configItemFormModal:
    type: FormModal
  clusterAddButton:
    type: Button
  configItemListFilter:
    type: ContractiveFilter

rendering:
  clusterAddButton:
    - name: configItemFormModal
      state:
        - name: visible
          value: "{{ clusterAddButton.configItemFormModalVisible }}"
        - name: formClear
          value: "{{ clusterAddButton.formClear }}"
  configItemList:
    - name: configItemFormModal
      state:
        - name: visible
          value: "{{ configItemList.configItemFormModalVisible }}"
        - name: formClear
          value: "{{ configItemList.formClear }}"
        - name: configSetItemID
          value: "{{ configItemList.configSetItemID }}"
  configItemFormModal:
    - name: configItemList
  configItemListFilter:
    - name: configItemList
      state:
        - name: searchCondition
          value: "{{ configItemListFilter.searchCondition }}"
        - name: isFirstFilter
          value: "{{ configItemListFilter.isFirstFilter }}"
`,
		"edge-site": `
scenario: edge-site

hierarchy:
  root: resourceManage
  structure:
    resourceManage:
      - topHead
      - siteNameFilter
      - siteList
      - siteFormModal
      - siteAddDrawer
    topHead:
      - siteAddButton
    siteAddDrawer:
      content: sitePreview

components:
  resourceManage:
    type: Container
  siteList:
    type: Table
  topHead:
    type: RowContainer
    props:
      isTopHead: true
  siteAddDrawer:
    type: Drawer
  siteFormModal:
    type: FormModal
  siteAddButton:
    type: Button
  sitePreview:
    type: InfoPreview
  siteNameFilter:
    type: ContractiveFilter
    props:
      delay: 1000

rendering:
  siteAddButton:
    - name: siteFormModal
      state:
        - name: visible
          value: "{{ siteAddButton.siteFormModalVisible }}"
        - name: formClear
          value: "{{ siteAddButton.formClear }}"
  siteFormModal:
    - name: siteList
  siteList:
    - name: siteFormModal
      state:
        - name: visible
          value: "{{ siteList.siteFormModalVisible }}"
        - name: siteID
          value: "{{ siteList.siteID }}"
        - name: formClear
          value: "{{ siteList.formClear }}"
    - name: sitePreview
      state:
        - name: visible
          value: "{{ siteList.sitePreviewVisible }}"
        - name: siteID
          value: "{{ siteList.siteID }}"
    - name: siteAddDrawer
      state:
        - name: visible
          value: "{{ siteList.siteAddDrawerVisible }}"
  siteNameFilter:
    - name: siteList
      state:
        - name: searchCondition
          value: "{{ siteNameFilter.searchCondition }}"
        - name: isFirstFilter
          value: "{{ siteNameFilter.isFirstFilter }}"`,
		"notify-config": `
scenario: notify-config

hierarchy:
  root: notifyConfig
  structure:
    notifyConfig:
      - notifyHead
      - notifyConfigTable
      - notifyConfigModal
    notifyHead:
      left: notifyTitle
      right: notifyAddButton

components:
  notifyConfig:
    type: Container
  notifyHead:
    type: LRContainer
  notifyTitle:
    type: Title
    Props:
      Title: "帮助您更好地组织通知项"
  notifyConfigTable:
    type: Table
  notifyConfigModal:
    type: FormModal
  notifyAddButton:
    type: Button

rendering:
  notifyConfigTable:
    - name: notifyConfigModal
      state:
        - name: "editId"
          value: "{{ notifyConfigTable.editId }}"
        - name: "operation"
          value: "{{ notifyConfigTable.operation }}"
        - name: "visible"
          value: "{{ notifyConfigTable.visible}}"
  notifyConfigModal:
    - name: notifyConfigTable
  notifyAddButton:
    - name: notifyConfigModal
      state:
        - name: "operation"
          value: "{{ notifyAddButton.operation }}"
        - name: "visible"
          value: "{{ notifyAddButton.visible }}"
        - name: "editId"
          value: "{{ notifyAddButton.editId }}"
`,
		"action": `
# 场景名
scenario: action

# 布局
hierarchy:
  root: actionForm

# 组件
components:
  actionForm:
    type: "ActionForm"
    props: "[后端动态注入]"
    data: {}
    operations:
      change:
        version:
          reload: true
    state:
      version: "[前端选择列表选择]"
`,
		"app-pipeline-tree": `
# 场景名
scenario: "app-pipeline-tree"

# 布局
hierarchy:
  root: "fileTreePage"
  structure:
    fileTreePage: ["fileTree", "nodeFormModal"]

rendering:
  # 前端触发组件
  # 先渲染前端触发组件，再渲染关联组件
  nodeFormModal:
    # 关联渲染组件列表
    - name: fileTree
      state:
        - name: "nodeFormModalAddNode"
          value: "{{ nodeFormModal.nodeFormModalAddNode }}"

# 组件
components:
  fileTreePage:
    type: "Container"
    props:
      fullHeight: true
  fileTree:
    type: "FileTree"
  nodeFormModal:
    type: "FormModal"
    state:
      visible: false
    props:
      title: "添加流水线"
      fields:
        - key: "branch"
          label: "分支"
          component: "input"
          required: true
          disabled: true
        - key: "name"
          label: "流水线名称"
          component: "input"
          required: true
          componentProps:
            maxLength: 30
    operations:
      submit:
        key: "submit"
        reload: true

`,
		"auto-test-plan-detail": `
# 场景名
scenario: auto-test-plan-detail

hierarchy:
  root: fileDetail
  structure:
    fileDetail:
      children:
        - fileConfig
        - fileExecute
      tabBarExtraContent:
        - tabExecuteButton
    fileConfig:
      - fileInfoHead
      - fileInfo
      - stagesTitle
      - stages
      - stagesOperations
    fileExecute:
      - executeHead
      - executeInfo
      - executeTaskTitle
      - executeTaskBreadcrumb
      - executeTaskTable
      - executeAlertInfo
      - envDrawer
    #      - resultDrawer
    envDrawer:
      content: envContainer
    envContainer:
      - envBaseInfoTitle
      - envBaseInfo
      - envHeaderTitle
      - envHeaderInfo
      - envGlobalTitle
      - envGlobalInfo
    envHeaderInfo:
      children:
        - envHeaderTable
        - envHeaderText
    envGlobalInfo:
      children:
        - envGlobalTable
        - envGlobalText
    fileInfoHead:
      left: fileInfoTitle
#      right: fileHistory
#    fileHistory:
#      children: fileHistoryButton
#      content: fileHistoryTable
    stagesOperations:
      - addScenesSetButton
      - scenesSetDrawer
    scenesSetDrawer:
      content:
        - scenesSetSelect
        - scenesSetInParams
    executeHead:
      left: executeInfoTitle
      right:
        - refreshButton
        - cancelExecuteButton
        - executeHistory
  #  resultDrawer:
  #    content: resultPreview
    executeHistory:
      children: executeHistoryButton
      content: executeHistoryPop
    executeHistoryPop:
      - executeHistoryRefresh
      - executeHistoryTable
components:
  envHeaderInfo:
    type: Tabs
  envHeaderTitle:
    type: Title
  envHeaderTable:
    type: Table
  envHeaderText:
    type: FileEditor
  envContainer:
    type: Container
  envDrawer:
    type: Drawer
  envBaseInfo:
    type: Panel
  envBaseInfoTitle:
    type: Title
  envGlobalInfo:
    type: Tabs
  envGlobalTitle:
    type: Title
  envGlobalTable:
    type: Table
  envGlobalText:
    type: FileEditor
  fileConfig:
    type: Container
  fileExecute:
    type: Container
  fileInfoHead:
    type: LRContainer
  executeHead:
    type: LRContainer
  executeInfoTitle:
    type: Title
  fileInfoTitle:
    type: Title
  fileHistory:
    type: Popover
  fileHistoryButton:
    type: Button
  stagesTitle:
    type: Title
  stagesOperations:
    type: RowContainer
  addScenesSetButton:
    type: Button
  resultDrawer:
    type: Drawer
  scenesSetDrawer:
    type: Drawer
  refreshButton:
    type: Button
  cancelExecuteButton:
    type: Button
  executeHistory:
    type: Popover
    props:
      placement: "bottomRight"
      size: "l"
      title: ${{ i18n.autotest.plan.execute.history }}
      trigger: "click"
  executeHistoryButton:
    type: Button
  executeHistoryPop:
    type: Container
  executeHistoryRefresh:
    type: Button
  executeTaskTitle:
    type: Title
  executeTaskTable:
    type: Table
  executeTaskBreadcrumb:
    type: Breadcrumb
  executeAlertInfo:
    type: Alert
  fileDetail:
    type: Tabs
  tabExecuteButton:
    type: Button
  fileHistoryTable:
    type: Table
  fileInfo:
    type: Panel
  stages:
    type: SortGroup
  executeHistoryTable:
    type: Table
  executeInfo:
    type: Panel
  resultPreview:
    type: InfoPreview
  scenesSetSelect:
    type: TreeSelect
  scenesSetInParams:
    type: Form
rendering:
  addScenesSetButton:
    - name: stages
    - name: scenesSetDrawer
      state:
        - name: "visible"
          value: "{{ addScenesSetButton.showScenesSetDrawer }}"
    - name: scenesSetSelect
      state:
        - name: "visible"
          value: "{{ scenesSetDrawer.visible }}"
        - name: "testPlanStepId"
          value: "{{ addScenesSetButton.testPlanStepId }}"
    - name: scenesSetInParams
      state:
        - name: "testPlanStepId"
          value: "{{ addScenesSetButton.testPlanStepId }}"
  stages:
    - name: scenesSetDrawer
      state:
        - name: "visible"
          value: "{{ stages.showScenesSetDrawer }}"
    - name: scenesSetSelect
      state:
        - name: "visible"
          value: "{{ stages.showScenesSetDrawer }}"
        - name: "testPlanStepId"
          value: "{{ stages.stepId }}"
    - name: scenesSetInParams
      state:
        - name: scenesSetId
          value: "{{ scenesSetSelect.scenesSetId }}"
        - name: "testPlanStepId"
          value: "{{ stages.stepId }}"
  scenesSetSelect:
    - name: scenesSetInParams
      state:
        - name: scenesSetId
          value: "{{ scenesSetSelect.scenesSetId }}"
        - name: "testPlanStepId"
          value: "{{ scenesSetSelect.testPlanStepId }}"
  scenesSetInParams:
    - name: scenesSetDrawer
      state:
        - name: "visible"
          value: "{{ scenesSetInParams.visible }}"
    - name: stages
  fileDetail:
    - name: executeHead
    - name: executeHistoryTable
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeTaskTitle
    - name: executeTaskBreadcrumb
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: cancelExecuteButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: "pipelineDetail"
          value: "{{ executeAlertInfo.pipelineDetail }}"
    - name: executeHistoryButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
    - name: refreshButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeInfoTitle
    - name: executeHistoryRefresh
    - name: envDrawer
    - name: envBaseInfoTitle
    - name: envBaseInfo
    - name: envGlobalTitle
    - name: envGlobalInfo
    - name: envGlobalTable
    - name: envGlobalText
    - name: envHeaderTitle
    - name: envHeaderTable
    - name: envHeaderInfo
    - name: envHeaderText
  executeHistoryTable:
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: "envName"
          value: "{{ executeHistoryTable.envName }}"
        - name: "envData"
          value: "{{ executeHistoryTable.envData }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: refreshButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: cancelExecuteButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: "pipelineDetail"
          value: "{{ executeAlertInfo.pipelineDetail }}"
    - name: executeTaskBreadcrumb
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: envDrawer
    - name: envBaseInfoTitle
    - name: envBaseInfo
    - name: envGlobalTitle
    - name: envGlobalInfo
    - name: envGlobalTable
    - name: envGlobalText
    - name: envHeaderTitle
    - name: envHeaderTable
    - name: envHeaderInfo
    - name: envHeaderText
  executeTaskTable:
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeTaskTable.pipelineId }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeTaskBreadcrumb
      state:
        - name: "name"
          value: "{{ executeTaskTable.name }}"
        - name: "pipelineId"
          value: "{{ executeTaskTable.pipelineId }}"
        - name: "unfold"
          value: "{{ executeTaskTable.unfold }}"
    - name: executeHistoryButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
    - name: refreshButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
    # - name: cancelExecuteButton
    #   state:
    #     - name: "visible"
    #       value: "{{ executeTaskBreadcrumb.visible }}"
  executeTaskBreadcrumb:
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeTaskBreadcrumb.pipelineId }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeHistoryButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
    - name: refreshButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: cancelExecuteButton
      state:
        # - name: "visible"
        #   value: "{{ executeTaskBreadcrumb.visible }}"
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: "pipelineDetail"
          value: "{{ executeAlertInfo.pipelineDetail }}"
  executeHistoryRefresh:
    - name: executeHistoryTable
      state:
        - name: pageNo
          value: "{{ executeHistoryRefresh.pageNo }}"
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: envDrawer
    - name: envBaseInfoTitle
    - name: envBaseInfo
    - name: envGlobalTitle
    - name: envGlobalInfo
    - name: envGlobalTable
    - name: envGlobalText
    - name: envHeaderTitle
    - name: envHeaderTable
    - name: envHeaderInfo
    - name: envHeaderText
  refreshButton:
    - name: executeHistoryTable
    - name: executeInfo
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeAlertInfo
    - name: cancelExecuteButton
      state:
        - name: "pipelineId"
          value: "{{ executeAlertInfo.pipelineId }}"
        - name: "pipelineDetail"
          value: "{{ executeAlertInfo.pipelineDetail }}"
    - name: envDrawer
    - name: envBaseInfoTitle
    - name: envBaseInfo
    - name: envGlobalTitle
    - name: envGlobalInfo
    - name: envGlobalTable
    - name: envGlobalText
    - name: envHeaderTitle
    - name: envHeaderTable
    - name: envHeaderInfo
    - name: envHeaderText
  cancelExecuteButton:
    - name: executeHistoryTable
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: refreshButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
  tabExecuteButton:
    - name: executeHistoryTable
    - name: executeInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: "envName"
          value: "{{ executeHistoryTable.envName }}"
        - name: "envData"
          value: "{{ executeHistoryTable.envData }}"
    - name: executeTaskTable
      state:
        - name: "pipelineDetail"
          value: "{{ executeInfo.pipelineDetail }}"
    - name: executeAlertInfo
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: cancelExecuteButton
      state:
        - name: pipelineId
          value: "{{ executeHistoryTable.pipelineId }}"
        - name: pipelineDetail
          value: "{{ executeAlertInfo.pipelineDetail }}"
    - name: executeTaskBreadcrumb
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: fileDetail
      state:
        - name: activeKey
          value: "{{ tabExecuteButton.activeKey }}"
    - name: refreshButton
      state:
        - name: "pipelineId"
          value: "{{ executeHistoryTable.pipelineId }}"
    - name: executeHistoryButton
      state:
        - name: "visible"
          value: "{{ executeTaskBreadcrumb.visible }}"
    - name: envDrawer
    - name: envBaseInfoTitle
    - name: envBaseInfo
    - name: envGlobalTitle
    - name: envGlobalInfo
    - name: envGlobalTable
    - name: envGlobalText
    - name: envHeaderTitle
    - name: envHeaderTable
    - name: envHeaderInfo
    - name: envHeaderText
  __DefaultRendering__:
    - name: fileDetail
    - name: fileConfig
      state:
        - name: activeKey
          value: "{{ fileDetail.activeKey }}"
    - name: fileInfoHead
    - name: fileInfoTitle
    - name: fileInfo
      state:
        - name: "testPlanId"
          value: "{{ fileDetail.testPlanId }}"
        - name: "visible"
          value: "{{ fileConfig.visible }}"
    - name: stagesTitle
    - name: stages
      state:
        - name: "testPlanId"
          value: "{{ fileDetail.testPlanId }}"
        - name: "visible"
          value: "{{ fileConfig.visible }}"
    - name: stagesOperations
    - name: tabExecuteButton
      state:
        - name: "testPlanId"
          value: "{{ fileDetail.testPlanId }}"
    - name: addScenesSetButton
      state:
        - name: "testPlanId"
          value: "{{ fileDetail.testPlanId }}"
`,
		"org-list-my": `
# 场景名
scenario: "org-list-my"

# 布局
hierarchy:
  root: page
  structure:
    page: 
      children:
        - myPage
        - empty
      tabBarExtraContent:
        - createButton
    myPage:
      - filter
      - list
      - emptyContainer
    emptyContainer: 
      - emptyText

components:
  page:
    type: Tabs
  myPage:
    type: Container
  createButton:
    type: Button
  filter:
    type: ContractiveFilter
  list:
    type: List
  emptyContainer:
    type: RowContainer
  emptyText:
    type: Text
  
rendering:
  filter:
    - name: list
      state:
        - name: "searchEntry"
          value: "{{ filter.searchEntry }}"
        - name: "searchRefresh"
          value: "{{ filter.searchRefresh }}"
  list:
    - name: emptyContainer
      state:
        - name: "isEmpty"
          value: "{{ list.isEmpty }}"
    - name: emptyText
      state:
        - name: "isEmpty"
          value: "{{ list.isEmpty }}"

  __DefaultRendering__:
    - name: page
    - name: myPage
    - name: createButton
    - name: filter
    - name: list
    - name: emptyContainer
      state:
        - name: "isEmpty"
          value: "{{ list.isEmpty }}"
    - name: emptyText
      state:
        - name: "isEmpty"
          value: "{{ list.isEmpty }}"
    
`,
	}
	for pName, pStr := range protocols {
		var p apistructs.ComponentProtocol
		if err := yaml.Unmarshal([]byte(pStr), &p); err != nil {
			panic(err)
		}
		protocol.DefaultProtocols[pName] = p
		if protocol.CpPlaceHolderRe.Match([]byte(pStr)) {
			protocol.DefaultProtocolsRaw[pName] = pStr
		}
	}
}
