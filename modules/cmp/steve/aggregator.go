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

package steve

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/k8sclient/config"
)

type group struct {
	ready  bool
	server *Server
	cancel context.CancelFunc
}

type Aggregator struct {
	ctx     context.Context
	bdl     *bundle.Bundle
	servers sync.Map
}

// NewAggregator new an aggregator with steve servers for all current clusters
func NewAggregator(ctx context.Context, bdl *bundle.Bundle) *Aggregator {
	a := &Aggregator{
		ctx: ctx,
		bdl: bdl,
	}
	a.init(bdl)
	go a.watchClusters(ctx)
	return a
}

func (a *Aggregator) watchClusters(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Hour):
			clusters, err := a.bdl.ListClusters("k8s")
			if err != nil {
				logrus.Errorf("failed to list clusters when watch: %v", err)
				continue
			}
			exists := make(map[string]struct{})
			for _, cluster := range clusters {
				exists[cluster.Name] = struct{}{}
				if _, ok := a.servers.Load(cluster.Name); ok {
					continue
				}
				a.Add(&cluster)
			}

			checkDeleted := func(key interface{}, value interface{}) (res bool) {
				res = true
				if _, ok := exists[key.(string)]; ok {
					return
				}
				if err = a.Delete(key.(string)); err != nil {
					logrus.Errorf("failed to stop steve server for cluster %s when watch, %v", key.(string), err)
					return
				}
				return
			}
			a.servers.Range(checkDeleted)
		}
	}
}

func (a *Aggregator) init(bdl *bundle.Bundle) {
	clusters, err := bdl.ListClusters("k8s")
	if err != nil {
		logrus.Errorf("failed to list clusters, %v", err)
		return
	}

	for i := range clusters {
		if clusters[i].ManageConfig == nil {
			continue
		}
		a.Add(&clusters[i])
	}
}

// Add starts a steve server for k8s cluster with clusterName and add it into aggregator
func (a *Aggregator) Add(clusterInfo *apistructs.ClusterInfo) {
	if clusterInfo.Type != "k8s" {
		return
	}

	if _, ok := a.servers.Load(clusterInfo.Name); ok {
		return
	}

	g := &group{ready: false}
	a.servers.Store(clusterInfo.Name, g)
	go func() {
		logrus.Infof("starting steve server for cluster %s", clusterInfo.Name)
		server, cancel, err := a.createSteve(clusterInfo)
		if err != nil {
			logrus.Errorf("failed to create steve server for cluster %s, %v", clusterInfo.Name, err)
			a.servers.Delete(clusterInfo.Name)
			return
		}

		g := &group{
			ready:  true,
			server: server,
			cancel: cancel,
		}
		a.servers.Store(clusterInfo.Name, g)
		logrus.Infof("steve server for cluster %s started", clusterInfo.Name)

		if err = a.createPredefinedResource(clusterInfo.Name); err != nil {
			logrus.Errorf("failed to create predefined resource for cluster %s, %v", clusterInfo.Name, err)
			a.servers.Delete(clusterInfo.Name)
		}
	}()
}

func (a *Aggregator) createPredefinedResource(clusterName string) error {
	client, err := k8sclient.New(clusterName)
	if err != nil {
		return err
	}

	if err := a.insureSystemNamespace(client); err != nil {
		return err
	}

	for _, sa := range predefinedServiceAccount {
		saClient := client.ClientSet.CoreV1().ServiceAccounts(sa.Namespace)
		if _, err = saClient.Create(a.ctx, sa, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	crClient := client.ClientSet.RbacV1().ClusterRoles()
	for _, cr := range predefinedClusterRole {
		if _, err = crClient.Create(a.ctx, cr, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	crbClient := client.ClientSet.RbacV1().ClusterRoleBindings()
	for _, crb := range predefinedClusterRoleBinding {
		if _, err = crbClient.Create(a.ctx, crb, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (a *Aggregator) insureSystemNamespace(client *k8sclient.K8sClient) error {
	nsClient := client.ClientSet.CoreV1().Namespaces()
	system := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "erda-system",
		},
	}
	newNs, err := nsClient.Create(a.ctx, system, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	for newNs.Status.Phase != v1.NamespaceActive {
		newNs, err = nsClient.Get(a.ctx, "erda-system", metav1.GetOptions{})
		if err != nil {
			return err
		}
		select {
		case <-a.ctx.Done():
			return errors.New("failed to watch system namespace, context canceled")
		case <-time.After(time.Second):
			logrus.Infof("creating erda system namespace...")
		}
	}
	return nil
}

// Delete closes a steve server for k8s cluster with clusterName and delete it from aggregator
func (a *Aggregator) Delete(clusterName string) error {
	g, ok := a.servers.Load(clusterName)
	if !ok {
		return nil
	}

	group, _ := g.(*group)
	if group.ready {
		group.cancel()
	}
	a.servers.Delete(clusterName)
	return nil
}

// ServeHTTP forwards API request to corresponding steve server
func (a *Aggregator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	clusterName := vars["clusterName"]

	if clusterName == "" {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(apistructs.NewSteveError(apistructs.NotFound, "cluster name is required").JSON())
		return
	}

	s, ok := a.servers.Load(clusterName)
	if !ok {
		cluster, err := a.bdl.GetCluster(clusterName)
		if err != nil {
			apiErr, _ := err.(*errorresp.APIError)
			if apiErr.HttpCode() != http.StatusNotFound {
				logrus.Errorf("failed to get cluster %s, %s", clusterName, apiErr.Error())
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(apistructs.NewSteveError(apistructs.ServerError, "Internal server error").JSON())
				return
			}
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(apistructs.NewSteveError(apistructs.NotFound,
				fmt.Sprintf("cluster %s not found", clusterName)).JSON())
			return
		}

		if cluster.Type != "k8s" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(apistructs.NewSteveError(apistructs.BadRequest,
				fmt.Sprintf("cluster %s is not a k8s cluster", clusterName)).JSON())
			return
		}

		logrus.Infof("steve for cluster %s not exist, starting a new server", cluster.Name)
		a.Add(cluster)
		if s, ok = a.servers.Load(cluster.Name); !ok {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(apistructs.NewSteveError(apistructs.ServerError, "Internal server error").JSON())
		}
	}

	group, _ := s.(*group)
	if !group.ready {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(apistructs.NewSteveError(apistructs.ServerError,
			fmt.Sprintf("k8s API for cluster %s is not ready, please wait", clusterName)).JSON())
		return
	}
	group.server.ServeHTTP(rw, req)
}

func (a *Aggregator) createSteve(clusterInfo *apistructs.ClusterInfo) (*Server, context.CancelFunc, error) {
	if clusterInfo.ManageConfig == nil {
		return nil, nil, fmt.Errorf("manageConfig of cluster %s is null", clusterInfo.Name)
	}

	restConfig, err := config.ParseManageConfig(clusterInfo.Name, clusterInfo.ManageConfig)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(a.ctx)
	prefix := GetURLPrefix(clusterInfo.Name)
	server, err := New(ctx, restConfig, &Options{
		AuthMiddleware: emptyMiddleware,
		Router:         RoutesWrapper(prefix),
		URLPrefix:      prefix,
	})
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return server, cancel, nil
}

func emptyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(resp, req)
	})
}
