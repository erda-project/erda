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

package daemonset

import (
	"fmt"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/strutil"
)

type DaemonsetOperator struct {
	k8s         addon.K8SUtil
	ns          addon.NamespaceUtil
	imageSecret addon.ImageSecretUtil
	healthcheck addon.HealthcheckUtil
	daemonset   addon.DaemonsetUtil
	overcommit  addon.OvercommitUtil
}

func New(k8s addon.K8SUtil, ns addon.NamespaceUtil, imageSecret addon.ImageSecretUtil,
	healthcheck addon.HealthcheckUtil, daemonset addon.DaemonsetUtil,
	overcommit addon.OvercommitUtil) *DaemonsetOperator {
	return &DaemonsetOperator{
		k8s:         k8s,
		ns:          ns,
		imageSecret: imageSecret,
		healthcheck: healthcheck,
		daemonset:   daemonset,
		overcommit:  overcommit,
	}
}

func (d *DaemonsetOperator) IsSupported() bool {
	return true
}

func (d *DaemonsetOperator) Validate(sg *apistructs.ServiceGroup) error {
	operator, ok := sg.Labels["USE_OPERATOR"]
	if !ok {
		return fmt.Errorf("[BUG] sg need USE_OPERATOR label")
	}
	if strutil.ToLower(operator) != "daemonset" {
		return fmt.Errorf("[BUG] value of label USE_OPERATOR should be 'daemonset'")
	}
	if len(sg.Services) != 1 {
		return fmt.Errorf("illegal services num: %d", len(sg.Services))
	}
	return nil
}

// TODO: volume support
func (d *DaemonsetOperator) Convert(sg *apistructs.ServiceGroup) interface{} {
	service := sg.Services[0]
	affinity := constraintbuilders.K8S(&sg.ScheduleInfo2, &service, nil, nil).Affinity
	probe := d.healthcheck.NewHealthcheckProbe(&service)
	container := corev1.Container{
		Name:  service.Name,
		Image: service.Image,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.fm", service.Resources.Cpu*1000)),
				corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%.fMi", service.Resources.Mem)),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse(
					fmt.Sprintf("%.fm", d.overcommit.CPUOvercommit(service.Resources.Cpu*1000))),
				corev1.ResourceMemory: resource.MustParse(
					fmt.Sprintf("%.dMi", d.overcommit.MemoryOvercommit(int(service.Resources.Mem)))),
			},
		},
		Command:        []string{"sh", "-c", service.Cmd},
		Env:            envs(service.Env),
		LivenessProbe:  probe,
		ReadinessProbe: probe,
	}
	return appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: genK8SNamespace(sg.Type, sg.ID),
		},
		Spec: appsv1.DaemonSetSpec{
			RevisionHistoryLimit: func(i int32) *int32 { return &i }(int32(3)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": service.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   service.Name,
					Labels: map[string]string{"app": service.Name},
				},
				Spec: corev1.PodSpec{
					EnableServiceLinks:    func(enable bool) *bool { return &enable }(false),
					ShareProcessNamespace: func(b bool) *bool { return &b }(false),
					Containers:            []corev1.Container{container},
					Affinity:              &affinity,
				},
			},
		},
	}
}

func (d *DaemonsetOperator) Create(k8syml interface{}) error {
	ds, ok := k8syml.(appsv1.DaemonSet)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be DaemonSet")
	}
	if err := d.ns.Exists(ds.Namespace); err != nil {
		if err := d.ns.Create(ds.Namespace, nil); err != nil && !strutil.Contains(err.Error(), "AlreadyExists") {
			logrus.Errorf("failed to create ns: %s, %v", ds.Namespace, err)
			return err
		}
	}
	if err := d.imageSecret.NewImageSecret(ds.Namespace); err != nil {
		logrus.Errorf("failed to NewImageSecret for ns: %s, %v", ds.Namespace, err)
		return err
	}
	if err := d.daemonset.Create(&ds); err != nil {
		return err
	}
	return nil
}

func (d *DaemonsetOperator) Inspect(sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	service := &(sg.Services[0])
	ds, err := d.daemonset.Get(genK8SNamespace(sg.Type, sg.ID), service.Name)
	if err != nil {
		logrus.Errorf("failed to get ds: %s/%s, %v", genK8SNamespace(sg.Type, sg.ID), service.Name, err)
		return nil, err
	}
	if ds.Status.DesiredNumberScheduled == ds.Status.NumberAvailable {
		service.Status = apistructs.StatusHealthy
	} else {
		service.Status = apistructs.StatusUnHealthy
	}
	return sg, nil
}

func (d *DaemonsetOperator) Remove(sg *apistructs.ServiceGroup) error {
	if err := d.ns.Delete(genK8SNamespace(sg.Type, sg.ID)); err != nil {
		logrus.Errorf("failed to remove ns: %s, %v", genK8SNamespace(sg.Type, sg.ID), err)
		return err
	}
	return nil
}

func (d *DaemonsetOperator) Update(k8syml interface{}) error {
	ds, ok := k8syml.(appsv1.DaemonSet)
	if !ok {
		return fmt.Errorf("[BUG] this k8syml should be DaemonSet")
	}
	if err := d.daemonset.Update(&ds); err != nil {
		logrus.Errorf("failed to update ds: %s/%s, %v", ds.Namespace, ds.Name, err)
		return err
	}
	return nil
}

func genK8SNamespace(namespace, name string) string {
	return strutil.Concat(namespace, "--", name)
}

func envs(envs map[string]string) []corev1.EnvVar {
	r := []corev1.EnvVar{}
	for k, v := range envs {
		r = append(r, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return r
}
