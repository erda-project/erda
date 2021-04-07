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

package edge

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/dbclient"
	"github.com/erda-project/erda/pkg/clientgo/apis/openyurt/v1alpha1"
	"github.com/erda-project/erda/pkg/httpclient"
)

// ListSite 获取全部边缘站点列表
func (e *Edge) ListSite(param *apistructs.EdgeSiteListPageRequest) (int, *[]apistructs.EdgeSiteInfo, error) {
	var (
		clusterIDs []int64
	)

	total, sites, err := e.db.ListEdgeSite(param)
	if err != nil {
		return 0, nil, err
	}

	if param.ClusterID > 0 {
		clusterIDs = []int64{param.ClusterID}
	} else {
		clusterIDs = e.getSitesClusterIDs(sites)
	}

	clusterNodePools, err := e.getNodePoolsByClusters(clusterIDs)
	if err != nil {
		return 0, nil, err
	}

	siteInfos := make([]apistructs.EdgeSiteInfo, 0, len(*sites))

	for i := range *sites {
		nodePoolMap := clusterNodePools[(*sites)[i].ClusterID]
		if nodePool, ok := nodePoolMap[(*sites)[i].Name]; ok {
			clusterInfo, err := e.getClusterInfo((*sites)[i].ClusterID)
			if err != nil {
				return 0, nil, err
			}
			item := convertToEdgeSiteInfo(&(*sites)[i], getFormatNodeNum(*nodePool), clusterInfo.Name)
			siteInfos = append(siteInfos, *item)
		} else {
			return 0, nil, fmt.Errorf("node pool %s not found in cluster(id: %d)", (*sites)[i].Name, (*sites)[i].ClusterID)
		}
	}

	return total, &siteInfos, nil
}

// GetEdgeSite
func (e *Edge) GetEdgeSite(edgeSiteID int64) (*apistructs.EdgeSiteInfo, error) {
	var (
		siteInfo *apistructs.EdgeSiteInfo
	)

	site, err := e.db.GetEdgeSite(edgeSiteID)
	if err != nil {
		return nil, err
	}

	clusterNodePool, err := e.getNodePoolsByCluster(site.ClusterID)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := e.getClusterInfo(site.ClusterID)
	if err != nil {
		return nil, err
	}

	siteInfo = convertToEdgeSiteInfo(site, getFormatNodeNum(*clusterNodePool[site.Name]), clusterInfo.Name)

	return siteInfo, nil
}

// CreateSite 创建边缘站点
func (e *Edge) CreateSite(req *apistructs.EdgeSiteCreateRequest) (uint64, error) {
	// 根据 origin id 以及 cluster id 可以创建唯一的站点，site name 作为 cluster 下的唯一标示，可以扩展到 origin 唯一
	// TODO: 创建流程事务
	var (
		// DefaultLabelKey: [clusterID].[orgName]
		// s.g. terminus-dev.terminus
		defaultLabelKey = "%s.%s"
		// Labels: s.g. beijing.terminus-dev.terminus: beijing
		labels             = make(map[string]string, 0)
		nodePoolAPIVersion = "apps.openyurt.io/v1alpha1"
		nodePoolKind       = "NodePool"
		// NodePoolType: Edge, Cloud
		edgeNodePoolType v1alpha1.NodePoolType = "Edge"
	)

	edgeSite := &dbclient.EdgeSite{
		OrgID:       req.OrgID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Logo:        req.Logo,
		ClusterID:   req.ClusterID,
		Status:      req.Status,
	}

	// TODO: Selector, Taints 自定义； 其他字段可操作性
	// 先创建对应的 node pool 如果创建成功，则插入数据
	nodePool := &v1alpha1.NodePool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: nodePoolAPIVersion,
			Kind:       nodePoolKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		Spec: v1alpha1.NodePoolSpec{
			Type:        edgeNodePoolType,
			Selector:    nil,
			Labels:      nil,
			Annotations: nil,
			Taints:      nil,
		},
	}

	clusterInfo, err := e.getClusterInfo(req.ClusterID)
	if err != nil {
		return 0, err
	}

	orgInfo, err := e.getOrgInfo(req.OrgID)
	if err != nil {
		return 0, err
	}

	labels[fmt.Sprintf(defaultLabelKey, clusterInfo.Name, orgInfo.Name)] = req.Name
	nodePool.Spec.Labels = labels
	nodePool.Spec.Annotations = labels

	_, err = e.k8s.CreateNodePool(clusterInfo.Name, nodePool)

	if err != nil {
		return 0, err
	}

	if err = e.db.CreateEdgeSite(edgeSite); err != nil {
		deleteErr := e.k8s.DeleteNodePool(clusterInfo.Name, req.Name)
		if deleteErr != nil {
			return 0, deleteErr
		}
		return 0, err
	}

	return edgeSite.ID, nil
}

