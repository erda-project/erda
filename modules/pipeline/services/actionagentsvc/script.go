package actionagentsvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func RunScript(clusterInfo apistructs.ClusterInfoData, scriptName string, params map[string]string) error {

	// 如果是 edas 集群且指定了打包集群，则去对应集群下载
	jobClusterName := clusterInfo.Get(apistructs.EDASJOB_CLUSTER_NAME)
	if clusterInfo.IsEDAS() && jobClusterName != "" {
		jobClusterInfo, err := bundle.New(bundle.WithScheduler()).QueryClusterInfo(jobClusterName)
		if err != nil {
			return err
		}
		return RunScript(jobClusterInfo, scriptName, params)
	}

	return bundle.New(bundle.WithSoldier(clusterInfo.MustGetPublicURL("soldier"))).RunSoldierScript(scriptName, params)
}
