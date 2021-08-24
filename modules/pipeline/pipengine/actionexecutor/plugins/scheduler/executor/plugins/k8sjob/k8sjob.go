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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubernetes/pkg/kubelet/events"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

var Kind = types.Kind("k8sjob")

var (
	defaultParallelism int32 = 1
	defaultCompletions int32 = 1
	// By default, k8s job has 6 retry opportunities.
	// Based on the existing business, the number of retries is set to 0, either success or failure
	defaultBackoffLimit int32 = 0
)

const (
	executorKind       = "K8SJOB"
	jobKind            = "Job"
	jobAPIVersion      = "batch/v1"
	initContainerName  = "pre-fetech-container"
	emptyDirVolumeName = "pre-fetech-volume"
	EnvRetainNamespace = "RETAIN_NAMESPACE"
)

const (
	ENABLE_SPECIFIED_K8S_NAMESPACE = "ENABLE_SPECIFIED_K8S_NAMESPACE"
)

var (
	errMissingNamespace = errors.New("action missing namespace")
	errMissingUUID      = errors.New("action missing UUID")
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

type K8sJob struct {
	name        types.Name
	client      *k8sclient.K8sClient
	clusterName string
	cluster     apistructs.ClusterInfo
}

func New(name types.Name, clusterName string, cluster apistructs.ClusterInfo) (*K8sJob, error) {
	k, err := k8sclient.New(clusterName)
	if err != nil {
		return nil, err
	}
	return &K8sJob{name: name, client: k, clusterName: clusterName, cluster: cluster}, nil
}

func (k *K8sJob) Kind() types.Kind {
	return Kind
}

func (k *K8sJob) Name() types.Name {
	return k.name
}

func (k *K8sJob) Status(ctx context.Context, action *spec.PipelineTask) (desc apistructs.StatusDesc, err error) {
	var (
		job     *batchv1.Job
		jobPods *corev1.PodList
	)
	jobName := logic.MakeJobName(action)
	job, err = k.client.ClientSet.BatchV1().Jobs(action.Extra.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			desc.Status = apistructs.StatusNotFoundInCluster
			return desc, nil
		}
		return
	}

	if job.Spec.Selector != nil {
		matchlabels := []string{}
		for k, v := range job.Spec.Selector.MatchLabels {
			matchlabels = append(matchlabels, fmt.Sprintf("%s=%v", k, v))
		}
		selector := strutil.Join(matchlabels, ",", true)

		jobPods, err = k.client.ClientSet.CoreV1().Pods(action.Extra.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return
		}
	}

	lastMsg, err := k.getLastMsg(ctx, action.Extra.Namespace, jobName)
	if err != nil {
		return
	}

	//status := generatePipelineStatus(job, jobPods)
	desc = generateKubeJobStatus(job, jobPods, lastMsg)
	return
}