// UpdateSite 更新边缘站点
func (e *Edge) UpdateSite(edgeSiteID int64, req *apistructs.EdgeSiteUpdateRequest) error {

	edgeSite, err := e.db.GetEdgeSite(edgeSiteID)
	if err != nil {
		return err
	}

	edgeSite.DisplayName = req.DisplayName
	edgeSite.Description = req.Description
	edgeSite.Logo = req.Logo
	edgeSite.Status = req.Status

	if err = e.db.UpdateEdgeSite(edgeSite); err != nil {
		return err
	}

	return nil
}

// Delete 删除边缘站点
func (e *Edge) DeleteSite(edgeSiteID int64) error {
	edgeSite, err := e.db.GetEdgeSite(edgeSiteID)
	if err != nil || edgeSite == nil {
		return fmt.Errorf("failed to get edgesite, (%v)", err)
	}

	clusterInfo, err := e.getClusterInfo(edgeSite.ClusterID)
	if err != nil {
		return fmt.Errorf("get cluster (id: %d) error: %v", edgeSite.ClusterID, err)
	}

	np, err := e.k8s.GetNodePool(clusterInfo.Name, edgeSite.Name)
	if err != nil {
		return fmt.Errorf("get nodepool %s in cluster %s error: %v", edgeSite.Name, clusterInfo.Name, err)
	}

	if len(np.Status.Nodes) > 0 {
		return fmt.Errorf("can't delete noempty nodepool, offline it first %s", fmt.Sprint(np.Status.Nodes))
	}

	// 删除站点，如果站点存在使用的应用，则不允许删除，并提示用户
	edgeSites, err := e.db.GetEdgeAppsBySiteName(edgeSite.Name, edgeSite.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to get releated edge app error: %v", err)
	}

	if len(*edgeSites) != 0 {
		return fmt.Errorf("failed to delete site, "+
			"site %s have releated edge app, please delete it first", edgeSite.Name)
	}

	// 如果存在对应的配置项，删除掉所有配置项, 并告知通知用户
	cfgItem, err := e.db.GetEdgeConfigSetItemsBySiteID(edgeSiteID)
	if err != nil {
		return fmt.Errorf("failed to get releated configset item in site: %s", edgeSite.Name)
	}

	rmItemError := ""

	for _, item := range *cfgItem {
		if err = e.DeleteConfigSetItem(int64(item.ID)); err != nil {
			rmItemError += err.Error()
		}
	}

	if err = e.db.DeleteEdgeSite(edgeSiteID); err != nil {
		return fmt.Errorf("failed to delete edgesite, (%v)", err)
	}

	err = e.k8s.DeleteNodePool(clusterInfo.Name, edgeSite.Name)

	if err != nil {
		return err
	}

	if rmItemError != "" {
		return fmt.Errorf("delete releated config item error: %v, please contact administrator", rmItemError)
	}

	return nil
}

// GetInitSiteShell 获取节点初始化脚本
func (e *Edge) GetInitSiteShell(edgeSiteID int64) (map[string][]string, error) {
	var (
		result           = make(map[string][]string)
		initNodeShell    = make([]string, 0)
		shellDownTmpl    = "wget %s/%s && chmod 755 %s"
		shellExeTmpl     = "./%s install -c %s -p %s -s %s -t %s"
		sysNamespace     = "kube-system"
		defaultNamespace = "default"
		diceClusterCM    = "dice-cluster-info"
		secName          = "bootstrap-token-teedge"
		tokenKey         = "token-secret"
		shellNameTmpl    = "initnode_%s"
		tokenTmpl        = "teedge.%s"
	)
	edgeSite, err := e.db.GetEdgeSite(edgeSiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get edgesite, (%v)", err)
	}

	clusterInfo, err := e.getClusterInfo(edgeSite.ClusterID)
	if err != nil {
		return nil, err
	}

	teEdgeData, err := e.k8s.GetSecret(clusterInfo.Name, sysNamespace, secName)
	if err != nil {
		return nil, err
	}

	clusterInfoCM, err := e.k8s.GetConfigMap(clusterInfo.Name, defaultNamespace, diceClusterCM)
	if err != nil {
		return nil, fmt.Errorf("get dice cluster info error: %v", err)
	}

	cloudAddr := clusterInfoCM.Data["EDGE_CLOUD_ADDR"]
	cloudPort := clusterInfoCM.Data["EDGE_CLOUD_PORT"]
	toolAddr := clusterInfoCM.Data["EDGE_TOOL_ADDR"]

	shellName := fmt.Sprintf(shellNameTmpl, clusterInfo.Name)
	token := fmt.Sprintf(tokenTmpl, string(teEdgeData.Data[tokenKey]))

	initNodeShell = append(initNodeShell, fmt.Sprintf(shellDownTmpl, toolAddr, shellName, shellName))
	initNodeShell = append(initNodeShell, fmt.Sprintf(shellExeTmpl, shellName, cloudAddr, cloudPort, edgeSite.Name, token))

	result[edgeSite.Name] = initNodeShell

	return result, nil
}

