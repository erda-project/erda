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

package k8sjob

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/event"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/clientgo/kubernetes"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	executorKind       = "K8SJOB"
	jobKind            = "Job"
	jobAPIVersion      = "batch/v1"
	initContainerName  = "pre-fetech-container"
	emptyDirVolumeName = "pre-fetech-volume"
)

var (
	defaultParallelism int32 = 1
	defaultCompletions int32 = 1
	// By default, k8s job has 6 retry opportunities.
	// Based on the existing business, the number of retries is set to 0, either success or failure
	defaultBackoffLimit int32 = 0
)

const (
	ENABLE_SPECIFIED_K8S_NAMESPACE = "ENABLE_SPECIFIED_K8S_NAMESPACE"
	// Specify Image Pull Policy with IfNotPresent,Always,Never
	SpecifyImagePullPolicy = "SPECIFY_IMAGE_PULL_POLICY"
)

func init() {
	executortypes.Register(executorKind, func(name executortypes.Name, clusterName string, options map[string]string, optionsPlus interface{}) (
		executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found k8s address in env variables")
		}
		var (
			client *kubernetes.Clientset
			err    error
		)
		client, err = kubernetes.NewKubernetesClientSet("")
		if err != nil {
			return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
				name, clusterName, err)
		}
		if strings.HasPrefix(addr, "inet://") {
			client, err = kubernetes.NewKubernetesClientSet(addr)
			if err != nil {
				return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
					name, clusterName, err)
			}
		}

		clusterInfo, err := clusterinfo.New(clusterName, clusterinfo.WithKubernetesClient(client))
		if err != nil {
			return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
				name, clusterName, err)
		}

		// Synchronize cluster info (every 10m)
		go clusterInfo.LoopLoadAndSync(context.Background(), false)

		return &k8sJob{
			name:        name,
			clusterName: clusterName,
			options:     options,
			addr:        addr,
			client:      client,
			clusterInfo: clusterInfo,
			event:       event.New(event.WithKubernetesClient(client)),
		}, nil
	})
}

type k8sJob struct {
	name        executortypes.Name
	clusterName string
	options     map[string]string
	addr        string
	client      *kubernetes.Clientset
	clusterInfo *clusterinfo.ClusterInfo
	event       *event.Event
}

// Kind executor kind
func (k *k8sJob) Kind() executortypes.Kind {
	return executorKind
}

// Name executor name
func (k *k8sJob) Name() executortypes.Name {
	return k.name
}

// Create create k8s job
func (k *k8sJob) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	job := specObj.(apistructs.Job)

	var namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)

	if namespace == "" {
		namespace = job.Namespace
		err := k.createNamespace(ctx, namespace)
		if err != nil {
			return nil, err
		}
	}

	if err := k.createImageSecretIfNotExist(namespace); err != nil {
		return nil, err
	}

	if len(job.Volumes) != 0 {
		_, _, pvcs := GenerateK8SVolumes(&job)
		for _, pvc := range pvcs {
			if pvc == nil {
				continue
			}
			_, err := k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
		}

		for i := range pvcs {
			if pvcs[i] == nil {
				continue
			}
			job.Volumes[i].ID = &(pvcs[i].Name)
		}
	}

	kubeJob, err := k.generateKubeJob(specObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s job")
	}

	_, err = k.client.BatchV1().Jobs(namespace).Create(ctx, kubeJob, metav1.CreateOptions{})

	name := kubeJob.Name
	if err != nil {
		errMsg := fmt.Sprintf("failed to create k8s job, name: %s", name)
		logrus.Errorf(errMsg)
		return nil, errors.Errorf(errMsg)
	}

	return job, nil
}

// Destroy Equivalent to Remove
func (k *k8sJob) Destroy(ctx context.Context, specObj interface{}) error {
	return k.Remove(ctx, specObj)
}

