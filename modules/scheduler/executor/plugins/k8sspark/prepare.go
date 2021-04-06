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

package k8sspark

import (
	"fmt"
	"strconv"
	"strings"

	sparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8sjob"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *k8sSpark) generateKubeSparkJob(job *apistructs.Job) (*sparkv1beta2.SparkApplication, error) {
	sparkApp := &sparkv1beta2.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       jobKind,
			APIVersion: jobAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.BigdataConf.Name,
			Namespace: job.BigdataConf.Namespace,
			Labels: map[string]string{
				"job-type":      "k8s-spark",
				"spark-version": s.sparkVersion,
			},
		},
		Spec: sparkv1beta2.SparkApplicationSpec{
			Type:            sparkv1beta2.SparkApplicationType(job.BigdataConf.Spec.SparkConf.Type),
			SparkVersion:    s.sparkVersion,
			Mode:            sparkv1beta2.DeployMode(job.BigdataConf.Spec.SparkConf.Kind),
			Image:           &job.BigdataConf.Spec.Image,
			ImagePullPolicy: stringptr(imagePullPolicyAlways),
			// FIXME: get default secret to fix sa
			ImagePullSecrets: []string{k8s.AliyunRegistry},
			MainClass:        stringptr(job.BigdataConf.Spec.Class),
			Arguments:        job.BigdataConf.Spec.Args,
			RestartPolicy: sparkv1beta2.RestartPolicy{
				Type: sparkv1beta2.Never,
			},
			SparkConf: job.BigdataConf.Spec.Properties,
			// FailureRetries: int32ptr(int32(0)), // never retry
		},
	}

	if sparkApp.Spec.Type == sparkv1beta2.PythonApplicationType {
		sparkApp.Spec.PythonVersion = stringptr("3")
		if job.BigdataConf.Spec.SparkConf.PythonVersion != nil {
			sparkApp.Spec.PythonVersion = job.BigdataConf.Spec.SparkConf.PythonVersion
		}
	}

	jarPath, err := addMainApplicationFile(job)
	if err != nil {
		return nil, err
	}
	sparkApp.Spec.MainApplicationFile = &jarPath

	// Volumes
	vols, volMounts, _ := k8sjob.GenerateK8SVolumes(job)

	// PreFetcher
	if job.PreFetcher != nil && job.PreFetcher.FileFromHost != "" {
		clusterInfo, err := s.clusterInfo.Get()
		if err != nil {
			return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)", s.clusterName, err)
		}
		hostPath, err := k8s.ParseJobHostBindTemplate(job.PreFetcher.FileFromHost, clusterInfo)
		if err != nil {
			return nil, err
		}

		vols = append(vols, corev1.Volume{
			Name: prefetechVolumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		})
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      prefetechVolumeName,
			MountPath: job.PreFetcher.ContainerPath,
			ReadOnly:  false, // rw
		})
	}
	sparkApp.Spec.Volumes = vols

	// DriverSpec
	sparkApp.Spec.Driver.SparkPodSpec = s.composePodSpec(job, "driver", volMounts)
	sparkApp.Spec.Driver.ServiceAccount = stringptr(sparkServiceAccountName)

	// ExecutorSpec
	sparkApp.Spec.Executor.SparkPodSpec = s.composePodSpec(job, "executor", volMounts)
	sparkApp.Spec.Executor.Instances = int32ptr(job.BigdataConf.Spec.SparkConf.ExecutorResource.Replica)

	// TODO: optimization driver & executor JavaOptions
	// TODO: SparkConf
	// TODO: HadoopConf

	logrus.Debugf("generate k8s spark application struct: %+v", sparkApp)
	return sparkApp, nil
}

func addMainApplicationFile(job *apistructs.Job) (string, error) {
	var appResource = job.BigdataConf.Spec.Resource

	if strings.HasPrefix(appResource, "local://") || strings.HasPrefix(appResource, "http://") {
		return appResource, nil
	}

	if strings.HasPrefix(appResource, "/") {
		return strutil.Concat("local://", appResource), nil
	}

	return "", errors.Errorf("invalid job spec, resource: %s", appResource)
}

