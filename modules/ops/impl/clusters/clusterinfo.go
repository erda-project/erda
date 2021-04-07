package clusters

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func (c *Clusters) ClusterInfo(ctx context.Context, clusterNames []string) ([]map[string]map[string]apistructs.NameValue, error) {
	i18n := ctx.Value("i18nPrinter").(*message.Printer)
	resultList := []map[string]map[string]apistructs.NameValue{}
	for _, clusterName := range clusterNames {
		clusterinfo, err := c.bdl.QueryClusterInfo(clusterName)
		if err != nil {
			errstr := fmt.Sprintf("failed to queryclusterinfo: %v, cluster: %v", err, clusterName)
			logrus.Errorf(errstr)
			continue
		}

		masterNum := len(strutil.Split(clusterinfo.Get(apistructs.MASTER_ADDR), ",", true))
		lbNum := len(strutil.Split(clusterinfo.Get(apistructs.LB_ADDR), ",", true))

		result := map[string]map[string]apistructs.NameValue{
			"basic": {
				"clusterName":    apistructs.NameValue{Name: i18n.Sprintf("cluster name"), Value: clusterinfo.Get(apistructs.DICE_CLUSTER_NAME)},
				"clusterType":    apistructs.NameValue{Name: i18n.Sprintf("cluster type"), Value: clusterinfo.Get(apistructs.DICE_CLUSTER_TYPE)},
				"clusterVersion": apistructs.NameValue{Name: i18n.Sprintf("cluster version"), Value: clusterinfo.Get(apistructs.DICE_VERSION)},
				"rootDomain":     apistructs.NameValue{Name: i18n.Sprintf("root domain"), Value: clusterinfo.Get(apistructs.DICE_ROOT_DOMAIN)},
				"edgeCluster":    apistructs.NameValue{Name: i18n.Sprintf("edge cluster"), Value: clusterinfo.Get(apistructs.DICE_IS_EDGE) == "true"},
				"masterNum":      apistructs.NameValue{Name: i18n.Sprintf("master num"), Value: masterNum},
				"lbNum":          apistructs.NameValue{Name: i18n.Sprintf("lb num"), Value: lbNum},
				"httpsEnabled": apistructs.NameValue{
					Name:  i18n.Sprintf("https enabled"),
					Value: strutil.Contains(clusterinfo.Get(apistructs.DICE_PROTOCOL), "https")},
			},
			"URLs": {
				"registry": apistructs.NameValue{Name: "registry", Value: clusterinfo.Get(apistructs.REGISTRY_ADDR)},
				"nexus":    apistructs.NameValue{Name: "nexus", Value: clusterinfo.Get(apistructs.NEXUS_ADDR)},
				"masters":  apistructs.NameValue{Name: "masters", Value: clusterinfo.Get(apistructs.MASTER_ADDR)},
				"lb":       apistructs.NameValue{Name: "lb", Value: clusterinfo.Get(apistructs.LB_ADDR)},
			},
		}
		resultList = append(resultList, result)
	}
	return resultList, nil
}