// Status Query k8s job status
func (k *k8sJob) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var (
		statusDesc apistructs.StatusDesc
		job        *batchv1.Job
		jobpods    *corev1.PodList
		err        error
	)

	kubeJob := specObj.(apistructs.Job)

	namespace := kubeJob.Namespace
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}
	name := strutil.Concat(kubeJob.Namespace, ".", kubeJob.Name)

	job, err = k.client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			statusDesc.Status = apistructs.StatusNotFoundInCluster
			return statusDesc, nil
		}
		return statusDesc, errors.Wrapf(err, "failed to get k8s job status, name: %s", name)
	}

	if job.Spec.Selector != nil {
		matchlabels := []string{}
		for k, v := range job.Spec.Selector.MatchLabels {
			matchlabels = append(matchlabels, fmt.Sprintf("%s=%v", k, v))
		}
		selector := strutil.Join(matchlabels, ",", true)

		jobpods, err = k.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return statusDesc, errors.Wrapf(err, "failed to get job pods, name: %s, err: %v", name, err)
		}
	}

	msgList, err := k.event.AnalyzePodEvents(namespace, name)
	if err != nil {
		logrus.Errorf("failed to analyze job events, namespace: %s, name: %s, (%v)",
			namespace, name, err)
	}

	var lastMsg string
	if len(msgList) > 0 {
		lastMsg = msgList[len(msgList)-1].Comment
	}

	statusDesc = generateKubeJobStatus(job, jobpods, lastMsg)

	return statusDesc, nil
}

func (k *k8sJob) removePipelineJobs(namespace string) error {
	return k.client.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
}

// Remove delete k8s job
func (k *k8sJob) Remove(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	var namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	if namespace == "" {
		namespace = job.Namespace
	}

	logrus.Infof("delete job name %s, namespace %s", job.Name, namespace)

	kubeJob, err := k.generateKubeJob(specObj)
	if err != nil {
		return errors.Wrapf(err, "failed to remove k8s job")
	}

	name := kubeJob.Name
	propagationPolicy := metav1.DeletePropagationBackground

	jb, err := k.client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}

	// when the err is nil, the job and DeletionTimestamp is not nil. scheduler should delete the job.
	if err == nil && jb.DeletionTimestamp == nil {
		err = k.client.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		})
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return errors.Wrapf(err, "failed to remove k8s job, name: %s", name)
			}
			logrus.Debugf("the job %s in namespace %s is not found", name, namespace)
		}

		for index := range job.Volumes {
			pvcName := fmt.Sprintf("%s-%s-%d", namespace, job.Name, index)
			err = k.client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					return errors.Wrapf(err, "failed to remove k8s pvc, name: %s", pvcName)
				}
				logrus.Debugf("the pvc %s in namespace %s is not found", name, namespace)
			}
		}
	}
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) == "" {
		jobs, err := k.client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errMsg := fmt.Errorf("list the job's pod error: %+v", err)
			return errMsg
		}
		remainCount := 0
		if len(jobs.Items) == 0 {
			for _, j := range jobs.Items {
				if j.DeletionTimestamp == nil {
					remainCount++
				}
			}
		}

		if remainCount < 1 {
			ns, err := k.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return nil
				}
				errMsg := fmt.Errorf("get the job's namespace error: %+v", err)
				return errMsg
			}
			if ns.DeletionTimestamp == nil {
				logrus.Infof("delete the job's namespace %s", namespace)
				err = k.client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
				if err != nil {
					if !strings.Contains(err.Error(), "not found") {
						errMsg := fmt.Errorf("delete the job's namespace error: %+v", err)
						return errMsg
					}
				}
			}
		}
	}
	return nil
}

// Update update k8s job
func (k *k8sJob) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	var (
		kubeJob *batchv1.Job
		err     error
	)

	kubeJob, err = k.generateKubeJob(specObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update k8s job")
	}

	if err = k.updateK8SJob(*kubeJob); err != nil {
		return nil, err
	}

	return nil, nil
}

// Inspect View k8s job details
func (k *k8sJob) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.New("job(k8s) not support inspect action")
}

// Cancel stop k8s job
func (k *k8sJob) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}
	var namespace = job.Namespace
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	name := strutil.Concat(namespace, ".", job.Name)

	// Stop the job by setting job.spec.parallelism = 0
	return nil, k.setJobParallelism(namespace, name, 0)
}
func (k *k8sJob) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

