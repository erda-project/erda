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

package instanceinfosync

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	vpatypes "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"

	"github.com/erda-project/erda/bundle"
	orgCache "github.com/erda-project/erda/internal/tools/orchestrator/cache/org"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/i18n"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/scaledobject"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/instanceinfo"
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

type hpaUtils interface {
	Get(namespace, name string) (*autoscalingv2beta2.HorizontalPodAutoscaler, error)
}

type eventUtils interface {
	// watch pod events in all namespaces, use ctx to cancel
	WatchPodEventsAllNamespaces(ctx context.Context, callback func(*corev1.Event)) error
	WatchHPAEventsAllNamespaces(ctx context.Context, callback func(*corev1.Event)) error
}

type Syncer struct {
	clustername     string
	addr            string
	deploy          deploymentUtils
	pod             podUtils
	sts             statefulSetUtils
	event           eventUtils
	hpa             hpaUtils
	dbupdater       *instanceinfo.Client
	bdl             *bundle.Bundle
	podScaledObject *scaledobject.ErdaScaledObject
}

func NewSyncer(clustername, addr string, db *instanceinfo.Client, bdl *bundle.Bundle,
	podutils podUtils, stsutils statefulSetUtils, deployutils deploymentUtils, eventutils eventUtils, hpautils hpaUtils, podScaledObject *scaledobject.ErdaScaledObject) *Syncer {
	return &Syncer{clustername, addr, deployutils, podutils, stsutils, eventutils, hpautils, db, bdl, podScaledObject}
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
	go s.watchSyncHPAEvent(ctx)
	go s.watchVPA(ctx)
}