func (k *K8sJob) Create(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	job, err := logic.TransferToSchedulerJob(action)
	if err != nil {
		return nil, err
	}
	if err = k.createNamespace(ctx, job.Namespace); err != nil {
		return nil, err
	}

	if err := k.createImageSecretIfNotExist(job.Namespace); err != nil {
		return nil, err
	}

	if len(job.Volumes) != 0 {
		_, _, pvcs := logic.GenerateK8SVolumes(&job)
		for _, pvc := range pvcs {
			if pvc == nil {
				continue
			}
			_, err := k.client.ClientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
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

	kubeJob, err := k.generateKubeJob(job)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s job")
	}

	_, err = k.client.ClientSet.BatchV1().Jobs(job.Namespace).Create(ctx, kubeJob, metav1.CreateOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to create k8s job, name: %s, err: %v", kubeJob.Name, err)
		logrus.Errorf(errMsg)
		return nil, errors.Errorf(errMsg)
	}

	return apistructs.Job{
		JobFromUser: job,
	}, nil
}

func (k *K8sJob) Remove(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	name := makeJobName(task.Extra.Namespace, task.Extra.UUID)
	namespace := job.Namespace
	propagationPolicy := metav1.DeletePropagationBackground

	jb, err := k.client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		logrus.Warningf("get the job %s in namespace %s is not found", name, namespace)
	}

	// when the err is nil, the job and DeletionTimestamp is not nil. scheduler should delete the job.
	if err == nil && jb.DeletionTimestamp == nil {
		logrus.Infof("start to delete job %s", name)
		err = k.client.ClientSet.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return nil, errors.Wrapf(err, "failed to remove k8s job, name: %s", name)
			}
			logrus.Warningf("delete the job %s in namespace %s is not found", name, namespace)
		}
		logrus.Infof("finish to delete job %s", name)

		for index := range job.Volumes {
			pvcName := fmt.Sprintf("%s-%d", name, index)
			logrus.Infof("start to delete pvc %s", pvcName)
			err = k.client.ClientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return nil, errors.Wrapf(err, "failed to remove k8s pvc, name: %s", pvcName)
				}
				logrus.Warningf("the job %s's pvc %s in namespace %s is not found", name, pvcName, namespace)
			}
			logrus.Infof("finish to delete pvc %s", pvcName)
		}
	}
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) == "" {
		jobs, err := k.client.ClientSet.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errMsg := fmt.Errorf("list the job's pod error: %+v", err)
			return nil, errMsg
		}

		remainCount := 0
		if len(jobs.Items) != 0 {
			for _, j := range jobs.Items {
				if j.DeletionTimestamp == nil {
					remainCount++
				}
			}
		}

		retainNamespace, err := strconv.ParseBool(job.Env[EnvRetainNamespace])
		if err != nil {
			logrus.Debugf("parse bool err %v when delete job %s in the namespace %s", err, job.Name, job.Namespace)
			retainNamespace = false
		}
		if remainCount < 1 && retainNamespace == false {
			ns, err := k.client.ClientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					logrus.Warningf("get namespace %s not found", namespace)
					return nil, nil
				}
				errMsg := fmt.Errorf("get the job's namespace error: %+v", err)
				return nil, errMsg
			}

			if ns.DeletionTimestamp == nil {
				logrus.Infof("start to delete the job's namespace %s", namespace)
				err = k.client.ClientSet.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
				if err != nil {
					if !k8serrors.IsNotFound(err) {
						errMsg := fmt.Errorf("delete the job's namespace error: %+v", err)
						return nil, errMsg
					}
					logrus.Warningf("not found the namespace %s", namespace)
				}
				logrus.Infof("clean namespace %s successfully", namespace)
			}
		}
	}
	return task.Extra.UUID, nil
}

func (k *K8sJob) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = k.Remove(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Inspect use kubectl describe pod information, return latest pod description for current job
func (k *K8sJob) Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error) {
	jobPods, err := k.client.ClientSet.CoreV1().Pods(task.Extra.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: logic.MakeJobLabelSelector(task),
	})
	if err != nil {
		return apistructs.TaskInspect{}, err
	}
	if len(jobPods.Items) == 0 {
		return apistructs.TaskInspect{}, errors.Errorf("get empty pods in job: %s", logic.MakeJobName(task))
	}
	d := describe.PodDescriber{k.client.ClientSet}
	s, err := d.Describe(task.Extra.Namespace, jobPods.Items[len(jobPods.Items)-1].Name, describe.DescriberSettings{
		ShowEvents: true,
	})
	if err != nil {
		return apistructs.TaskInspect{}, err
	}
	return apistructs.TaskInspect{Desc: s}, nil
}

func (k *K8sJob) JobVolumeCreate(ctx context.Context, jobVolume apistructs.JobVolume) (string, error) {
	var namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	if namespace == "" {
		namespace = jobVolume.Namespace
	}
	if err := k.createNamespace(ctx, namespace); err != nil {
		return "", err
	}

	sc := logic.WhichStorageClass(jobVolume.Type)
	id := fmt.Sprintf("%s-%s", jobVolume.Namespace, jobVolume.Name)
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
	if err := k.CreatePVCIfNotExists(ctx, &pvc); err != nil {
		return "", err
	}
	return id, nil
}

