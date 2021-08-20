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

package clusters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/conf"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	statusPending         = "pending"
	statusOnline          = "online"
	statusOffline         = "offline"
	statusInitializing    = "initializing"
	statusInitializeError = "initialize error"
	statusUnknown         = "unknown"
)

var (
	diceOperator = "/apis/dice.terminus.io/v1beta1/namespaces/%s/dices/dice"
	erdaOperator = "/apis/erda.terminus.io/v1beta1/namespaces/%s/erdas/erda"
	checkCRDs    = []string{erdaOperator, diceOperator}
)

func (c *Clusters) ClusterInfo(ctx context.Context, orgID uint64, clusterNames []string) ([]map[string]map[string]apistructs.NameValue, error) {
	i18n := ctx.Value("i18nPrinter").(*message.Printer)
	resultList := make([]map[string]map[string]apistructs.NameValue, 0)

	for _, clusterName := range clusterNames {
		clusterMetaData, err := c.bdl.GetCluster(clusterName)
		if err != nil {
			logrus.Error(err)
			continue
		}

		baseInfo := map[string]apistructs.NameValue{
			"manageType":         {Name: i18n.Sprintf("manage type"), Value: parseManageType(clusterMetaData.ManageConfig)},
			"clusterName":        {Name: i18n.Sprintf("cluster name"), Value: clusterName},
			"clusterDisplayName": {Name: i18n.Sprintf("cluster display name"), Value: clusterMetaData.DisplayName},
			"initJobClusterName": {Name: i18n.Sprintf("init job cluster name"), Value: os.Getenv("DICE_CLUSTER_NAME")},
		}

		urlInfo := map[string]apistructs.NameValue{}

		if ci, err := c.bdl.QueryClusterInfo(clusterName); err != nil {
			errStr := fmt.Sprintf("failed to queryclusterinfo: %v, cluster: %v", err, clusterName)
			logrus.Errorf(errStr)
		} else {
			baseInfo["clusterType"] = apistructs.NameValue{Name: i18n.Sprintf("cluster type"), Value: ci.Get(apistructs.DICE_CLUSTER_TYPE)}
			baseInfo["clusterVersion"] = apistructs.NameValue{Name: i18n.Sprintf("cluster version"), Value: ci.Get(apistructs.DICE_VERSION)}
			baseInfo["rootDomain"] = apistructs.NameValue{Name: i18n.Sprintf("root domain"), Value: ci.Get(apistructs.DICE_ROOT_DOMAIN)}
			baseInfo["edgeCluster"] = apistructs.NameValue{Name: i18n.Sprintf("edge cluster"), Value: ci.Get(apistructs.DICE_IS_EDGE) == "true"}
			baseInfo["masterNum"] = apistructs.NameValue{
				Name:  i18n.Sprintf("master num"),
				Value: len(strutil.Split(ci.Get(apistructs.MASTER_ADDR), ",", true)),
			}
			baseInfo["lbNum"] = apistructs.NameValue{
				Name:  i18n.Sprintf("lb num"),
				Value: len(strutil.Split(ci.Get(apistructs.LB_ADDR), ",", true)),
			}
			baseInfo["httpsEnabled"] = apistructs.NameValue{
				Name:  i18n.Sprintf("https enabled"),
				Value: strutil.Contains(ci.Get(apistructs.DICE_PROTOCOL), "https"),
			}

			urlInfo["registry"] = apistructs.NameValue{Name: "registry", Value: ci.Get(apistructs.REGISTRY_ADDR)}
			urlInfo["nexus"] = apistructs.NameValue{Name: "nexus", Value: ci.Get(apistructs.NEXUS_ADDR)}
			urlInfo["masters"] = apistructs.NameValue{Name: "masters", Value: ci.Get(apistructs.MASTER_ADDR)}
			urlInfo["lb"] = apistructs.NameValue{Name: "lb", Value: ci.Get(apistructs.LB_ADDR)}
		}

		kc, err := k8sclient.NewWithTimeOut(clusterName, getClusterTimeout)
		if err != nil {
			logrus.Errorf("get k8sclient error: %v", err)
			result := map[string]map[string]apistructs.NameValue{
				"basic": baseInfo,
				"URLs":  urlInfo,
			}

			resultList = append(resultList, result)
			continue
		}

		nodes, err := kc.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			logrus.Error(err)
		}
		baseInfo["nodeCount"] = apistructs.NameValue{Name: i18n.Sprintf("node count"), Value: len(nodes.Items)}

		if status, err := c.getClusterStatus(kc, clusterMetaData); err != nil {
			logrus.Errorf("get cluster status error: %v", err)
		} else {
			baseInfo["clusterStatus"] = apistructs.NameValue{Name: i18n.Sprintf("cluster status"), Value: status}
		}

		result := map[string]map[string]apistructs.NameValue{
			"basic": baseInfo,
			"URLs":  urlInfo,
		}

		resultList = append(resultList, result)
	}
	return resultList, nil
}

func (c *Clusters) getClusterStatus(kc *k8sclient.K8sClient, meta *apistructs.ClusterInfo) (string, error) {
	if kc == nil || kc.ClientSet == nil {
		return "", fmt.Errorf("kubernetes client is nil")
	}

	// if manage config is nil, cluster import with inet or other
	if meta.ManageConfig == nil {
		return statusOffline, nil
	}

	switch meta.ManageConfig.Type {
	case apistructs.ManageProxy:
		if meta.ManageConfig.Token == "" || meta.ManageConfig.Address == "" {
			return statusPending, nil
		}
	case apistructs.ManageToken:
		if meta.ManageConfig.Token == "" || meta.ManageConfig.Address == "" {
			return statusOffline, nil
		}
	case apistructs.ManageCert:
		if meta.ManageConfig.CertData == "" && meta.ManageConfig.KeyData == "" {
			return statusOffline, nil
		}
	}

	ec := &apistructs.DiceCluster{}

	var (
		res           []byte
		err           error
		resourceExist bool
	)

	for _, selfLink := range checkCRDs {
		res, err = kc.ClientSet.RESTClient().Get().
			AbsPath(fmt.Sprintf(selfLink, conf.ErdaNamespace())).
			DoRaw(context.Background())
		if err != nil {
			logrus.Error(err)
			continue
		}
		resourceExist = true
		break
	}

	if !resourceExist {
		return statusUnknown, nil
	}

	if err = json.Unmarshal(res, &ec); err != nil {
		logrus.Errorf("unmarsharl data error, data: %v", string(res))
		return statusUnknown, err
	}

	switch ec.Status.Phase {
	case apistructs.ClusterPhaseRunning:
		return statusOnline, nil
	case apistructs.ClusterPhaseNone, apistructs.ClusterPhaseCreating, apistructs.ClusterPhaseInitJobs,
		apistructs.ClusterPhasePending, apistructs.ClusterPhaseUpdating:
		return statusInitializing, nil
	case apistructs.ClusterPhaseFailed:
		return statusInitializeError, nil
	default:
		return statusUnknown, nil
	}
}

func parseManageType(mc *apistructs.ManageConfig) string {
	if mc == nil {
		return "create"
	}

	switch mc.Type {
	case apistructs.ManageProxy:
		return "agent"
	case apistructs.ManageToken, apistructs.ManageCert:
		return "import"
	default:
		return "create"
	}
}
