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

package rocketmq

import (
	"bytes"
	"encoding/json"
	"fmt"

	rocketmqv1alpha1 "erda.cloud/rocketmq/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	svcNameSrv = "rocketmq-namesrv"
	svcBroker  = "rocketmq-broker"
	svcConsole = "rocketmq-console"
)

var (
	scStorageMode = "StorageClass"
)

type RocketMQOperator struct {
	k8s         addon.K8SUtil
	ns          addon.NamespaceUtil
	client      *httpclient.HTTPClient
	overcommit  addon.OverCommitUtil
	statefulset addon.StatefulsetUtil
}

func New(k8s addon.K8SUtil, ns addon.NamespaceUtil, client *httpclient.HTTPClient, overcommit addon.OverCommitUtil, sts addon.StatefulsetUtil) *RocketMQOperator {
	return &RocketMQOperator{
		k8s:         k8s,
		ns:          ns,
		client:      client,
		overcommit:  overcommit,
		statefulset: sts,
	}
}

func (r *RocketMQOperator) IsSupported() bool {
	resp, err := r.client.Get(r.k8s.GetK8SAddr()).
		Path("/apis/addons.erda.cloud/v1alpha1").
		Do().
		DiscardBody()
	if err != nil {
		logrus.Errorf("failed to query /apis/addons.erda.cloud/v1alpha1, host: %v, err: %v",
			r.k8s.GetK8SAddr(), err)
		return false
	}
	if !resp.IsOK() {
		return false
	}
	return true
}

func (r *RocketMQOperator) Validate(sg *apistructs.ServiceGroup) error {
	operator, ok := sg.Labels["USE_OPERATOR"]
	if !ok {
		return fmt.Errorf("[BUG] sg need USE_OPERATOR label")
	}
	if strutil.ToLower(operator) != "rocketmq" {
		return fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'rocketmq'")
	}
	if len(sg.Services) != 3 {
		return fmt.Errorf("illegal services num: %d", len(sg.Services))
	}
	if _, ok := sg.Labels["VERSION"]; !ok {
		return fmt.Errorf("[BUG] sg need VERSION label")
	}
	return nil
}

// Convert sg to cr, which is kubernetes yaml
func (r *RocketMQOperator) Convert(sg *apistructs.ServiceGroup) (any, error) {
	var nameSrvSpec rocketmqv1alpha1.NameServiceSpec
	var brokerSpec rocketmqv1alpha1.BrokerSpec
	var consoleSpec rocketmqv1alpha1.ConsoleSpec

	scheinfo := sg.ScheduleInfo2
	scheinfo.Stateful = true
	affinity := constraintbuilders.K8S(&scheinfo, nil, nil, nil).Affinity.NodeAffinity

	for i := range sg.Services {
		svc := sg.Services[i]
		workspace, _ := util.GetDiceWorkspaceFromEnvs(svc.Env)
		containerResources, err := r.overcommit.ResourceOverCommit(workspace, svc.Resources)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate service %s container resources, err: %v", svc.Name, err)
		}

		switch svc.Name {
		case svcNameSrv:
			nameSrvSpec = r.convertNameSrv(svc, affinity, containerResources)
		case svcBroker:
			brokerSpec = r.convertBroker(svc, affinity, containerResources)
		case svcConsole:
			consoleSpec = r.convertConsole(svc, affinity, containerResources)
		}
	}
	rocketMQ := rocketmqv1alpha1.RocketMQ{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RocketMQ",
			APIVersion: "addons.erda.cloud/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sg.ID,
			Namespace: genK8SNamespace(sg.Type, sg.ID),
		},
		Spec: rocketmqv1alpha1.RocketMQSpec{
			NameServiceSpec: nameSrvSpec,
			BrokerSpec:      brokerSpec,
			ConsoleSpec:     consoleSpec,
		},
	}
	return rocketMQ, nil

}

func (r *RocketMQOperator) Create(k8syml interface{}) error {
	rocketMQ, ok := k8syml.(rocketmqv1alpha1.RocketMQ)
	if !ok {
		return fmt.Errorf("invalid type, the k8syml should be RocketMQ")
	}
	if err := r.ns.Exists(rocketMQ.Namespace); err != nil {
		if err := r.ns.Create(rocketMQ.Namespace, nil); err != nil {
			return err
		}
	}
	var b bytes.Buffer
	resp, err := r.client.Post(r.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/addons.erda.cloud/v1alpha1/namespaces/%s/rocketmqs", rocketMQ.Namespace)).
		JSONBody(rocketMQ).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to create rocketmq, %s/%s, err: %v, body: %v", rocketMQ.Namespace, rocketMQ.Name, err, b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to create rocketmq, %s/%s, statuscode: %v, body: %v",
			rocketMQ.Namespace, rocketMQ.Name, resp.StatusCode(), b.String())
	}

	return nil
}

