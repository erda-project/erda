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

type Aggregator struct {
	ctx     context.Context
	bdl     *bundle.Bundle
	servers sync.Map
	cancels sync.Map
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

	server, cancel, err := a.createSteve(clusterInfo)
	if err != nil {
		return err
	}

	a.servers.Store(clusterInfo.Name, server)
	a.cancels.Store(clusterInfo.Name, cancel)
	return nil
}

// Delete closes a steve server for k8s cluster with clusterName and delete it from aggregator
func (a *Aggregator) Delete(clusterName string) error {
	if _, ok := a.servers.Load(clusterName); !ok {
		return nil
	}
	a.servers.Delete(clusterName)

	c, ok := a.cancels.Load(clusterName)
	if !ok {
		return nil
	}
	cancel, _ := c.(context.CancelFunc)
	cancel()
	a.cancels.Delete(clusterName)

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
		server, cancel, err := a.createSteve(cluster)
		if err != nil {
			logrus.Errorf("failed to create steve for cluster %s", cluster.Name)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Internal server error"))
			return
		}
		a.servers.Store(cluster.Name, server)
		a.cancels.Store(cluster.Name, cancel)

		server.ServeHTTP(rw, req)
		return
	}

	server := s.(*Server)
	server.ServeHTTP(rw, req)
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