func (k *K8sJob) CreatePVCIfNotExists(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	_, err := k.client.ClientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(ctx, pvc.Name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Errorf("faile to get pvc, name: %s, err: %v", pvc.Name, err)
		}
		_, createErr := k.client.ClientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
		if createErr != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return errors.Errorf("failed to create pvc, name: %s, err: %v", pvc.Name, createErr)
			}
		}
	}

	return nil
}

func (k *K8sJob) createNamespace(ctx context.Context, name string) error {
	_, err := k.client.ClientSet.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("failed to get k8s namespace %s: %v", name, err)
			logrus.Errorf(errMsg)
			return errors.Errorf(errMsg)
		}

		newNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}

		_, err := k.client.ClientSet.CoreV1().Namespaces().Create(ctx, newNamespace, metav1.CreateOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				errMsg := fmt.Sprintf("failed to create namespace: %v", err)
				logrus.Errorf(errMsg)
				return errors.Errorf(errMsg)
			}
		}
	}
	return nil
}

func (k *K8sJob) generateKubeJob(specObj interface{}) (*batchv1.Job, error) {
	job, ok := specObj.(apistructs.JobFromUser)
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
		vols, volMounts, _ = logic.GenerateK8SVolumes(&job)
	}

	scheduleInfo2, _, err := logic.GetScheduleInfo(k.cluster, string(k.Name()), string(Kind), job)
	if err != nil {
		return nil, err
	}

	backofflimit := int32(job.BackoffLimit)
	kubeJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       jobKind,
			APIVersion: jobAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeJobName(job.Namespace, job.Name),
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
					Tolerations:      logic.GenTolerations(),
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: apistructs.AliyunRegistry}},
					Affinity:         &constraintbuilders.K8S(&scheduleInfo2, nil, nil, nil).Affinity,
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
							ImagePullPolicy: logic.GetPullImagePolicy(),
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
	clusterInfo, err := logic.GetCLusterInfo(k.clusterName)
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
			ImagePullPolicy: logic.GetPullImagePolicy(),
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

