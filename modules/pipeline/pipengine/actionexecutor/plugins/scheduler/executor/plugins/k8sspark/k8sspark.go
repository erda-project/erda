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

package k8sspark

import (
	"context"
	"fmt"
	"strconv"

	sparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	types.MustRegister(Kind, func(name types.Name, clusterName string, cluster apistructs.ClusterInfo) (types.TaskExecutor, error) {
		k, err := New(name, clusterName, cluster)
		if err != nil {
			return nil, err
		}
		return k, nil
	})
}

func (k *K8sSpark) Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, fmt.Errorf("invalid job spec")
	}

	ns := &corev1.Namespace{}
	if ns, err = k.client.ClientSet.CoreV1().Namespaces().Get(ctx, job.Namespace, metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("get namespace err: %v", err)
		}

		logrus.Infof("create namespace : %s", job.Namespace)
		ns.Name = job.Namespace
		if ns, err = k.client.ClientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil {
			return nil, fmt.Errorf("create namespace err: %v", err)
		}
	}

	if err := k.createImageSecretIfNotExist(job.Namespace); err != nil {
		return nil, fmt.Errorf("failed to create aliyun-registry image secrets, namespace: %s, err: %v", job.Namespace, err)
	}

	if err := k.createSparkServiceAccountIfNotExist(job.Namespace); err != nil {
		return nil, fmt.Errorf("failed to create spark service account, namespace: %s, err: %v", job.Namespace, err)
	}

	if err := k.createSparkRoleIfNotExist(job.Namespace); err != nil {
		return nil, fmt.Errorf("failed to create spark role, namespace: %s, err: %v", job.Namespace, err)
	}

	if err := k.createSparkRolebindingIfNotExist(job.Namespace); err != nil {
		return nil, fmt.Errorf("failed to create spark rolebinding, namespace: %s, err: %v", job.Namespace, err)
	}

	_, _, pvcs := logic.GenerateK8SVolumes(&job)
	for _, pvc := range pvcs {
		if pvc == nil {
			continue
		}
		if _, err := k.client.ClientSet.CoreV1().PersistentVolumeClaims(job.Namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}
	for i := range pvcs {
		if pvcs[i] == nil {
			continue
		}
		job.Volumes[i].ID = &(pvcs[i].Name)
	}

	conf, err := logic.GetBigDataConf(task)
	if err != nil {
		return nil, fmt.Errorf("failed to get big data conf, err: %v", err)
	}
	sparkApp, err := k.generateKubeSparkJob(&job, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spark appliation, namespace: %s, name: %s, err: %v", job.Namespace, job.Name, err)
	}

	if err := k.client.CRClient.Create(ctx, sparkApp); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return apistructs.Job{
				JobFromUser: job,
			}, nil
		}
		return nil, fmt.Errorf("failed to create spark application, namespace: %s, name: %s", job.Namespace, job.Name)
	}

	logrus.Debugf("succeed to create spark application, namespace: %s, name: %s", job.Namespace, job.Name)
	return apistructs.Job{
		JobFromUser: job,
		StatusDesc: apistructs.StatusDesc{
			Status: apistructs.StatusUnschedulable,
		},
	}, nil
}

func (k *K8sSpark) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error) {
	var statusDesc apistructs.StatusDesc

	sparkApp, err := k.getSparkApp(ctx, task)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get the status of k8s spark job, name: %s, err: %v", task.Extra.UUID, err)
		logrus.Warningf(errMsg)

		if k8serrors.IsNotFound(err) {
			statusDesc.Status = apistructs.StatusNotFoundInCluster
			return statusDesc, nil
		}

		return statusDesc, errors.New(errMsg)
	}

	statusDesc.LastMessage = sparkApp.Status.AppState.ErrorMessage
	switch sparkApp.Status.AppState.State {
	case sparkv1beta2.NewState, sparkv1beta2.SubmittedState:
		statusDesc.Status = apistructs.StatusUnschedulable
	case sparkv1beta2.RunningState:
		statusDesc.Status = apistructs.StatusRunning
	case sparkv1beta2.CompletedState, sparkv1beta2.SucceedingState:
		statusDesc.Status = apistructs.StatusStoppedOnOK
	case sparkv1beta2.FailingState, sparkv1beta2.FailedState, sparkv1beta2.FailedSubmissionState,
		sparkv1beta2.InvalidatingState, sparkv1beta2.PendingRerunState:
		statusDesc.Status = apistructs.StatusStoppedOnFailed
	case sparkv1beta2.UnknownState:
		statusDesc.Status = apistructs.StatusUnknown
	default:
		statusDesc.Status = apistructs.StatusUnknown
		statusDesc.LastMessage = fmt.Sprintf("unknown status, sparkAppState: %v", sparkApp.Status.AppState.State)
	}

	logrus.Debugf("succedd to get spark application status, namespace: %s, name: %s, status: %v",
		task.Extra.Namespace, task.Extra.UUID, statusDesc)

	return statusDesc, nil
}