func (s *k8sSpark) composePodSpec(job *apistructs.Job, podType string, mount []corev1.VolumeMount) sparkv1beta2.SparkPodSpec {
	podSpec := sparkv1beta2.SparkPodSpec{}

	resource := apistructs.BigdataResource{}

	switch podType {
	case "driver":
		resource = job.BigdataConf.Spec.SparkConf.DriverResource
	case "executor":
		resource = job.BigdataConf.Spec.SparkConf.ExecutorResource
	}

	s.appendResource(&podSpec, &resource)
	podSpec.Env = job.BigdataConf.Spec.Envs
	s.appendEnvs(&podSpec, &resource)
	podSpec.Labels = addLabels()
	podSpec.VolumeMounts = mount
	return podSpec
}

func (s *k8sSpark) appendResource(podSpec *sparkv1beta2.SparkPodSpec, resource *apistructs.BigdataResource) {
	// 资源不超卖
	// executor cores 必须大于 1c
	cpu, err := strconv.ParseInt(resource.CPU, 10, 32)
	if err != nil {
		cpu = 1
		logrus.Error(err)
	}
	if cpu < 1 {
		cpu = 1
	}

	cpuString := strutil.Concat(strconv.Itoa(int(cpu)))
	// spark-submit 会将 m 转换成 Mi
	memory := strutil.Concat(resource.Memory, "m")

	podSpec.Cores = int32ptr(int32(cpu))
	podSpec.CoreLimit = stringptr(cpuString)
	podSpec.Memory = stringptr(memory)
}