func (k *k8sJob) generateKubeJob(specObj interface{}) (*batchv1.Job, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	//logrus.Debugf("input object to k8s job, body: %+v", job)

	cpu := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.CPU*1000)), "m"))
	memory := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.Memory)), "Mi"))

	var (
		vols      []corev1.Volume
		volMounts []corev1.VolumeMount
	)

	if len(job.Volumes) != 0 {
		vols, volMounts, _ = GenerateK8SVolumes(&job)
	}

	var imagePullPolicy corev1.PullPolicy
	switch corev1.PullPolicy(os.Getenv(SpecifyImagePullPolicy)) {
	case corev1.PullAlways:
		imagePullPolicy = corev1.PullAlways
	case corev1.PullNever:
		imagePullPolicy = corev1.PullNever
	default:
		imagePullPolicy = corev1.PullIfNotPresent
	}

	backofflimit := int32(job.BackoffLimit)
	kubeJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       jobKind,
			APIVersion: jobAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      strutil.Concat(job.Namespace, ".", job.Name),
			Namespace: job.Namespace,
			// TODO: Job.Labels cannot be used directly now, which does not comply with the rules of k8s labels
			//Labels:    job.Labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &defaultParallelism,
			// Completions = nil, It means that as long as there is one success, it will be completed
			Completions: &defaultCompletions,
			// TODO: add ActiveDeadlineSeconds
			//ActiveDeadlineSeconds: &defaultActiveDeadlineSeconds,
			// default value: 6
			BackoffLimit: &backofflimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      job.Name,
					Namespace: job.Namespace,
					Labels:    jobLabels(),
				},
				Spec: corev1.PodSpec{
					Tolerations:      toleration.GenTolerations(),
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: k8s.AliyunRegistry}},
					Affinity:         &constraintbuilders.K8S(&job.ScheduleInfo2, nil, nil, nil).Affinity,
					Containers: []corev1.Container{
						{
							Name:  job.Name,
							Image: job.Image,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpu,
									corev1.ResourceMemory: memory,
									//corev1.ResourceStorage: resource.MustParse(strconv.Itoa(int(job.Disk)) + "M"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpu,
									corev1.ResourceMemory: memory,
									//corev1.ResourceStorage: resource.MustParse(strconv.Itoa(int(job.Disk)) + "M"),
								},
							},
							ImagePullPolicy: imagePullPolicy,
							VolumeMounts:    volMounts,
						},
					},
					RestartPolicy:         corev1.RestartPolicyNever,
					Volumes:               vols,
					EnableServiceLinks:    func(enable bool) *bool { return &enable }(false),
					ShareProcessNamespace: func(b bool) *bool { return &b }(false),
				},
			},
		},
	}

	pod := &kubeJob.Spec.Template
	// According to the current business, only one Pod and one Container are supported
	container := &pod.Spec.Containers[0]

	// cmd
	if job.Cmd != "" {
		container.Command = append(container.Command, []string{"sh", "-c", job.Cmd}...)
	}

	// get cluster info
	clusterInfo, err := k.clusterInfo.Get()
	if err != nil {
		return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)", k.clusterName, err)
	}

	// envs
	env, err := k.generateContainerEnvs(&job, clusterInfo)
	if err != nil {
		logrus.Errorf("failed to set job container envs, name: %s, namespace: %s, (%v)",
			job.Name, job.Namespace, err)
		return nil, err
	}
	container.Env = append(container.Env, env...)

	// volumes
	if err := k.setBinds(pod, job.Binds, clusterInfo); err != nil {
		errMsg := fmt.Sprintf("failed to set job binds, name: %s, (%v)", job.Name, err)
		logrus.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	// preFecther to initContainer
	if job.PreFetcher != nil {
		initContainer := corev1.Container{
			Name:  initContainerName,
			Image: job.PreFetcher.FileFromImage,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    cpu,
					corev1.ResourceMemory: memory,
					//corev1.ResourceStorage: resource.MustParse(strconv.Itoa(int(job.Disk)) + "M"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    cpu,
					corev1.ResourceMemory: memory,
					//corev1.ResourceStorage: resource.MustParse(strconv.Itoa(int(job.Disk)) + "M"),
				},
			},
			ImagePullPolicy: imagePullPolicy,
		}

		volumeMount := corev1.VolumeMount{
			Name:      emptyDirVolumeName,
			MountPath: job.PreFetcher.ContainerPath,
			ReadOnly:  false, // rw
		}
		initContainer.Env = append(initContainer.Env, env...)
		initContainer.VolumeMounts = append(initContainer.VolumeMounts, volumeMount)

		container := &pod.Spec.Containers[0]
		container.VolumeMounts = append(container.VolumeMounts, volumeMount)

		pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: emptyDirVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// logrus.Debugf("generate k8s job, body: %+v", kubeJob)
	return kubeJob, nil
}

