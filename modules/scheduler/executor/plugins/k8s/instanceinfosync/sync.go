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

package instanceinfosync

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/modules/scheduler/instanceinfo"
)

//Synchronization strategy:
// 0. Periodically list all deployment, statefulset, and pod states
// 1. watch deployment, statefulset, pod
// TODO: 2. watch event for more detail messages

type deploymentUtils interface {
	// watch deployment in all namespace, use ctx to cancel
	WatchAllNamespace(ctx context.Context, add, update, delete func(*appsv1.Deployment)) error
	// list deployments with limit
	// return deployment-list, continue, error
	// if returned continue = nil, means that this is the last part of the list
	LimitedListAllNamespace(limit int, cont *string) (*appsv1.DeploymentList, *string, error)
}
type podUtils interface {
	// watch pod in all namespace, use ctx to cancel
	WatchAllNamespace(ctx context.Context, add, update, delete func(*corev1.Pod)) error
	ListAllNamespace(fieldSelectors []string) (*corev1.PodList, error)
	Get(namespace, name string) (*corev1.Pod, error)
}
type statefulSetUtils interface {
	// watch sts in all namespace, use ctx to cancel
	WatchAllNamespace(ctx context.Context, add, update, delete func(*appsv1.StatefulSet)) error
	// list sts with limit
	// return sts-list, continue, error
	// if returned continue = nil, means that this is the last part of the list
	LimitedListAllNamespace(limit int, cont *string) (*appsv1.StatefulSetList, *string, error)
}

type eventUtils interface {
	// watch pod events in all namespaces, use ctx to cancel
	WatchPodEventsAllNamespaces(ctx context.Context, callback func(*corev1.Event)) error
}

type Syncer struct {
	clustername string
	addr        string
	deploy      deploymentUtils
	pod         podUtils
	sts         statefulSetUtils
	event       eventUtils
	dbupdater   *instanceinfo.Client
	bdl         *bundle.Bundle
}

func NewSyncer(clustername, addr string, db *instanceinfo.Client, bdl *bundle.Bundle,
	podutils podUtils, stsutils statefulSetUtils, deployutils deploymentUtils, eventutils eventUtils) *Syncer {
	return &Syncer{clustername, addr, deployutils, podutils, stsutils, eventutils, db, bdl}
}

func (s *Syncer) Sync(ctx context.Context) {
	s.listSync(ctx)
	s.watchSync(ctx)
	s.gc(ctx)
	<-ctx.Done()
}

func (s *Syncer) listSync(ctx context.Context) {
	go s.listSyncPod(ctx)
}

func (s *Syncer) watchSync(ctx context.Context) {
	go s.watchSyncPod(ctx)
	go s.watchSyncEvent(ctx)
}

func (s *Syncer) gc(ctx context.Context) {
	go s.gcDeadInstances(ctx)
	go s.gcServices(ctx)
}

func (s *Syncer) listSyncDeployment(ctx context.Context) {
	var cont *string
	var deploylist *appsv1.DeploymentList
	var err error
	for {
		wait := waitSeconds(cont)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		deploylist, cont, err = s.deploy.LimitedListAllNamespace(100, cont)
		if err != nil {
			logrus.Errorf("failed to list deployments: %v", err)
			cont = nil
			continue
		}
		if err := updateStatelessServiceDeployment(s.dbupdater, deploylist, false); err != nil {
			logrus.Errorf("failed to update statless-service serviceinfo: %v", err)
			continue
		}
		if err := updateAddonDeployment(s.dbupdater, deploylist, false); err != nil {
			logrus.Errorf("failed to update addon serviceinfo: %v", err)
			continue
		}
	}
}

func (s *Syncer) listSyncStatefulSet(ctx context.Context) {
	var cont *string
	var stslist *appsv1.StatefulSetList
	var err error
	for {
		wait := waitSeconds(cont)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		waitSeconds(cont)
		stslist, cont, err = s.sts.LimitedListAllNamespace(100, cont)
		if err != nil {
			logrus.Errorf("failed to list statefulset: %v", err)
			cont = nil
			continue
		}
		if err := updateAddonStatefulSet(s.dbupdater, stslist, false); err != nil {
			logrus.Errorf("failed to update addon serviceinfo: %v", err)
			continue
		}
	}
}