func (r *RocketMQOperator) Inspect(sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	rocketMQ, err := r.getRocketMQ(genK8SNamespace(sg.Type, sg.ID), sg.ID)
	if err != nil {
		return nil, err
	}
	var rocketMQService *apistructs.Service
	var consoleService *apistructs.Service
	for i := range sg.Services {
		if sg.Services[i].Name == svcNameSrv {
			rocketMQService = &sg.Services[i]
		}
		if sg.Services[i].Name == svcConsole {
			consoleService = &sg.Services[i]
		}
	}
	rocketMQService.Vip = strutil.Join([]string{rocketMQ.Spec.NameServiceSpec.Name, rocketMQ.Namespace, "svc.cluster.local"}, ".")
	consoleService.Vip = strutil.Join([]string{rocketMQ.Spec.ConsoleSpec.Name, rocketMQ.Namespace, "svc.cluster.local"}, ".")
	switch rocketMQ.Status.ConditionStatus {
	case rocketmqv1alpha1.ConditionReady:
		sg.Status = apistructs.StatusHealthy
	case rocketmqv1alpha1.ConditionFailed:
		sg.Status = apistructs.StatusUnHealthy
	default:
	}
	return sg, nil
}

func (r *RocketMQOperator) Remove(sg *apistructs.ServiceGroup) error {
	ns := genK8SNamespace(sg.Type, sg.ID)
	if err := r.statefulset.Delete(ns, svcNameSrv); err != nil && err != k8serror.ErrNotFound {
		return err
	}
	if err := r.statefulset.Delete(ns, svcBroker); err != nil && err != k8serror.ErrNotFound {
		return err
	}
	if err := r.statefulset.Delete(ns, svcConsole); err != nil && err != k8serror.ErrNotFound {
		return err
	}
	var b bytes.Buffer
	resp, err := r.client.Delete(r.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/addons.erda.cloud/v1alpha1/namespaces/%s/rocketmqs/%s", ns, sg.ID)).
		JSONBody(k8sapi.DeleteOptions).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to delete rocketmq, %s/%s, err: %v", ns, sg.ID, err)
	}
	if !resp.IsOK() && !resp.IsNotfound() {
		return fmt.Errorf("failed to delete rocketmq, %s/%s, statuscode: %v, body: %v", ns, sg.ID, resp.StatusCode(), b.String())
	}
	if err := r.ns.Delete(ns); err != nil {
		return fmt.Errorf("failed to delete namespace, %s, err: %v", ns, err)
	}
	return nil
}

func (r *RocketMQOperator) Update(k8syml interface{}) error {
	rocketMQ, ok := k8syml.(rocketmqv1alpha1.RocketMQ)
	if !ok {
		return fmt.Errorf("invalid type, the k8syml should be RocketMQ")
	}
	if err := r.ns.Exists(rocketMQ.Namespace); err != nil {
		return err
	}

	ro, err := r.getRocketMQ(rocketMQ.Namespace, rocketMQ.Name)
	if err != nil {
		return err
	}

	rocketMQ.ResourceVersion = ro.ResourceVersion
	rocketMQ.Annotations = ro.Annotations

	var b bytes.Buffer
	resp, err := r.client.Put(r.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/addons.erda.cloud/v1alpha1/namespaces/%s/rocketmqs/%s", rocketMQ.Namespace, rocketMQ.Name)).
		JSONBody(rocketMQ).
		Do().
		Body(&b)
	if err != nil {
		return fmt.Errorf("failed to update rocketmq, %s/%s, err: %v, body: %v", rocketMQ.Namespace, rocketMQ.Name, err, b.String())
	}
	if !resp.IsOK() {
		return fmt.Errorf("failed to update rocketmq, %s/%s, statuscode: %v, body: %v",
			rocketMQ.Namespace, rocketMQ.Name, resp.StatusCode(), b.String())
	}
	return nil
}

func (r *RocketMQOperator) getRocketMQ(namespace, name string) (*rocketmqv1alpha1.RocketMQ, error) {
	var b bytes.Buffer
	resp, err := r.client.Get(r.k8s.GetK8SAddr()).
		Path(fmt.Sprintf("/apis/addons.erda.cloud/v1alpha1/namespaces/%s/rocketmqs/%s", namespace, name)).
		Do().
		Body(&b)
	if err != nil {
		return nil, fmt.Errorf("failed to get rocketmq, %s/%s, err: %v, body: %v", namespace, name, err, b.String())
	}
	if !resp.IsOK() {
		return nil, fmt.Errorf("failed to get rocketmq, %s/%s, statuscode: %v, body: %v",
			namespace, name, resp.StatusCode(), b.String())
	}
	var rocketMQ rocketmqv1alpha1.RocketMQ
	if err := json.NewDecoder(&b).Decode(&rocketMQ); err != nil {
		return nil, err
	}
	return &rocketMQ, nil
}

func (r *RocketMQOperator) convertNameSrv(svc apistructs.Service, affinity *corev1.NodeAffinity, resources corev1.ResourceRequirements) rocketmqv1alpha1.NameServiceSpec {
	var nameSrvSpec rocketmqv1alpha1.NameServiceSpec
	nameSrvSpec.Name = svc.Name
	nameSrvSpec.Image = svc.Image
	nameSrvSpec.Affinity = &corev1.Affinity{NodeAffinity: affinity}
	nameSrvSpec.Size = int32(svc.Scale)
	nameSrvSpec.StorageMode = scStorageMode
	nameSrvSpec.Resources = resources
	scname, capacity := getStorageCapacity(svc)
	nameSrvSpec.Env = convertEnvs(svc.Env)
	nameSrvSpec.Labels = svc.Labels
	nameSrvSpec.Labels["ADDON_ID"] = svc.Env["ADDON_ID"]
	nameSrvSpec.Labels[apistructs.DICE_CLUSTER_NAME.String()] = svc.Env[apistructs.DICE_CLUSTER_NAME.String()]
	nameSrvSpec.Labels[apistructs.EnvDiceOrgName] = svc.Env[apistructs.EnvDiceOrgName]
	nameSrvSpec.Labels[apistructs.EnvDiceOrgID] = svc.Env[apistructs.EnvDiceOrgID]
	nameSrvSpec.EnableMetrics = true
	nameSrvSpec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namesrv-log-storage",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &scname,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse(capacity),
					},
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
			},
		},
	}
	return nameSrvSpec
}

