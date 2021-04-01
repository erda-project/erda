package kubernetes

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/clientgo"
)

var (
	// 对象内存中映射关系目前只用于 ID/Name/调度器请求地址 等不可变参数
	// TODO:如果涉及其他参数，可调整为每次请求查询 / 开启定时数据内存同步 goroutine
	clusterInfos = make(map[string]*apistructs.ClusterInfo, 0)

	clientSets = make(map[string]*clientgo.ClientSet, 0)
)

type Kubernetes struct {
	bdl *bundle.Bundle
}

// Option Foobar 配置选项
type Option func(*Kubernetes)

// New Foobar service
func New(options ...Option) *Kubernetes {
	r := &Kubernetes{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(k *Kubernetes) {
		k.bdl = bdl
	}
}

// getClusterInfo 从内存中获取 或根据 cluster name 查询 cluster info
func (k *Kubernetes) getClusterInfo(clusterName string) (*apistructs.ClusterInfo, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}

	if clusterInfo, ok := clusterInfos[clusterName]; ok {
		return clusterInfo, nil
	}
	clusterInfo, err := k.bdl.GetCluster(clusterName)
	if err != nil {
		return nil, fmt.Errorf("query cluster info failed, cluster:%s, err:%v", clusterName, err)
	}
	clusterInfos[clusterName] = clusterInfo
	return clusterInfo, nil
}

// getClientSet 从内存中获取 或根据 cluster addr 新建 clientSet
func (k *Kubernetes) getClientSet(clusterName string) (*clientgo.ClientSet, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("empty cluster name")
	}

	if clientSet, ok := clientSets[clusterName]; ok {
		return clientSet, nil
	}

	clusterInfo, err := k.getClusterInfo(clusterName)
	if err != nil {
		return nil, err
	}

	if clusterInfo.SchedConfig == nil || clusterInfo.SchedConfig.MasterURL == "" {
		return nil, fmt.Errorf("empty inet address, cluster:%s", clusterName)
	}

	clientSet, err := clientgo.New(clusterInfo.SchedConfig.MasterURL)
	if err != nil {
		logrus.Errorf("cluster %s clientset create error, parse master url: %s, error: %+v", clusterName, clusterInfo.SchedConfig.MasterURL, err)
		return nil, fmt.Errorf("cluster %s clientset create error", clusterName)
	}

	clientSets[clusterName] = clientSet
	return clientSet, nil
}