func (k *k8sJob) setBinds(pod *corev1.PodTemplateSpec, binds []apistructs.Bind, clusterInfo map[string]string) error {
	for i, bind := range binds {
		if len(bind.HostPath) == 0 || len(bind.ContainerPath) == 0 {
			errMsg := fmt.Sprintf("invalid params, hostPath: %s, containerPath: %s",
				bind.HostPath, bind.ContainerPath)
			logrus.Errorf("failed to generate k8s job (%v)", errMsg)
			continue
		}

		hostPath, err := k8s.ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
		if err != nil {
			return err
		}

		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: "volume" + strconv.Itoa(i),
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		})

		container := &pod.Spec.Containers[0]
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "volume" + strconv.Itoa(i),
			MountPath: bind.ContainerPath,
			ReadOnly:  bind.ReadOnly,
		})
	}

	return nil
}

func jobLabels() map[string]string {
	return map[string]string{
		// "dice/job": ""
		labelconfig.K8SLabelPrefix + "job": "",
	}
}

func (k *k8sJob) generateContainerEnvs(job *apistructs.Job, clusterInfo map[string]string) ([]corev1.EnvVar, error) {
	env := []corev1.EnvVar{}
	envMap := job.Env

	for k, v := range envMap {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	// add K8S label
	env = append(env, corev1.EnvVar{
		Name:  "IS_K8S",
		Value: "true",
	})

	// add namespace label
	env = append(env, corev1.EnvVar{
		Name:  "DICE_NAMESPACE",
		Value: job.Namespace,
	})
	env = append(env,
		corev1.EnvVar{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			}},
		corev1.EnvVar{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			}},
		corev1.EnvVar{
			Name:  "DICE_CPU_ORIGIN",
			Value: fmt.Sprintf("%f", job.CPU),
		},
		corev1.EnvVar{
			Name:  "DICE_MEM_ORIGIN",
			Value: fmt.Sprintf("%f", job.Memory),
		},
		corev1.EnvVar{
			Name:  "DICE_CPU_REQUEST",
			Value: fmt.Sprintf("%f", job.CPU),
		},
		corev1.EnvVar{
			Name:  "DICE_MEM_REQUEST",
			Value: fmt.Sprintf("%f", job.Memory),
		},
		corev1.EnvVar{
			Name:  "DICE_CPU_LIMIT",
			Value: fmt.Sprintf("%f", job.CPU),
		},
		corev1.EnvVar{
			Name:  "DICE_MEM_LIMIT",
			Value: fmt.Sprintf("%f", job.Memory),
		},
	)

	if len(clusterInfo) > 0 {
		for k, v := range clusterInfo {
			env = append(env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	return env, nil
}

func generateKubeJobStatus(job *batchv1.Job, jobpods *corev1.PodList, lastMsg string) apistructs.StatusDesc {
	var statusDesc apistructs.StatusDesc

	var podsPending bool
	for _, pod := range jobpods.Items {
		if pod.Status.Phase == corev1.PodPending {
			podsPending = true
		}
	}

	// job controller Have not yet processed the job
	if job.Status.StartTime == nil {
		statusDesc.Status = apistructs.StatusUnschedulable
		return statusDesc
	}

	if job.Status.Failed > 0 {
		statusDesc.Status = apistructs.StatusStoppedOnFailed
	} else if job.Status.CompletionTime == nil {
		if job.Status.Active > 0 && !podsPending {
			statusDesc.Status = apistructs.StatusRunning
		} else {
			statusDesc.Status = apistructs.StatusUnschedulable
		}
	} else {
		// TODO: How to determine if a job is stopped?
		if job.Status.Succeeded >= *job.Spec.Completions {
			statusDesc.Status = apistructs.StatusStoppedOnOK
		} else {
			statusDesc.Status = apistructs.StatusStoppedOnFailed
		}
	}

	statusDesc.LastMessage = lastMsg
	return statusDesc
}

func (k *k8sJob) inspectK8SJob(ns, name string) (*batchv1.Job, error) {

	kubeJob, err := k.client.BatchV1().Jobs(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to inspect k8s job, name: %s", name)
	}

	return kubeJob, nil
}

func (k *k8sJob) updateK8SJob(job batchv1.Job) error {
	ns := job.Namespace
	name := job.Name

	_, err := k.client.BatchV1().Jobs(ns).Update(context.Background(), &job, metav1.UpdateOptions{})

	if err != nil {
		return errors.Wrapf(err, "failed to update k8s job, name: %s", name)
	}

	return nil
}

func (k *k8sJob) setJobParallelism(ns, name string, parallelism int32) error {
	var (
		kubJob *batchv1.Job
		err    error
	)

	if kubJob, err = k.inspectK8SJob(ns, name); err != nil {
		return err
	}

	kubJob.Spec.Parallelism = &parallelism

	if err = k.updateK8SJob(*kubJob); err != nil {
		return err
	}

	return nil
}
func (k *k8sJob) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in k8sJob")
}

func (k *k8sJob) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

// GenerateK8SVolumes According to job configuration, production volume related configuration
func GenerateK8SVolumes(job *apistructs.Job) ([]corev1.Volume, []corev1.VolumeMount, []*corev1.PersistentVolumeClaim) {
	vols := []corev1.Volume{}
	volMounts := []corev1.VolumeMount{}
	pvcs := []*corev1.PersistentVolumeClaim{}
	for i, v := range job.Volumes {
		var pvcid string
		if v.ID == nil { // if ID == nil, we need to create a new pvc
			sc := whichStorageClass(v.Storage)
			id := fmt.Sprintf("%s-%s-%d", job.Namespace, job.Name, i)
			pvcs = append(pvcs, &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      id,
					Namespace: job.Namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					StorageClassName: &sc,
				},
			})
			pvcid = id
		} else {
			pvcs = append(pvcs, nil) // append a placeholder
			pvcid = *v.ID
		}
		volName := fmt.Sprintf("vol-%d", i)
		vols = append(vols, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcid,
				},
			},
		})
		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: v.Path,
		})
	}
	return vols, volMounts, pvcs
}