func (k *K8sJob) setBinds(pod *corev1.PodTemplateSpec, binds []apistructs.Bind, clusterInfo map[string]string) error {
	for i, bind := range binds {
		if len(bind.HostPath) == 0 || len(bind.ContainerPath) == 0 {
			errMsg := fmt.Sprintf("invalid params, hostPath: %s, containerPath: %s",
				bind.HostPath, bind.ContainerPath)
			logrus.Errorf("failed to generate k8s job (%v)", errMsg)
			continue
		}

		hostPath, err := logic.ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
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

func (k *K8sJob) generateContainerEnvs(job *apistructs.JobFromUser, clusterInfo map[string]string) ([]corev1.EnvVar, error) {
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

func jobLabels() map[string]string {
	return map[string]string{
		// "dice/job": ""
		labelconfig.K8SLabelPrefix + "job": "",
	}
}

// One-stage analysis, the reasons are intuitively visible
var interestedEventCommentFirstMap = map[string]string{
	// events.FailedToInspectImage:     errInspectImage,
	events.ErrImageNeverPullPolicy:  errImageNeverPull,
	events.NetworkNotReady:          errNetworkNotReady,
	events.FailedAttachVolume:       errMountVolume,
	events.FailedMountVolume:        errMountVolume,
	events.VolumeResizeFailed:       errMountVolume,
	events.FileSystemResizeFailed:   errMountVolume,
	events.FailedMapVolume:          errMountVolume,
	events.WarnAlreadyMountedVolume: errAlreadyMountedVolume,
	events.NodeRebooted:             errNodeRebooted,
}

// wo-stage analysis requires event.message analysis
func secondAnalyzePodEventComment(reason, message string) (string, error) {
	switch reason {
	case "FailedScheduling":
		return parseFailedScheduling(message)
	case events.FailedToPullImage:
		return parseFailedReason(message)
	default:
		// TODO: 补充更多的 reason
		return "", errors.Errorf("invalid event reason: %s", reason)
	}
}

// Analyze the reason for scheduling failure
func parseFailedScheduling(message string) (string, error) {
	var (
		totalNodes    int
		notMatchNodes int
		tmpStr        string
	)

	splitFunc := func(r rune) bool {
		return r == ':' || r == ','
	}
	msgSlice := strings.FieldsFunc(message, splitFunc)

	// Node resource is unavailable
	// 1. Labes don't match。 Example："0/8 nodes are available: 8 node(s) didn't match node selector."
	// 2. Insufficient CPU。Example："0/8 nodes are available: 5 Insufficient cpu, 5 node(s) didn't match node selector."
	// 3. Insufficient Memory。Example："0/8 nodes are available: 5 node(s) didn't match node selector, 5 Insufficient memory."
	if strings.Contains(message, "nodes are available") {
		_, err := fmt.Sscanf(message, "0/%d nodes are available: %s", &totalNodes, &tmpStr)
		if err != nil {
			return "", errors.Errorf("failed to parse totalNodes num, body: %s, (%v)", message, err)
		}

		for _, msg := range msgSlice {
			if strings.Contains(msg, "node(s) didn't match node selector") {
				_, err := fmt.Sscanf(msg, "%d node(s) didn't match node selector", &notMatchNodes)
				if err != nil {
					return "", errors.Errorf("failed to parse notMatchNodes num, body: %s, (%v)", msg, err)
				}
			}
		}

		if totalNodes > 0 && (totalNodes == notMatchNodes) {
			return errNodeSelectorMismatching, nil
		}

		if strings.Contains(message, "Insufficient cpu") {
			return errInsufficientFreeCPU, nil
		}

		if strings.Contains(message, "Insufficient memory") {
			return errInsufficientFreeMemory, nil
		}
	}

	// TODO: Add more information
	return "", errors.New("unexpected")
}

// Analyze the reason for the failure
func parseFailedReason(message string) (string, error) {
	switch {
	// Invalid image name
	case strings.Contains(message, "InvalidImageName"):
		return errInvalidImageName, nil
	// Invalid image
	case strings.Contains(message, "ImagePullBackOff"):
		return errPullImage, nil
	default:
		// TODO: Analyze the reason for the failure
		return "", errors.New("unexpected")
	}
}

func (k *K8sJob) createImageSecretIfNotExist(namespace string) error {
	var err error

	if _, err = k.client.ClientSet.CoreV1().Secrets(namespace).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{}); err == nil {
		return nil
	}

	if !k8serrors.IsNotFound(err) {
		return err
	}

	// When the cluster is initialized, a secret to pull the mirror will be created in the default namespace
	s, err := k.client.ClientSet.CoreV1().Secrets(metav1.NamespaceDefault).Get(context.Background(), apistructs.AliyunRegistry, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
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

	if _, err = k.client.ClientSet.CoreV1().Secrets(namespace).Create(context.Background(), mysecret, metav1.CreateOptions{}); err != nil {
		if strutil.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}

func generatePipelineStatus(job *batchv1.Job, jobPods *corev1.PodList) (status apistructs.PipelineStatus) {
	var podsPending bool
	for _, pod := range jobPods.Items {
		if pod.Status.Phase == corev1.PodPending {
			podsPending = true
		}
	}

	if job.Status.StartTime == nil {
		status = apistructs.PipelineStatusQueue
		return
	}

	if job.Status.Failed > 0 {
		status = apistructs.PipelineStatusFailed
		return
	} else if job.Status.CompletionTime == nil {
		if job.Status.Active > 0 && !podsPending {
			status = apistructs.PipelineStatusRunning
		} else {
			status = apistructs.PipelineStatusQueue
		}
	} else {
		if job.Status.Succeeded >= *job.Spec.Completions {
			status = apistructs.PipelineStatusSuccess
		} else {
			status = apistructs.PipelineStatusFailed
		}
	}
	return
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

func makeJobName(namespace string, taskUUID string) string {
	return strutil.Concat(namespace, ".", taskUUID)
}
