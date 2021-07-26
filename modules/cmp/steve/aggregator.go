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

package steve

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
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
				if err = a.Add(&cluster); err != nil {
					logrus.Errorf("failed to add steve server for cluster %s when watch, %v", cluster.Name, err)
					continue
				}
				logrus.Infof("start steve server for cluster %s when watch", cluster.Name)
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

	for _, cluster := range clusters {
		if cluster.ManageConfig == nil {
			continue
		}
		if err = a.Add(&cluster); err != nil {
			logrus.Errorf("failed to start steve for cluster %s when init aggragetor, %v", cluster.Name, err)
		}
	}
}

// Add starts a steve server for k8s cluster with clusterName and add it into aggregator
func (a *Aggregator) Add(clusterInfo *apistructs.ClusterInfo) error {
	if clusterInfo.Type != "k8s" {
		return nil
	}

	if _, ok := a.servers.Load(clusterInfo.Name); ok {
		return nil
	}

	g := &group{ready: false}
	a.servers.Store(clusterInfo.Name, g)
	go func() {
		server, cancel, err := a.createSteve(clusterInfo)
		if err != nil {
			logrus.Errorf("failed to create steve server for cluster %s", clusterInfo.Name)
			a.servers.Delete(clusterInfo.Name)
			return
		}

		g := &group{
			ready:  true,
			server: server,
			cancel: cancel,
		}
		a.servers.Store(clusterInfo.Name, g)
	}()
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
		rw.Write([]byte("cluster name is required"))
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
				rw.Write([]byte("Internal server error"))
				return
			}
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(fmt.Sprintf("cluster %s not found", clusterName)))
			return
		}

		if cluster.Type != "k8s" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("cluster %s is not a k8s cluster", clusterName)))
			return
		}

		logrus.Infof("steve for cluster %s not exist, starting a new server", cluster.Name)
		if err = a.Add(cluster); err != nil {
			logrus.Errorf("failed to start steve server for cluster %s, %v", cluster.Name, err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Internal server error"))
		}
		s, _ = a.servers.Load(cluster.Name)
	}

	group, _ := s.(*group)
	if !group.ready {
		rw.WriteHeader(http.StatusAccepted)
		rw.Write([]byte(fmt.Sprintf("k8s API for cluster %s is not ready, please wait", clusterName)))
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
		Router:    RoutesWrapper(prefix),
		URLPrefix: prefix,
	})
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return server, cancel, nil
}
