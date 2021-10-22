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
	"github.com/rancher/apiserver/pkg/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	apierrors2 "github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/k8sclient/config"
	"github.com/erda-project/erda/pkg/strutil"
)

type group struct {
	ready  bool
	server *Server
	cancel context.CancelFunc
}

type Aggregator struct {
	Ctx     context.Context
	Bdl     *bundle.Bundle
	servers sync.Map
}

// NewAggregator new an aggregator with steve servers for all current clusters
func NewAggregator(ctx context.Context, bdl *bundle.Bundle) *Aggregator {
	a := &Aggregator{
		Ctx: ctx,
		Bdl: bdl,
	}
	a.init()
	go a.watchClusters(ctx)
	return a
}

func (a *Aggregator) GetAllClusters() []string {
	var clustersNames []string
	a.servers.Range(func(key, _ interface{}) bool {
		if clusterName, ok := key.(string); ok {
			clustersNames = append(clustersNames, clusterName)
		}
		return true
	})
	return clustersNames
}

func (a *Aggregator) watchClusters(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Hour):
			clusters, err := a.listClusterByType("k8s", "edas")
			if err != nil {
				logrus.Errorf("failed to list k8s clusters when watch: %v", err)
				continue
			}
			exists := make(map[string]struct{})
			for _, cluster := range clusters {
				if cluster.ManageConfig == nil {
					logrus.Infof("manage config for cluster %s is nil, skip it", cluster.Name)
					continue
				}
				exists[cluster.Name] = struct{}{}
				if _, ok := a.servers.Load(cluster.Name); ok {
					continue
				}
				a.Add(cluster)
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

func (a *Aggregator) listClusterByType(types ...string) ([]apistructs.ClusterInfo, error) {
	var result []apistructs.ClusterInfo
	for _, typ := range types {
		clusters, err := a.Bdl.ListClustersWithType(typ)
		if err != nil {
			return nil, err
		}
		result = append(result, clusters...)
	}
	return result, nil
}

func (a *Aggregator) init() {
	clusters, err := a.listClusterByType("k8s", "edas")
	if err != nil {
		logrus.Errorf("failed to list clusters, %v", err)
		return
	}

	for i := range clusters {
		if clusters[i].ManageConfig == nil {
			logrus.Infof("manage config for cluster %s is nil, skip it", clusters[i].Name)
			continue
		}
		a.Add(clusters[i])
	}
}

// Add starts a steve server for k8s cluster with clusterName and add it into aggregator
func (a *Aggregator) Add(clusterInfo apistructs.ClusterInfo) {
	if clusterInfo.Type != "k8s" && clusterInfo.Type != "edas" {
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
		if err = saClient.Delete(a.Ctx, sa.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if _, err = saClient.Create(a.Ctx, sa, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	crClient := client.ClientSet.RbacV1().ClusterRoles()
	for _, cr := range predefinedClusterRole {
		if err = crClient.Delete(a.Ctx, cr.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if _, err = crClient.Create(a.Ctx, cr, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	crbClient := client.ClientSet.RbacV1().ClusterRoleBindings()
	for _, crb := range predefinedClusterRoleBinding {
		if err = crbClient.Delete(a.Ctx, crb.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if _, err = crbClient.Create(a.Ctx, crb, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
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
	newNs, err := nsClient.Create(a.Ctx, system, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	for newNs.Status.Phase != v1.NamespaceActive {
		newNs, err = nsClient.Get(a.Ctx, "erda-system", metav1.GetOptions{})
		if err != nil {
			return err
		}
		select {
		case <-a.Ctx.Done():
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
		cluster, err := a.Bdl.GetCluster(clusterName)
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

		if cluster.Type != "k8s" && cluster.Type != "edas" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(apistructs.NewSteveError(apistructs.BadRequest,
				fmt.Sprintf("cluster %s is not a k8s or edas cluster", clusterName)).JSON())
			return
		}

		logrus.Infof("steve for cluster %s not exist, starting a new server", cluster.Name)
		a.Add(*cluster)
		if s, ok = a.servers.Load(cluster.Name); !ok {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(apistructs.NewSteveError(apistructs.ServerError, "Internal server error").JSON())
			return
		}
	}

	group, _ := s.(*group)
	if !group.ready {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(apistructs.NewSteveError(apistructs.ServerError,
			fmt.Sprintf("API for cluster %s is not ready, please wait", clusterName)).JSON())
		return
	}
	group.server.ServeHTTP(rw, req)
}

func (a *Aggregator) Serve(clusterName string, apiOp *types.APIRequest) error {
	s, ok := a.servers.Load(clusterName)
	if !ok {
		cluster, err := a.Bdl.GetCluster(clusterName)
		if err != nil {
			apiErr, _ := err.(*errorresp.APIError)
			if apiErr.HttpCode() != http.StatusNotFound {
				return apierrors2.ErrInvoke.InternalError(errors.Errorf("failed to get cluster %s, %s", clusterName, apiErr.Error()))
			}
			return apierrors2.ErrInvoke.InvalidParameter(errors.Errorf("cluster %s not found", clusterName))
		}

		if cluster.Type != "k8s" && cluster.Type != "edas" {
			return apierrors2.ErrInvoke.InvalidParameter(errors.Errorf("cluster %s is not a k8s or edas cluster", clusterName))
		}

		logrus.Infof("steve for cluster %s not exist, starting a new server", cluster.Name)
		a.Add(*cluster)
		if s, ok = a.servers.Load(cluster.Name); !ok {
			return apierrors2.ErrInvoke.InternalError(errors.Errorf("failed to start steve server for cluster %s", cluster.Name))
		}
	}

	group, _ := s.(*group)
	if !group.ready {
		return apierrors2.ErrInvoke.InternalError(errors.Errorf("k8s API for cluster %s is not ready, please wait", clusterName))
	}
	return group.server.Handle(apiOp)
}

func (a *Aggregator) createSteve(clusterInfo apistructs.ClusterInfo) (*Server, context.CancelFunc, error) {
	if clusterInfo.ManageConfig == nil {
		return nil, nil, fmt.Errorf("manageConfig of cluster %s is null", clusterInfo.Name)
	}

	restConfig, err := config.ParseManageConfig(clusterInfo.Name, clusterInfo.ManageConfig)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(a.Ctx)
	prefix := GetURLPrefix(clusterInfo.Name)
	server, err := New(ctx, restConfig, &Options{
		AuthMiddleware: emptyMiddleware,
		Router:         RoutesWrapper(prefix),
		URLPrefix:      prefix,
		ClusterName:    clusterInfo.Name,
	})
	if err != nil {
		cancel()
		return nil, nil, err
	}
	go a.preloadCache(server, "node")
	go a.preloadCache(server, "pod")
	return server, cancel, nil
}

func (a *Aggregator) preloadCache(server *Server, resType string) {
	for {
		logrus.Infof("preload cache for %s in cluster %s", resType, server.ClusterName)
		code := a.list(server, resType)
		if code == 200 {
			logrus.Infof("preload cache for %s in cluster %s succeeded", resType, server.ClusterName)
			return
		}
		logrus.Infof("preload cache for %s in cluster %s failed, retry after 5 seconds", resType, server.ClusterName)
		time.Sleep(time.Second * 5)
	}
}

func (a *Aggregator) list(server *Server, resType string) int {
	user := &apiuser.DefaultInfo{
		Name: "admin",
		UID:  "admin",
		Groups: []string{
			"system:masters",
			"system:authenticated",
		},
	}
	withUser := request.WithUser(a.Ctx, user)
	path := strutil.JoinPath("/api/k8s/clusters", server.ClusterName, "v1", resType)
	req, err := http.NewRequestWithContext(withUser, http.MethodGet, path, nil)

	resp := &Response{}
	apiOp := &types.APIRequest{
		Type:           resType,
		Method:         http.MethodGet,
		ResponseWriter: resp,
		Request:        req,
		Response:       &StatusCodeGetter{Response: resp},
	}

	if err != nil {
		logrus.Errorf("failed to new http request when preload cache for %s, %v", resType, err)
		return 500
	}

	apiOp.Request = req
	if err = server.Handle(apiOp); err != nil {
		logrus.Errorf("failed to preload cache for %s, %v", resType, err)
	}
	return resp.StatusCode
}

func emptyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(resp, req)
	})
}
