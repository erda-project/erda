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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubelet/events"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1"
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
	jobKind            = "Job"
	jobAPIVersion      = "batch/v1"
	initContainerName  = "pre-fetech-container"
	emptyDirVolumeName = "pre-fetech-volume"
	EnvRetainNamespace = "RETAIN_NAMESPACE"
)

const (
	ENABLE_SPECIFIED_K8S_NAMESPACE = "ENABLE_SPECIFIED_K8S_NAMESPACE"
	// Specify Image Pull Policy with IfNotPresent,Always,Never
	SpecifyImagePullPolicy = "SPECIFY_IMAGE_PULL_POLICY"
)

var (
	errMissingNamespace = errors.New("action missing namespace")
	errMissingUUID      = errors.New("action missing UUID")
)

func init() {
	types.MustRegister(Kind, func(name types.Name, clusterName string, options map[string]string) (types.TaskExecutor, error) {
		k, err := New(name, clusterName, options)
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
}

func New(name types.Name, clustername string, options map[string]string) (*K8sJob, error) {
	k, err := k8sclient.New(clustername)
	if err != nil {
		return nil, err
	}
	return &K8sJob{name: name, client: k, clusterName: clustername}, nil
}

func (k *K8sJob) Kind() types.Kind {
	return Kind
}

func (k *K8sJob) Name() types.Name {
	return k.name
}

func validateAction(action *spec.PipelineTask) error {
	if action.Extra.Namespace == "" {
		return errMissingNamespace
	}
	if action.Extra.UUID == "" {
		return errMissingUUID
	}
	return nil
}

func printActionInfo(action *spec.PipelineTask) string {
	return fmt.Sprintf("pipelineID: %d, id: %d, name: %s, namespace: %s, schedulerJobID: %s",
		action.PipelineID, action.ID, action.Name, action.Extra.Namespace, task_uuid.MakeJobID(action))
}

func (k *K8sJob) Status(ctx context.Context, action *spec.PipelineTask) (desc apistructs.StatusDesc, err error) {
	var (
		job     *batchv1.Job
		jobPods *corev1.PodList
	)
	jobName := strutil.Concat(action.Extra.Namespace, ".", action.Extra.UUID)
	job, err = k.client.ClientSet.BatchV1().Jobs(action.Extra.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		if util.IsNotFound(err) {
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
	job, err := transferToSchedulerJob(action)
	if err != nil {
		return nil, err
	}
	if err = k.createNamespace(ctx, job.Namespace); err != nil {
		return nil, err
	}

	if len(job.Volumes) != 0 {
		_, _, pvcs := GenerateK8SVolumes(&job)
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
	job, err := transferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	kubeJob, err := k.generateKubeJob(job)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to remove k8s job")
	}

	name := kubeJob.Name
	namespace := job.Namespace
	propagationPolicy := metav1.DeletePropagationBackground

	jb, err := k.client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !util.IsNotFound(err) {
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
			if !util.IsNotFound(err) {
				return nil, errors.Wrapf(err, "failed to remove k8s job, name: %s", name)
			}
			logrus.Warningf("delete the job %s in namespace %s is not found", name, namespace)
		}
		logrus.Infof("finish to delete job %s", name)

		for index := range job.Volumes {
			pvcName := fmt.Sprintf("%s-%s-%d", namespace, job.Name, index)
			logrus.Infof("start to delete pvc %s", pvcName)
			err = k.client.ClientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
			if err != nil {
				if !util.IsNotFound(err) {
					return nil, errors.Wrapf(err, "failed to remove k8s pvc, name: %s", pvcName)
				}
				logrus.Warningf("the job %s's pvc %s in namespace %s is not found", job.Name, pvcName, namespace)
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
				if util.IsNotFound(err) {
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
					if !util.IsNotFound(err) {
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

func (k *K8sJob) createNamespace(ctx context.Context, name string) error {
	_, err := k.client.ClientSet.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !util.IsNotFound(err) {
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
					Affinity:         nil,
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
	clusterInfo, err := k.getCLusterInfo()
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

func (k *K8sJob) getCLusterInfo() (map[string]string, error) {
	var clusterInfoRes struct {
		Data map[string]string `json:"data"`
	}

	var body bytes.Buffer
	resp, err := httpclient.New().Get(discover.Scheduler()).
		Path(strutil.Concat("/api/clusterinfo/", k.clusterName)).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Errorf("get cluster info failed, err: %v", err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	if err := json.Unmarshal([]byte(respBody), &clusterInfoRes); err != nil {
		return nil, errors.Errorf("get cluster info failed, statueCode: %d, err: %v", statusCode, err)
	}

	return clusterInfoRes.Data, nil
}

func (k *K8sJob) setBinds(pod *corev1.PodTemplateSpec, binds []apistructs.Bind, clusterInfo map[string]string) error {
	for i, bind := range binds {
		if len(bind.HostPath) == 0 || len(bind.ContainerPath) == 0 {
			errMsg := fmt.Sprintf("invalid params, hostPath: %s, containerPath: %s",
				bind.HostPath, bind.ContainerPath)
			logrus.Errorf("failed to generate k8s job (%v)", errMsg)
			continue
		}

		hostPath, err := ParseJobHostBindTemplate(bind.HostPath, clusterInfo)
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

// ParseJobHostBindTemplate Analyze the hostPath template and convert it to the cluster info value
func ParseJobHostBindTemplate(hostPath string, clusterInfo map[string]string) (string, error) {
	var b bytes.Buffer

	if hostPath == "" {
		return "", errors.New("hostPath is empty")
	}

	t, err := template.New("jobBind").
		Option("missingkey=error").
		Parse(hostPath)
	if err != nil {
		return "", errors.Errorf("failed to parse bind, hostPath: %s, (%v)", hostPath, err)
	}

	err = t.Execute(&b, &clusterInfo)
	if err != nil {
		return "", errors.Errorf("failed to execute bind, hostPath: %s, (%v)", hostPath, err)
	}

	return b.String(), nil
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

func transferToSchedulerJob(task *spec.PipelineTask) (job apistructs.JobFromUser, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}
	}()

	return apistructs.JobFromUser{
		Name: task_uuid.MakeJobID(task),
		Kind: func() string {
			switch task.Type {
			case string(pipelineymlv1.RES_TYPE_FLINK):
				return string(apistructs.Flink)
			case string(pipelineymlv1.RES_TYPE_SPARK):
				return string(apistructs.Spark)
			default:
				return ""
			}
		}(),
		Namespace: task.Extra.Namespace,
		ClusterName: func() string {
			if len(task.Extra.ClusterName) == 0 {
				panic(errors.New("missing cluster name in pipeline task"))
			}
			return task.Extra.ClusterName
		}(),
		Image:      task.Extra.Image,
		Cmd:        strings.Join(append([]string{task.Extra.Cmd}, task.Extra.CmdArgs...), " "),
		CPU:        task.Extra.RuntimeResource.CPU,
		Memory:     task.Extra.RuntimeResource.Memory,
		Binds:      task.Extra.Binds,
		Volumes:    makeVolume(task),
		PreFetcher: task.Extra.PreFetcher,
		Env:        task.Extra.PublicEnvs,
		Labels:     task.Extra.Labels,
		// flink/spark
		Resource:  task.Extra.FlinkSparkConf.JarResource,
		MainClass: task.Extra.FlinkSparkConf.MainClass,
		MainArgs:  task.Extra.FlinkSparkConf.MainArgs,
		// 重试不依赖 scheduler，由 pipeline engine 自己实现，保证所有 action executor 均适用
		Params: task.Extra.Action.Params,
	}, nil
}

func makeVolume(task *spec.PipelineTask) []diceyml.Volume {
	diceVolumes := make([]diceyml.Volume, 0)
	for _, vo := range task.Extra.Volumes {
		if vo.Type == string(spec.StoreTypeDiceVolumeFake) || vo.Type == string(spec.StoreTypeDiceCacheNFS) {
			// fake volume,没有实际挂载行为,不传给scheduler
			continue
		}
		diceVolume := diceyml.Volume{
			Path: vo.Value,
			Storage: func() string {
				switch vo.Type {
				case string(spec.StoreTypeDiceVolumeNFS):
					return "nfs"
				case string(spec.StoreTypeDiceVolumeLocal):
					return "local"
				default:
					panic(errors.Errorf("%q has not supported volume type: %s", vo.Name, vo.Type))
				}
			}(),
		}
		if vo.Labels != nil {
			if id, ok := vo.Labels["ID"]; ok {
				diceVolume.ID = &id
				goto AppendDiceVolume
			}
		}
		// labels == nil or labels["ID"] not exist
		// 如果 id 不存在，说明上一次没有生成 volume，并且是 optional 的，则不创建 diceVolume
		if vo.Optional {
			continue
		}
	AppendDiceVolume:
		diceVolumes = append(diceVolumes, diceVolume)
	}
	return diceVolumes
}

// GenerateK8SVolumes According to job configuration, production volume related configuration
func GenerateK8SVolumes(job *apistructs.JobFromUser) ([]corev1.Volume, []corev1.VolumeMount, []*corev1.PersistentVolumeClaim) {
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