// 环境变量注入参考：https://dice.app.terminus.io/workBench/projects/70/apps/178/repo/tree/master/docs/dice-env-vars.md
func (s *k8sSpark) appendEnvs(podSpec *sparkv1beta2.SparkPodSpec, resource *apistructs.BigdataResource) {
	// 资源不超卖
	// executor cores 必须大于 1c
	cpu, err := strconv.ParseFloat(resource.CPU, 32)
	if err != nil {
		cpu = 1.0
		logrus.Error(err)
	}
	if cpu < 1.0 {
		cpu = 1.0
	}

	var envMap = map[string]string{
		"DICE_CPU_ORIGIN":  resource.CPU,
		"DICE_MEM_ORIGIN":  resource.Memory,
		"DICE_CPU_REQUEST": fmt.Sprintf("%f", cpu),
		"DICE_MEM_REQUEST": resource.Memory,
		"DICE_CPU_LIMIT":   fmt.Sprintf("%f", cpu),
		"DICE_MEM_LIMIT":   resource.Memory,
		"IS_K8S":           "true",
	}

	for k, v := range envMap {
		podSpec.Env = append(podSpec.Env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	ciEnvs, err := s.clusterInfo.Get()
	if err != nil {
		logrus.Errorf("failed to add spark job envs (%v)", err)
	}

	if len(ciEnvs) > 0 {
		for k, v := range ciEnvs {
			podSpec.Env = append(podSpec.Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}
}

func addLabels() map[string]string {
	// TODO: 现在无法直接使用job.Labels，不符合k8s labels的规
	// if job.Labels == nil {
	// 	job.Labels = make(map[string]string)
	// }

	labels := make(map[string]string)

	// "dice/job": ""
	jobKey := strutil.Concat(labelconfig.K8SLabelPrefix, apistructs.TagJob)
	labels[jobKey] = ""

	// "dice/bigdata": ""
	bigdataKey := strutil.Concat(labelconfig.K8SLabelPrefix, apistructs.TagBigdata)
	labels[bigdataKey] = ""

	return labels
}

func (s *k8sSpark) preparePVCForJob(job *apistructs.Job) error {
	_, _, pvcs := k8sjob.GenerateK8SVolumes(job)
	logrus.Infof("create spark application pvc: %v, vol: %+v", pvcs, job.Volumes)

	for _, pvc := range pvcs {
		if pvc == nil {
			continue
		}
		if err := s.pvc.Create(pvc); err != nil {
			return err
		}
	}
	for i := range pvcs {
		if pvcs[i] == nil {
			continue
		}
		job.Volumes[i].ID = &(pvcs[i].Name)
	}

	return nil
}

func (s *k8sSpark) prepareNamespaceResouce(ns string) error {
	var err error

	if err = s.createNamespaceIfNotExist(ns); err != nil {
		return errors.Errorf("failed to create sparkapplication namespace, ns: %s, (%v)", ns, err)
	}

	// imageSecrets
	if err = s.createImageSecretIfNotExist(ns, k8s.AliyunRegistry); err != nil {
		return errors.Errorf("failed to create aliyun-registry imageSecrets, namespace: %s, (%v)",
			ns, err)
	}

	// spark sa
	if err = s.createSparkServiceAccountIfNotExist(ns); err != nil {
		return errors.Errorf("failed to create spark serviceaccount, namespace: %s, (%v)",
			ns, err)
	}

	// spark role
	if err = s.createSparkRoleIfNotExist(ns); err != nil {
		return errors.Errorf("failed to create spark role, namespace: %s, (%v)",
			ns, err)
	}

	// spark rolebinding
	if err = s.createSparkRolebindingIfNotExist(ns); err != nil {
		return errors.Errorf("failed to create spark rolebinding, namespace: %s, (%v)",
			ns, err)
	}

	return nil
}

func (s *k8sSpark) createNamespaceIfNotExist(ns string) error {
	if s.namespace.Exists(ns) != nil {
		if err := s.namespace.Create(ns, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *k8sSpark) createSparkServiceAccountIfNotExist(ns string) error {
	if s.sa.Exists(ns, sparkServiceAccountName) != nil {
		if err := s.createSparkServiceAccount(ns); err != nil {
			return err
		}
	}

	return nil
}

func (s *k8sSpark) createSparkRoleIfNotExist(ns string) error {
	if err := s.role.Exists(ns, sparkRoleName); err != nil {
		if err := s.createSparkRole(ns); err != nil {
			return err
		}
	}

	return nil
}

func (s *k8sSpark) createSparkRolebindingIfNotExist(ns string) error {
	if err := s.rolebinding.Exists(ns, sparkRoleBindingName); err != nil {
		if err := s.createSparkRolebinding(ns); err != nil {
			return err
		}
	}

	return nil
}

func (s *k8sSpark) createSparkServiceAccount(ns string) error {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sparkServiceAccountName,
			Namespace: ns,
		},
	}

	if err := s.sa.Create(sa); err != nil {
		return err
	}

	return nil
}

func (s *k8sSpark) createSparkRole(ns string) error {
	r := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbacAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sparkRoleName,
			Namespace: ns,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"*"},
			},
		},
	}

	if err := s.role.Create(r); err != nil {
		return err
	}

	return nil
}

func (s *k8sSpark) createSparkRolebinding(ns string) error {
	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sparkRoleBindingName,
			Namespace: ns,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sparkServiceAccountName,
				Namespace: ns,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     sparkRoleName,
			APIGroup: rbacAPIGroup,
		},
	}

	if err := s.rolebinding.Create(rb); err != nil {
		return err
	}

	return nil
}

func (s *k8sSpark) createImageSecretIfNotExist(ns, defaultSecret string) error {
	if _, err := s.secret.Get(ns, k8s.AliyunRegistry); err == nil {
		return nil
	}

	// 集群初始化的时候会在 default namespace 下创建一个拉镜像的 secret
	se, err := s.secret.Get("default", defaultSecret)
	if err != nil {
		return err
	}
	mysecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      se.Name,
			Namespace: ns,
		},
		Data:       se.Data,
		StringData: se.StringData,
		Type:       se.Type,
	}

	if err := s.secret.Create(mysecret); err != nil {
		return err
	}
	return nil
}
