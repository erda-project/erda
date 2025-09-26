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
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda-infra/base/logs"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_mcp_server"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
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
	logger   logs.Logger
}

func NewAggregator(ctx context.Context, svc clusterpb.ClusterServiceServer, handler *handler_mcp_server.MCPHandler, logger logs.Logger) *Aggregator {
	a := &Aggregator{
		ClusterSvc: svc,
		clusters:   make(map[string]*ClusterController),
		handles:    make(chan *ClusterController, 10),
		lock:       sync.Mutex{},
		register:   NewRegister(handler, logger),
		logger:     logger,
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
			a.logger.Error("Failed to get cluster info", err)
			return
		}

		for _, info := range cluster.Data {
			a.logger.Infof("get cluster info: %+v", info.Name)
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
			if handle == nil {
				a.logger.Warnf("cluster handler is nil")
				continue
			}
			a.logger.Infof("load cluster: %+v", handle.info.Name)
			restConfig, err := config.ParseManageConfigPb(handle.info.Name, handle.info.ManageConfig)
			if err != nil {
				a.logger.Error(err)
				continue
			}
			ctx, cancelFunc := context.WithCancel(ctx)
			handle.cancel = cancelFunc
			go func() {
				// if running error, delete cluster info, retry in next sync
				if err = a.run(ctx, restConfig, handle.info.Name); err != nil {
					a.logger.Errorf("run cluster failed: %v", err)
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
		LabelSelector: fmt.Sprintf("%s=mcp-server", vars.LabelMcpErdaCloudComponent),
	})
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		svc, ok := event.Object.(*corev1.Service)
		if !ok {
			a.logger.Errorf("event object is not a service: %+v", event.Object)
			continue
		}
		a.logger.Infof("[%s] Type: %s, Service: %s, ClusterIP: %s\n",
			svc.Namespace, event.Type, svc.Name, svc.Spec.ClusterIP)

		go a.watchEndpointSlices(ctx, clientset, clusterName, svc)
	}
	return nil
}

func (a *Aggregator) watchEndpointSlices(ctx context.Context, clientset *kubernetes.Clientset, clusterName string, svc *corev1.Service) {
	selector := labels.SelectorFromSet(svc.Spec.Selector).String()

	a.logger.Infof("watch EndpointSlices: %s with service %v", selector, svc.Name)
	watcher, err := clientset.DiscoveryV1().EndpointSlices(svc.Namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		a.logger.Errorf("list endpoint slice failed: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watcher.ResultChan():
			eps, ok := event.Object.(*discoveryv1.EndpointSlice)
			if !ok {
				a.logger.Errorf("event object is not a EndpointSlice: %+v in cluster: %s", event.Object, clusterName)
				return
			}

			// All endpoints must be in the ready state before registration proceeds.
			var ready = true
			for _, endpoint := range eps.Endpoints {
				if endpoint.Conditions.Ready != nil {
					ready = ready && *endpoint.Conditions.Ready
				}
			}

			if !ready {
				a.logger.Infof("not all endpoints are ready yet")
				continue
			}

			a.logger.Infof("%s all endpoints ready", eps.Name)

			err = a.register.register(ctx, svc, clusterName)
			if err != nil {
				a.logger.Errorf("register service failed: %v", err)
				continue
			}
			a.logger.Infof("register mcp server success, service: %s, namespace: %s", svc.Name, svc.Namespace)
		}
	}
}