// OfflineEdgeHost 下线
func (e *Edge) OfflineEdgeHost(edgeSiteID int64, siteIP string) error {
	var (
		nodeFormatter  = "edgenode-%s-%s"
		appAddonType   = "addon"
		appNsFormatter = "edgeapp-%s"
		stsLabels      = "statefulset.kubernetes.io/pod-name"
	)

	edgeSite, err := e.db.GetEdgeSite(edgeSiteID)
	if err != nil {
		return fmt.Errorf("get edge site (id: %d) error: %v", edgeSiteID, err)
	}

	clusterInfo, err := e.getClusterInfo(edgeSite.ClusterID)
	if err != nil {
		return fmt.Errorf("get cluster (id: %d) info error: %v", edgeSite.ClusterID, err)
	}

	edgeApps, err := e.db.ListEdgeAppBySiteName(edgeSite.OrgID, edgeSite.ClusterID, edgeSite.Name)
	if err != nil {
		return fmt.Errorf("get reference apps in site %s eror: %v", edgeSite.Name, err)
	}

	tipsStsNames := make([]string, 0)
	tipsDeployNames := make([]string, 0)

	for _, edgeApp := range *edgeApps {
		// Only mysql-edge is sts type in this version.
		if edgeApp.Type != appAddonType {
			tipsDeployNames = append(tipsDeployNames, edgeApp.Name)
			continue
		}

		nsName := fmt.Sprintf(appNsFormatter, edgeApp.Name)
		podList, err := e.k8s.ListPod(clusterInfo.Name, nsName)
		if err != nil {
			return fmt.Errorf("list pod in namespaces %s error: %v", nsName, err)
		}

		for _, po := range podList.Items {
			if _, ok := po.Labels[stsLabels]; !ok {
				continue
			}
			//if po.Status.Phase == v1.PodRunning && po.Status.HostIP == siteIP {
			if po.Status.HostIP == siteIP {
				tipsStsNames = append(tipsStsNames, edgeApp.Name)
			}
		}
	}

	if len(tipsStsNames) != 0 {
		return fmt.Errorf("there are statefulset app %s in site: %s", fmt.Sprint(tipsStsNames), edgeSite.Name)
	}

	if len(tipsDeployNames) != 0 {
		np, err := e.k8s.GetNodePool(clusterInfo.Name, edgeSite.Name)
		if err != nil {
			return fmt.Errorf("get node pool error: %v", err)
		}
		if len(np.Status.Nodes) <= 1 {
			return fmt.Errorf("there are applications %s on the last host", fmt.Sprint(tipsDeployNames))
		}
	}

	nodeName := fmt.Sprintf(nodeFormatter, edgeSite.Name, siteIP)

	node, err := e.k8s.GetNode(clusterInfo.Name, nodeName)
	if err != nil {
		return err
	}

	delete(node.Labels, v1alpha1.LabelDesiredNodePool)

	if err = e.k8s.UpdateNode(clusterInfo.Name, node); err != nil {
		return fmt.Errorf("update node labels error: %v", err)
	}

	if err = e.k8s.DeleteNode(clusterInfo.Name, nodeName); err != nil {
		return fmt.Errorf("delete node %s eror: %v", nodeName, err)
	}

	// delete monitor data.
	if err = cleanOfflineData(clusterInfo.Name, siteIP); err != nil {
		logrus.Errorf("clean monitor data %s error: %v", siteIP, err)
	}

	return nil
}