func (r *RocketMQOperator) convertBroker(svc apistructs.Service, affinity *corev1.NodeAffinity, resources corev1.ResourceRequirements) rocketmqv1alpha1.BrokerSpec {
	var brokerSpec rocketmqv1alpha1.BrokerSpec
	brokerSpec.Name = svc.Name
	brokerSpec.Image = svc.Image
	brokerSpec.Affinity = &corev1.Affinity{NodeAffinity: affinity}
	brokerSpec.Size = int32(svc.Scale)
	brokerSpec.StorageMode = scStorageMode
	brokerSpec.Resources = resources
	scname, capacity := getStorageCapacity(svc)
	brokerSpec.Env = convertEnvs(svc.Env)
	brokerSpec.Labels = svc.Labels
	brokerSpec.Labels["ADDON_ID"] = svc.Env["ADDON_ID"]
	brokerSpec.Labels[apistructs.DICE_CLUSTER_NAME.String()] = svc.Env[apistructs.DICE_CLUSTER_NAME.String()]
	brokerSpec.Labels[apistructs.EnvDiceOrgName] = svc.Env[apistructs.EnvDiceOrgName]
	brokerSpec.Labels[apistructs.EnvDiceOrgID] = svc.Env[apistructs.EnvDiceOrgID]
	brokerSpec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "broker-data-storage",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &scname,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse(capacity),
					},
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
			},
		},
	}
	return brokerSpec
}

func (r *RocketMQOperator) convertConsole(svc apistructs.Service, affinity *corev1.NodeAffinity, resources corev1.ResourceRequirements) rocketmqv1alpha1.ConsoleSpec {
	var consoleSpec rocketmqv1alpha1.ConsoleSpec
	consoleSpec.Name = svc.Name
	consoleSpec.Image = svc.Image
	consoleSpec.Affinity = &corev1.Affinity{NodeAffinity: affinity}
	consoleSpec.Resources = resources
	return consoleSpec
}

func convertEnvs(envs map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for k, v := range envs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}

func getStorageCapacity(svc apistructs.Service) (string, string) {
	scName := apistructs.DiceLocalVolumeSC
	capacity := "20Gi"
	if len(svc.Volumes) > 0 {
		if svc.Volumes[0].SCVolume.Capacity >= diceyml.AddonVolumeSizeMin && svc.Volumes[0].SCVolume.Capacity <= diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", svc.Volumes[0].SCVolume.Capacity)
		}

		if svc.Volumes[0].SCVolume.Capacity > diceyml.AddonVolumeSizeMax {
			capacity = fmt.Sprintf("%dGi", diceyml.AddonVolumeSizeMax)
		}
		if svc.Volumes[0].SCVolume.StorageClassName != "" {
			scName = svc.Volumes[0].SCVolume.StorageClassName
		}
	}
	return scName, capacity
}

func genK8SNamespace(namespace, name string) string {
	return strutil.Concat(namespace, "--", name)
}