func (s *Syncer) listSyncPod(ctx context.Context) {
	var podlist *corev1.PodList
	var err error
	var initUpdateTime time.Time
	for {
		wait := waitSeconds(nil)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		initUpdateTime = time.Now()
		logrus.Infof("start listpods for: %s", s.addr)

		fieldSelectors := []string{"metadata.namespace!=kube-system"}

		podlist, err = s.pod.ListAllNamespace(fieldSelectors)
		if err != nil {
			logrus.Errorf("failed to list pod: %v", err)
			continue
		}
		logrus.Infof("listpods(%d) for: %s", len(podlist.Items), s.addr)
		if err := updatePodAndInstance(s.dbupdater, podlist, false, nil); err != nil {
			logrus.Errorf("failed to update instanceinfo: %v", err)
			continue
		}
		logrus.Infof("export podlist info start: %s", s.addr)
		exportPodErrInfo(s.bdl, podlist)
		logrus.Infof("export podlist info end: %s", s.addr)
		logrus.Infof("updatepods for: %s", s.addr)
		// it is last part of pod list, so execute gcAliveInstancesInDB
		// GcAliveInstancesInDB is triggered after every 2 complete traversals
		cost := int(time.Now().Sub(initUpdateTime).Seconds())
		if err := gcAliveInstancesInDB(s.dbupdater, cost, s.clustername); err != nil {
			logrus.Errorf("failed to gcAliveInstancesInDB: %v", err)
		}
		cost2 := int(time.Now().Sub(initUpdateTime).Seconds())
		if err := gcPodsInDB(s.dbupdater, cost2, s.clustername); err != nil {
			logrus.Errorf("failed to gcPodsInDB: %v", err)
		}
		logrus.Infof("gcAliveInstancesInDB for: %s", s.addr)
	}
}

func (s *Syncer) watchSyncDeployment(ctx context.Context) {
	addOrUpdate, del := updateDeploymentOnWatch(s.dbupdater)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(10) * time.Second):
		}
		if err := s.deploy.WatchAllNamespace(ctx, addOrUpdate, addOrUpdate, del); err != nil {
			logrus.Errorf("failed to watch update deployment: %v", err)
		}
	}
}

func (s *Syncer) watchSyncStatefulset(ctx context.Context) {
	addOrUpdate, del := updateAddonStatefulSetOnWatch(s.dbupdater)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(10) * time.Second):
		}
		if err := s.sts.WatchAllNamespace(ctx, addOrUpdate, addOrUpdate, del); err != nil {
			logrus.Errorf("failed to watch update statefulset: %v", err)
		}
	}
}

func (s *Syncer) watchSyncPod(ctx context.Context) {
	addOrUpdate, del := updatePodOnWatch(s.bdl, s.dbupdater, s.addr)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(10) * time.Second):
		}
		if err := s.pod.WatchAllNamespace(ctx, addOrUpdate, addOrUpdate, del); err != nil {
			logrus.Errorf("failed to watch update pod: %v, addr: %s", err, s.addr)
		}
	}
}

func (s *Syncer) watchSyncEvent(ctx context.Context) {
	callback := func(e *corev1.Event) {
		if e.Type == "Normal" {
			return
		}
		ns := e.InvolvedObject.Namespace
		name := e.InvolvedObject.Name
		pod, err := s.pod.Get(ns, name)
		if err != nil {
			if !util.IsNotFound(err) {
				logrus.Errorf("failed to get pod: %s/%s, err: %v", ns, name, err)
			}
			return
		}
		if err := updatePodAndInstance(s.dbupdater, &corev1.PodList{Items: []corev1.Pod{*pod}}, false,
			map[string]*corev1.Event{pod.Namespace + "/" + pod.Name: e}); err != nil {
			logrus.Errorf("failed to updatepod: %v", err)
			return
		}
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(10) * time.Second):
		}
		if err := s.event.WatchPodEventsAllNamespaces(ctx, callback); err != nil {
			logrus.Errorf("failed to watch event: %v, addr: %s", err, s.addr)
		}
	}
}

func (s *Syncer) gcDeadInstances(ctx context.Context) {
	if err := gcDeadInstancesInDB(s.dbupdater); err != nil {
		logrus.Errorf("failed to gcInstancesInDB: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(24) * time.Hour):
		}
		if err := gcDeadInstancesInDB(s.dbupdater); err != nil {
			logrus.Errorf("failed to gcInstancesInDB: %v", err)
		}
	}
}

func (s *Syncer) gcServices(ctx context.Context) {
	if err := gcServicesInDB(s.dbupdater); err != nil {
		logrus.Errorf("failed to gcServicesInDB: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(24) * time.Hour):
		}
		if err := gcServicesInDB(s.dbupdater); err != nil {
			logrus.Errorf("failed to gcServicesInDB: %v", err)
		}
	}
}

func waitSeconds(cont *string) time.Duration {
	randsec := rand.Intn(5)
	wait := time.Duration(180+randsec) * time.Second
	if cont == nil {
		wait = time.Duration(60+randsec) * time.Second
	}
	return wait
}