func (k *K8sSpark) Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	if task.Extra.UUID == "" {
		return nil, k.removePipelineJobs(task.Extra.Namespace)
	}

	sparkApp, err := k.getSparkApp(ctx, task)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Errorf("failed to get k8s spark job, namespace: %s, name: %s", task.Extra.Namespace, task.Extra.UUID)
	}

	if err := k.client.CRClient.Delete(ctx, sparkApp); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Errorf("failed to remove spark application, namespace: %s, name: %s, err: %v", task.Extra.Namespace, task.Extra.UUID, err)
	}
	return nil, nil
}

func (k *K8sSpark) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (interface{}, error) {
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err := k.Remove(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Inspect k8sspark doesn`t support inspect, sparkapp`s logs are too long
func (k *K8sSpark) Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error) {
	return apistructs.TaskInspect{}, errors.New("k8sspark don`t support inspect")
}

func (k *K8sSpark) getSparkApp(ctx context.Context, task *spec.PipelineTask) (*sparkv1beta2.SparkApplication, error) {
	var sparkApp sparkv1beta2.SparkApplication
	if err := k.client.CRClient.Get(ctx, client.ObjectKey{Name: task.Extra.UUID, Namespace: task.Extra.Namespace}, &sparkApp); err != nil {
		return nil, err
	}
	return &sparkApp, nil
}

func (k *K8sSpark) removePipelineJobs(ns string) error {
	if err := k.client.ClientSet.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (k *K8sSpark) createImageSecretIfNotExist(ns string) error {
	if _, err := k.client.ClientSet.CoreV1().Secrets(ns).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{}); err == nil {
		return nil
	}

	s, err := k.client.ClientSet.CoreV1().Secrets(metav1.NamespaceDefault).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{})
	if err != nil {
		return err
	}
	mySecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: ns,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if _, err := k.client.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), mySecret, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (k *K8sSpark) createSparkServiceAccountIfNotExist(ns string) error {
	if _, err := k.client.ClientSet.CoreV1().ServiceAccounts(ns).Get(context.Background(), sparkServiceAccountName, metav1.GetOptions{}); err == nil {
		return nil
	}

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

	if _, err := k.client.ClientSet.CoreV1().ServiceAccounts(ns).Create(context.Background(), sa, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (k *K8sSpark) createSparkRoleIfNotExist(ns string) error {
	if _, err := k.client.ClientSet.RbacV1().Roles(ns).Get(context.Background(), sparkRoleName, metav1.GetOptions{}); err == nil {
		return nil
	}

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

	if _, err := k.client.ClientSet.RbacV1().Roles(ns).Create(context.Background(), r, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func (k *K8sSpark) createSparkRolebindingIfNotExist(ns string) error {
	if _, err := k.client.ClientSet.RbacV1().RoleBindings(ns).Get(context.Background(), sparkRoleBindingName, metav1.GetOptions{}); err == nil {
		return nil
	}

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
				Namespace: ns,
				Name:      sparkServiceAccountName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     sparkRoleName,
			APIGroup: rbacAPIGroup,
		},
	}

	if _, err := k.client.ClientSet.RbacV1().RoleBindings(ns).Create(context.Background(), rb, metav1.CreateOptions{}); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (k *K8sSpark) generateKubeSparkJob(job *apistructs.JobFromUser, conf *apistructs.BigdataConf) (*sparkv1beta2.SparkApplication, error) {
	sparkApp := &sparkv1beta2.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       jobKind,
			APIVersion: jobAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      conf.Name,
			Namespace: conf.Namespace,
			Labels: map[string]string{
				"job-type":      "k8s-spark",
				"spark-version": sparkVersion,
			},
		},
		Spec: sparkv1beta2.SparkApplicationSpec{
			Type:             sparkv1beta2.SparkApplicationType(conf.Spec.SparkConf.Type),
			Image:            &conf.Spec.Image,
			SparkVersion:     sparkVersion,
			Mode:             sparkv1beta2.DeployMode(conf.Spec.SparkConf.Kind),
			ImagePullPolicy:  stringptr(imagePullPolicyAlways),
			ImagePullSecrets: []string{apistructs.AliyunRegistry},
			MainClass:        stringptr(conf.Spec.Class),
			Arguments:        conf.Spec.Args,
			RestartPolicy: sparkv1beta2.RestartPolicy{
				Type: sparkv1beta2.Never,
			},
			SparkConf: conf.Spec.Properties,
		},
	}

	if sparkApp.Spec.Type == sparkv1beta2.PythonApplicationType {
		sparkApp.Spec.PythonVersion = stringptr("3")
		if conf.Spec.SparkConf.PythonVersion != nil {
			sparkApp.Spec.PythonVersion = conf.Spec.SparkConf.PythonVersion
		}
	}

	jarPath, err := addMainApplicationFile(conf)
	if err != nil {
		return nil, err
	}
	sparkApp.Spec.MainApplicationFile = &jarPath

	vols, volMounts, _ := logic.GenerateK8SVolumes(job)

	if job.PreFetcher != nil && job.PreFetcher.FileFromHost != "" {
		clusterInfo, err := logic.GetCLusterInfo(k.clusterName)
		if err != nil {
			return nil, errors.Errorf("failed to get cluster info, cluster name: %s, err: %v", k.clusterName, err)
		}
		hostPath, err := logic.ParseJobHostBindTemplate(job.PreFetcher.FileFromHost, clusterInfo)
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
			ReadOnly:  false,
		})
	}
	sparkApp.Spec.Volumes = vols

	sparkApp.Spec.Driver.SparkPodSpec = k.composePodSpec(conf, "driver", volMounts)
	sparkApp.Spec.Driver.ServiceAccount = stringptr(sparkServiceAccountName)

	sparkApp.Spec.Executor.SparkPodSpec = k.composePodSpec(conf, "executor", volMounts)
	sparkApp.Spec.Executor.Instances = int32ptr(conf.Spec.SparkConf.ExecutorResource.Replica)

	return sparkApp, nil
}