// getNodePoolsByClusters 获取多个集群下所有 node pool 信息
func (e *Edge) getNodePoolsByClusters(clusterIDs []int64) (map[int64]NodePools, error) {
	var (
		clustersNodePools = make(map[int64]NodePools, 0)
	)
	for _, clusterID := range clusterIDs {
		nodePools, err := e.getNodePoolsByCluster(clusterID)
		if err != nil {
			return clustersNodePools, fmt.Errorf("get node pool (cluster: %v) error, %v", clusterID, err)
		}
		clustersNodePools[clusterID] = nodePools
	}
	return clustersNodePools, nil
}

// getNodePoolsByCluster 获取集群下所有 node pool 信息
func (e *Edge) getNodePoolsByCluster(clusterID int64) (NodePools, error) {
	var (
		nodePoolList *v1alpha1.NodePoolList
		nodePools    = make(NodePools, 0)
	)

	clusterInfo, err := e.getClusterInfo(clusterID)
	if err != nil {
		return nil, err
	}

	nodePoolList, err = e.k8s.ListNodePool(clusterInfo.Name)
	if err != nil {
		return nodePools, err
	}

	return convertNPListToMap(nodePoolList), nil
}

// getSitesClusterIDs 获取多集群站点的 cluster id
func (e *Edge) getSitesClusterIDs(sites *[]dbclient.EdgeSite) []int64 {
	var (
		tmpMap     = make(map[int64]struct{}, 0)
		clusterIDs = make([]int64, 0, len(*sites))
	)

	if sites == nil {
		return clusterIDs
	}

	for _, item := range *sites {
		if _, ok := tmpMap[item.ClusterID]; !ok {
			tmpMap[item.ClusterID] = struct{}{}
			clusterIDs = append(clusterIDs, item.ClusterID)
		}
	}

	return clusterIDs
}

// 将 edgeSite 存储结构转换为API所需结构
func convertToEdgeSiteInfo(edgeSite *dbclient.EdgeSite, nodeCount, clusterName string) *apistructs.EdgeSiteInfo {
	return &apistructs.EdgeSiteInfo{
		ID:          int64(edgeSite.ID),
		OrgID:       edgeSite.OrgID,
		Name:        edgeSite.Name,
		DisplayName: edgeSite.DisplayName,
		Description: edgeSite.Description,
		Logo:        edgeSite.Logo,
		ClusterID:   edgeSite.ClusterID,
		ClusterName: clusterName,
		Status:      edgeSite.Status,
		NodeCount:   nodeCount,
		CreatedAt:   edgeSite.CreatedAt,
		UpdatedAt:   edgeSite.UpdatedAt,
	}
}

// convertNPListToMap 转换 node pool list 类型为以 node pool name 为 key 的 map
func convertNPListToMap(nodePoolList *v1alpha1.NodePoolList) map[string]*v1alpha1.NodePool {
	var (
		res = make(map[string]*v1alpha1.NodePool, 0)
	)

	if nodePoolList == nil || len(nodePoolList.Items) == 0 {
		logrus.Errorf("can't convert type NodePoolList to Map, NodePoolList is nil.")
		return res
	}
	for index := range nodePoolList.Items {
		res[nodePoolList.Items[index].Name] = &nodePoolList.Items[index]
	}

	return res
}

// getFormatNodeNum format node num.
func getFormatNodeNum(nodePool v1alpha1.NodePool) string {
	return fmt.Sprintf(siteNodeCountFormat, nodePool.Status.ReadyNodeNum, nodePool.Status.ReadyNodeNum+nodePool.Status.UnreadyNodeNum)
}

// cleanOfflineData clean offline host data
func cleanOfflineData(clusterName, hostIP string) error {
	const (
		reqUrl     = "/api/resources/hosts/actions/offline"
		monitorEnv = "MONITOR_ADDR"
	)

	resp, err := httpclient.New().Post(os.Getenv(monitorEnv)).
		Path(reqUrl).
		Header("Content-Type", "application/json").
		JSONBody(struct {
			ClusterName string   `json:"clusterName"`
			HostIPs     []string `json:"hostIPs"`
		}{
			ClusterName: clusterName,
			HostIPs:     []string{hostIP},
		}).Do().DiscardBody()

	if err != nil || !resp.IsOK() {
		logrus.Errorf("call monitor offline api failed: %+v, %+v", err, resp)
		return err
	}

	return nil
}
