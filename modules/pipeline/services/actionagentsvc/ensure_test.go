package actionagentsvc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

var s *ActionAgentSvc

func init() {
	// db client
	dbClient, err := dbclient.New()
	if err != nil {
		panic(err)
	}
	// etcd
	js, err := jsonstore.New()
	if err != nil {
		panic(err)
	}
	etcdClient, err := etcd.New()
	if err != nil {
		panic(err)
	}
	// bundle
	os.Setenv("CMDB_ADDR", "cmdb.marathon.l4lb.thisdcos.directory:9093")
	os.Setenv("AGENT_IMAGE_FILE_PATH", "/opt/action/agent")
	bdl := bundle.New(
		bundle.WithCMDB(),
		bundle.WithOps(),
	)
	s = New(dbClient, bdl, js, etcdClient)
}

func TestActionAgentSvc_Ensure(t *testing.T) {
	agentImage := "registry.cn-hangzhou.aliyuncs.com/dice/action-agent:3.4.0-20190715-78211b9c4c"
	agentMD5 := "771821eb0aeab82dc963446a3da381aa"
	var cluster = apistructs.ClusterInfoData{
		"DICE_CLUSTER_TYPE": "k8s",
	}
	err := s.Ensure(cluster, agentImage, agentMD5)
	assert.NoError(t, err)
}