func (k *K8sSpark) composePodSpec(conf *apistructs.BigdataConf, podType string, mount []corev1.VolumeMount) sparkv1beta2.SparkPodSpec {
	podSpec := sparkv1beta2.SparkPodSpec{}

	resource := apistructs.BigdataResource{}

	switch podType {
	case "driver":
		resource = conf.Spec.SparkConf.DriverResource
	case "executor":
		resource = conf.Spec.SparkConf.ExecutorResource
	}

	k.appendResource(&podSpec, &resource)
	podSpec.Env = conf.Spec.Envs
	k.appendEnvs(&podSpec, &resource)
	podSpec.Labels = addLabels()
	podSpec.VolumeMounts = mount
	return podSpec
}

func (k *K8sSpark) appendResource(podSpec *sparkv1beta2.SparkPodSpec, resource *apistructs.BigdataResource) {
	cpu, err := strconv.ParseInt(resource.CPU, 10, 32)
	if err != nil {
		cpu = 1
		logrus.Error(err)
	}
	if cpu < 1 {
		cpu = 1
	}

	cpuString := strutil.Concat(strconv.Itoa(int(cpu)))
	// memory valid example: 1024m
	memory := strutil.Concat(resource.Memory, "")

	podSpec.Cores = int32ptr(int32(cpu))
	podSpec.CoreLimit = stringptr(cpuString)
	podSpec.Memory = stringptr(memory)
}

func (k *K8sSpark) appendEnvs(podSpec *sparkv1beta2.SparkPodSpec, resource *apistructs.BigdataResource) {
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

	clusterInfo, err := logic.GetCLusterInfo(k.clusterName)
	if err != nil {
		logrus.Errorf("failed to add spark job envs %v", err)
	}

	if len(clusterInfo) > 0 {
		for k, v := range clusterInfo {
			podSpec.Env = append(podSpec.Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}
}