func (s *Syncer) gc(ctx context.Context) {
	go s.gcDeadInstances(ctx, s.clustername)
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
		orgs, err := updatePodAndInstance(s.dbupdater, podlist, false, nil)
		if err != nil {
			logrus.Errorf("failed to update instanceinfo: %v", err)
			continue
		}
		logrus.Infof("export podlist info start: %s", s.addr)
		exportPodErrInfo(s.bdl, podlist, orgs)
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
		if e.Type == "Normal" && e.Reason != pstypes.ErdaVPAPodEvictEventReason {
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

		if e.Reason == pstypes.ErdaVPAPodEvictEventReason {
			logrus.Infof("generate deployment log for VPA evict pod %s/%s", ns, name)
			var locale string
			if org, ok := orgCache.GetOrgByOrgID(pod.Labels["DICE_ORG_ID"]); ok {
				locale = org.Locale
			}
			newMessage := fmt.Sprintf("Pod %s evicted by VPA Updater to apply resource recommendation.", name)
			buildVPAEventInfo(s.bdl, pod,
				fmt.Sprintf("Pod %s evicted by VPA for update. event Type: %s, Reason:%s, Message:%s",
					name, e.Type, e.Reason, newMessage /* e.Message*/),
				i18n.Sprintf(locale, "AutoScaleService", pod.Labels["DICE_SERVICE_NAME"], newMessage /*e.Message*/),
				"podautoscaled")
		}
		if _, err := updatePodAndInstance(s.dbupdater, &corev1.PodList{Items: []corev1.Pod{*pod}}, false,
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

func (s *Syncer) gcDeadInstances(ctx context.Context, clusterName string) {
	if err := gcDeadInstancesInDB(s.dbupdater, clusterName); err != nil {
		logrus.Errorf("failed to gcInstancesInDB: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(1) * time.Hour):
		}
		if err := gcDeadInstancesInDB(s.dbupdater, clusterName); err != nil {
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

func (s *Syncer) watchSyncHPAEvent(ctx context.Context) {
	callback := func(e *corev1.Event) {

		logrus.Infof("Begin process event Type: %v  Reason: %v  InvolvedObject.Kind:  %v, InvolvedObject.Name: %v", e.Type, e.Reason, e.InvolvedObject.Kind, e.InvolvedObject.Name)

		ns := e.InvolvedObject.Namespace
		name := e.InvolvedObject.Name
		hpa, err := s.hpa.Get(ns, name)
		if err != nil {
			if !util.IsNotFound(err) {
				logrus.Errorf("failed to get hpa %s/%s for event %s, err: %v", ns, name, e.Name, err)
			}
			return
		}

		if hpa.Labels == nil {
			logrus.Errorf("hpa %s/%s not create by erda, skip hpa event create", ns, name)
			return
		}

		_, ok := hpa.Labels[pstypes.ErdaPAObjectRuntimeIDLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s runtime info: label %s not set", ns, name, pstypes.ErdaPAObjectRuntimeIDLabel)
			return
		}
		_, ok = hpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s serviceName info: label %s not set", ns, name, pstypes.ErdaPAObjectRuntimeServiceNameLabel)
			return
		}
		_, ok = hpa.Labels[pstypes.ErdaPAObjectRuleIDLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s ruleId info: label %s not set", ns, name, pstypes.ErdaPAObjectRuleIDLabel)
			return
		}
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s ruleId info: label %s not set", ns, name, pstypes.ErdaPAObjectOrgIDLabel)
			return
		}

		var locale string
		if org, ok := orgCache.GetOrgByOrgID(hpa.Labels[pstypes.ErdaPAObjectOrgIDLabel]); ok {
			locale = org.Locale
		}
		buildHPAEventInfo(s.bdl, *hpa,
			fmt.Sprintf("Service %s HorizontalPodAutoscaler event Type: %s, Reason:%s, Message:%s",
				hpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel], e.Type, e.Reason, e.Message),
			i18n.Sprintf(locale, "AutoScaleService", hpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel], e.Message),
			"podautoscaled")

		// TODO: may save hpa events in mysql
		/*
			runtimeId, err := strconv.ParseUint(hpa.Labels[pstypes.ErdaPAObjectRuntimeIDLabel], 10, 64)
			if err != nil {
				logrus.Errorf("failed to MarShall hpa event for %s/%s, parse runtimeId err: %v", ns, name, err)
				return
			}

			eventDetail := dbclient.EventDetail{
				//eventDetail := instanceinfo.EventDetail{
				LastTimestamp: e.LastTimestamp,
				Type:          e.Type,
				Reason:        e.Reason,
				Message:       e.Message,
			}

			eventBytes, err := json.Marshal(eventDetail)
			if err != nil {
				logrus.Errorf("failed to MarShall hpa event for %s/%s, err: %v", ns, name, err)
				return
			}

			hpaEvent := dbclient.HPAEventInfo{
				ID:          hpa.Labels[pstypes.ErdaHPAObjectRuleIDLabel],
				RuntimeID:   runtimeId,
				OrgID:       org.ID,
				OrgName:     org.Name,
				ServiceName: hpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel],
				Event:       string(eventBytes),
			}

			err = s.dbupdater.CreateHPAEventInfo(hpaEvent)
			if err != nil {
				logrus.Errorf("failed to create hpa event: %v", err)
				return
			}
		*/
		logrus.Infof("Send erda hpa event successfully. Type: %v  Name:%s  Reason: %v  InvolvedObject.Kind:  %v, InvolvedObject.Name: %v   ResourceVersion: %v ", e.Type, e.Name, e.Reason, e.InvolvedObject.Kind, e.InvolvedObject.Name, e.ResourceVersion)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
		if err := s.event.WatchHPAEventsAllNamespaces(ctx, callback); err != nil {
			logrus.Errorf("failed to watch event: %v, addr: %s", err, s.addr)
		}
	}
}

func (s *Syncer) watchVPA(ctx context.Context) {
	callback := func(vpa *vpatypes.VerticalPodAutoscaler) {
		logrus.Infof("Begin process VerticalPodAutoscaler object: %s/%s", vpa.Namespace, vpa.Name)

		// delete old recommendations data which not in vpa recommender windows
		s.dbupdater.DeletedOldVPARecommendations(pstypes.ErdaVPARecommendationsHistory)

		if vpa.Labels == nil {
			logrus.Errorf("vpa %s/%s not create by erda, skip hpa event create", vpa.Namespace, vpa.Name)
			return
		}

		_, ok := vpa.Labels[pstypes.ErdaPAObjectRuntimeIDLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s runtime info: label %s not set", vpa.Namespace, vpa.Name, pstypes.ErdaPAObjectRuntimeIDLabel)
			return
		}
		_, ok = vpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s serviceName info: label %s not set", vpa.Namespace, vpa.Name, pstypes.ErdaPAObjectRuntimeServiceNameLabel)
			return
		}
		_, ok = vpa.Labels[pstypes.ErdaPAObjectRuleIDLabel]
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s ruleId info: label %s not set", vpa.Namespace, vpa.Name, pstypes.ErdaPAObjectRuleIDLabel)
			return
		}
		if !ok {
			logrus.Errorf("failed to get hpa %s/%s ruleId info: label %s not set", vpa.Namespace, vpa.Name, pstypes.ErdaPAObjectOrgIDLabel)
			return
		}

		org, ok := orgCache.GetOrgByOrgID(vpa.Labels[pstypes.ErdaPAObjectOrgIDLabel])
		if !ok {
			// org = xxxx
		}

		runtimeId, err := strconv.ParseUint(vpa.Labels[pstypes.ErdaPAObjectRuntimeIDLabel], 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse runtimeId from vpa %s/%s, parse runtimeId err: %v", vpa.Namespace, vpa.Name, err)
			return
		}

		if vpa.Status.Recommendation != nil && len(vpa.Status.Recommendation.ContainerRecommendations) > 0 {
			for _, cr := range vpa.Status.Recommendation.ContainerRecommendations {
				needKeep := false
				latestVcr, err := s.dbupdater.GetLatestVPARecommendation(runtimeId, vpa.Labels[pstypes.ErdaPAObjectRuleIDLabel], vpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel], cr.ContainerName)
				if err != nil {
					logrus.Errorf("failed to get latet vpa container recommendation for vpa %s/%s, error: %v", vpa.Namespace, vpa.Name, err)
					needKeep = true
				} else {
					logrus.Infof("get latet vpa container recommendation for vpa %s/%s, ID: %v", vpa.Namespace, vpa.Name, latestVcr.ID)
					if latestVcr.LowerCPURequest != float64(int(cr.LowerBound.Cpu().AsApproximateFloat64()*100))/100 ||
						latestVcr.LowerMemoryRequest != float64(cr.LowerBound.Memory().Value()/1024/1024) ||
						latestVcr.UpperCPURequest != float64(int(cr.UpperBound.Cpu().AsApproximateFloat64()*100))/100 ||
						latestVcr.UpperMemoryRequest != float64(cr.UpperBound.Memory().Value()/1024/1024) ||
						latestVcr.TargetCPURequest != float64(int(cr.Target.Cpu().AsApproximateFloat64()*100))/100 ||
						latestVcr.TargetMemoryRequest != float64(cr.Target.Memory().Value()/1024/1024) ||
						latestVcr.UncappedCPURequest != float64(int(cr.UncappedTarget.Cpu().AsApproximateFloat64()*100))/100 ||
						latestVcr.UncappedMemoryRequest != float64(cr.UncappedTarget.Memory().Value()/1024/1024) {
						needKeep = true
					}
				}

				if needKeep {
					vpaRecommendation := dbclient.RuntimeVPAContainerRecommendation{
						ID:                    uuid.NewString(),
						RuleName:              "",
						RuleID:                vpa.Labels[pstypes.ErdaPAObjectRuleIDLabel],
						RuleNameSpace:         "",
						OrgID:                 org.ID,
						OrgName:               org.Name,
						ProjectID:             0,
						ProjectName:           "",
						ApplicationID:         0,
						ApplicationName:       "",
						RuntimeID:             runtimeId,
						RuntimeName:           "",
						Workspace:             "",
						ClusterName:           "",
						ServiceName:           vpa.Labels[pstypes.ErdaPAObjectRuntimeServiceNameLabel],
						ContainerName:         cr.ContainerName,
						LowerCPURequest:       float64(int(cr.LowerBound.Cpu().AsApproximateFloat64()*100)) / 100,
						LowerMemoryRequest:    float64(cr.LowerBound.Memory().Value() / 1024 / 1024),
						UpperCPURequest:       float64(int(cr.UpperBound.Cpu().AsApproximateFloat64()*100)) / 100,
						UpperMemoryRequest:    float64(cr.UpperBound.Memory().Value() / 1024 / 1024),
						TargetCPURequest:      float64(int(cr.Target.Cpu().AsApproximateFloat64()*100)) / 100,
						TargetMemoryRequest:   float64(cr.Target.Memory().Value() / 1024 / 1024),
						UncappedCPURequest:    float64(int(cr.UncappedTarget.Cpu().AsApproximateFloat64()*100)) / 100,
						UncappedMemoryRequest: float64(cr.UncappedTarget.Memory().Value() / 1024 / 1024),
						SoftDeletedAt:         0,
					}

					err = s.dbupdater.CreateVPARecommendation(vpaRecommendation)
					if err != nil {
						logrus.Errorf("failed to create vpa container recommendation info: %v", err)
						return
					}
				}
			}
			logrus.Infof("Send vpa recommendation successfully for vpa %s/%s.", vpa.Namespace, vpa.Name)
		}

	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
		if err := s.podScaledObject.WatchVPAsAllNamespaces(ctx, callback); err != nil {
			if !strings.Contains(err.Error(), "could not find the requested resource") {
				logrus.Errorf("failed to watch vpa: %v, addr: %s", err, s.addr)
			}
		}
	}
}
