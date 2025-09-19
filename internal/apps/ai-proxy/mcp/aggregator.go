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

package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/k8sclient/config"
)

type ClusterController struct {
	info   *clusterpb.ClusterInfo
	cancel context.CancelFunc
}

func (c *ClusterController) NeedUpdate(newInfo *clusterpb.ClusterInfo) bool {
	return !proto.Equal(c.info, newInfo)
}

type Aggregator struct {
	ClusterSvc clusterpb.ClusterServiceServer
	clusters   map[string]*ClusterController
	handles    chan *ClusterController
	lock       sync.Mutex

	register *Register
}

func NewAggregator(ctx context.Context, svc clusterpb.ClusterServiceServer, handler *handler_mcp_server.MCPHandler) *Aggregator {
	a := &Aggregator{
		ClusterSvc: svc,
		clusters:   make(map[string]*ClusterController),
		handles:    make(chan *ClusterController, 10),
		lock:       sync.Mutex{},
		register:   NewRegister(handler),
	}
	go a.syncRestConfig(ctx)
	return a
}

func (a *Aggregator) syncRestConfig(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)

	for {
		c := apis.WithInternalClientContext(ctx, discover.SvcAIProxy)

		cluster, err := a.ClusterSvc.ListCluster(c, &clusterpb.ListClusterRequest{
			ClusterType: "k8s",
			OrgID:       0,
		})
		if err != nil {
			logrus.Error("Failed to get cluster info", err)
			return
		}

		for _, info := range cluster.Data {
			logrus.Infof("get cluster info: %+v", info.Name)
			control := ClusterController{
				info: info,
			}

			a.lock.Lock()
			controller, ok := a.clusters[info.Name]
			if !ok {
				a.handles <- &control
				a.clusters[info.Name] = &control
			} else if controller != nil && controller.NeedUpdate(info) {
				if controller.cancel != nil {
					controller.cancel()
				}
				a.handles <- &control
				a.clusters[info.Name] = &control
			}
			a.lock.Unlock()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case handle := <-a.handles:
			logrus.Infof("load cluster: %+v", handle == nil)
			restConfig, err := config.ParseManageConfigPb(handle.info.Name, handle.info.ManageConfig)
			if err != nil {
				logrus.Error(err)
				continue
			}
			ctx, cancelFunc := context.WithCancel(ctx)
			handle.cancel = cancelFunc
			go func() {
				// if running error, delete cluster info, retry in next sync
				if err = a.run(ctx, restConfig, handle.info.Name); err != nil {
					logrus.Errorf("run cluster failed: %v", err)
					a.lock.Lock()
					delete(a.clusters, handle.info.Name)
					a.lock.Unlock()
				}
			}()
		}
	}
}

func (a *Aggregator) run(ctx context.Context, conf *rest.Config, clusterName string) error {
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return err
	}
	watcher, err := clientset.CoreV1().Services("").Watch(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/component=mcp-server",
	})
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		svc, ok := event.Object.(*v1.Service)
		if !ok {
			logrus.Errorf("event object is not a service: %+v", event.Object)
			continue
		}
		fmt.Printf("[%s] Type: %s, Service: %s, ClusterIP: %s\n",
			svc.Namespace, event.Type, svc.Name, svc.Spec.ClusterIP)

		selector := labels.SelectorFromSet(svc.Spec.Selector).String()

		pods, err := clientset.CoreV1().Pods(svc.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return err
		}

		if len(pods.Items) == 0 {
			return errors.New("no available pods")
		}

		err = a.register.register(ctx, svc, pods.Items[0], clusterName)
		if err != nil {
			logrus.Errorf("register service failed: %v", err)
			continue
		}
	}
	return nil
}