func whichStorageClass(tp string) string {
	switch tp {
	case "local":
		return "dice-local-volume"
	case "nfs":
		return "dice-nfs-volume"
	default:
		return "dice-local-volume"
	}
}

func (k *k8sJob) createImageSecretIfNotExist(namespace string) error {
	var err error

	if _, err = k.client.CoreV1().Secrets(namespace).Get(context.Background(), k8s.AliyunRegistry, metav1.GetOptions{}); err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), "not found") {
		return err
	}

	// When the cluster is initialized, a secret to pull the mirror will be created in the default namespace
	s, err := k.client.CoreV1().Secrets(metav1.NamespaceDefault).Get(context.Background(), k8s.AliyunRegistry, metav1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	mysecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: namespace,
		},
		Data:       s.Data,
		StringData: s.StringData,
		Type:       s.Type,
	}

	if _, err = k.client.CoreV1().Secrets(namespace).Create(context.Background(), mysecret, metav1.CreateOptions{}); err != nil {
		if strutil.Contains(err.Error(), "AlreadyExists") {
			return nil
		}
		return err
	}
	return nil
}

func (k *k8sJob) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for k8sjob")
}

func (*k8sJob) CleanUpBeforeDelete() {}

func (k *k8sJob) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	jobvolume := spec.(apistructs.JobVolume)
	var namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	if namespace == "" {
		namespace = jobvolume.Namespace
	}
	if err := k.createNamespace(ctx, namespace); err != nil {
		return "", err
	}
	sc := whichStorageClass(jobvolume.Type)
	id := fmt.Sprintf("%s-%s", jobvolume.Namespace, jobvolume.Name)
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
			StorageClassName: &sc,
		},
	}
	if err := k.CreatePVCIfNotExists(&pvc); err != nil {
		return "", err
	}
	return id, nil
}
func (*k8sJob) KillPod(podname string) error {
	return fmt.Errorf("not support for k8sJob")
}

func (k *k8sJob) createNamespace(ctx context.Context, name string) error {

	_, err := k.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			errMsg := fmt.Sprintf("failed to get k8s namespace %s", name)
			logrus.Errorf(errMsg)
			return errors.Errorf(errMsg)
		}

		newNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}

		_, err = k.client.CoreV1().Namespaces().Create(ctx, newNamespace, metav1.CreateOptions{})
		if err != nil {
			errMsg := fmt.Sprintf("failed to create namespace: %v", err)
			logrus.Errorf(errMsg)
			return errors.Errorf(errMsg)
		}
	}
	return nil
}

func (k *k8sJob) CreatePVCIfNotExists(pvc *corev1.PersistentVolumeClaim) error {
	_, err := k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(context.Background(), pvc.Name, metav1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return errors.Errorf("failed to get pvc, name: %s, (%v)", pvc.Name, err)
		}
		_, createErr := k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
		if createErr != nil {
			return errors.Errorf("failed to create pvc, name: %s, (%v)", pvc.Name, err)
		}
	}

	return nil
}
