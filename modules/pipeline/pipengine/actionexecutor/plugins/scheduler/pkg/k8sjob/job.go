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
	"sort"
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

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"

	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"

	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1"

	"github.com/erda-project/erda/modules/scheduler/executor/util"

	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/strutil"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindK8sJob)

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
	// Specify Image Pull Policy with IfNotPresent,Always,Never
	SpecifyImagePullPolicy = "SPECIFY_IMAGE_PULL_POLICY"
)

var (
	errMissingNamespace = errors.New("action missing namespace")
	errMissingUUID      = errors.New("action missing UUID")
)

type K8sJob struct {
	name   types.Name
	client k8sclient.K8sClient
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

func (k *K8sJob) Status(ctx context.Context, action *spec.PipelineTask) (desc apistructs.PipelineStatusDesc, err error) {
	var (
		job     *batchv1.Job
		jobPods *corev1.PodList
	)
	jobName := strutil.Concat(action.Extra.Namespace, ".", action.Extra.UUID)
	job, err = k.client.ClientSet.BatchV1().Jobs(action.Extra.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
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

	status := generatePipelineStatus(job, jobPods)
	desc.Status = status
	desc.Desc = lastMsg
	return
}

func (k *K8sJob) Exist(ctx context.Context, action *spec.PipelineTask) (created, started bool, err error) {
	statusDesc, err := k.Status(ctx, action)
	if err != nil {
		created = false
		started = false
		// 该 ErrMsg 表示记录在 etcd 中不存在，即未创建
		if strutil.Contains(err.Error(), "failed to inspect job, err: not found") {
			err = nil
			return
		}
		// 获取 job 状态失败
		return
	}
	// err 为空，说明在 etcd 中存在记录，即已经创建成功
	created = true

	// 根据状态判断是否实际 job(k8s job, DC/OS job) 是否已开始执行
	switch statusDesc.Status {
	// err
	case apistructs.PipelineStatusError, apistructs.PipelineStatusUnknown:
		err = errors.Errorf("failed to judge job exist or not, detail: %s", statusDesc)
	// not started
	case apistructs.PipelineStatusCreated, apistructs.PipelineStatusStartError:
		started = false
	// started
	case apistructs.PipelineStatusQueue, apistructs.PipelineStatusRunning,
		apistructs.PipelineStatusSuccess, apistructs.PipelineStatusFailed,
		apistructs.PipelineStatusStopByUser:
		started = true

	// default
	default:
		started = false
	}
	return
}

func (k *K8sJob) Create(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	created, _, err := k.Exist(ctx, action)
	if err != nil {
		return nil, err
	}
	if created {
		logrus.Warnf("job already created")
		return nil, nil
	}
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

	return job, nil
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
	//clusterInfo, err := k.clusterInfo.Get()
	//if err != nil {
	//	return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)", k.clusterName, err)
	//}
	// TODO
	clusterInfoStr := `{"CLUSTER_DNS":"10.96.0.3","DICE_CLUSTER_NAME":"terminus-dev","DICE_CLUSTER_TYPE":"kubernetes","DICE_HTTPS_PORT":"443","DICE_HTTP_PORT":"80","DICE_INSIDE":"false","DICE_IS_EDGE":"false","DICE_PROTOCOL":"https,http","DICE_ROOT_DOMAIN":"dev.terminus.io","DICE_SIZE":"test","DICE_STORAGE_MOUNTPOINT":"/netdata","DICE_VERSION":"4.0","ETCD_MONITOR_URL":"http://10.0.6.198:2381","GLUSTERFS_MONITOR_URL":"http://10.0.6.198:24007,http://10.0.6.199:24007,http://10.0.6.200:24007","ISTIO_ALIYUN":"false","ISTIO_INSTALLED":"true","ISTIO_VERSION":"1.1.4","IS_FDP_CLUSTER":"true","KUBERNETES_VENDOR":"dice","KUBERNETES_VERSION":"v1.16.4","LB_ADDR":"10.0.6.199:80","LB_MONITOR_URL":"http://10.0.6.199:80","MASTER_ADDR":"10.0.6.198:6443","MASTER_MONITOR_ADDR":"10.0.6.198:6443","MASTER_VIP_ADDR":"10.96.0.1:443","MASTER_VIP_URL":"https://10.96.0.1:443","NETPORTAL_URL":"inet://ingress-nginx.kube-system.svc.cluster.local?direct=on\u0026ssl=on","NEXUS_ADDR":"addon-nexus.default.svc.cluster.local:8081","NEXUS_PASSWORD":"sybHZT0nh9VQ16T8f5ldKr7223n4a1","NEXUS_USERNAME":"admin","REGISTRY_ADDR":"addon-registry.default.svc.cluster.local:5000"}`
	clusterInfo := map[string]string{}
	if err := json.Unmarshal([]byte(clusterInfoStr), &clusterInfo); err != nil {
		return nil, err
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

// Message Event message
type Message struct {
	Timestamp int64
	Reason    string
	Message   string
	Comment   string
}

// MessageList 事件消息列表
type MessageList []Message

func (em MessageList) Len() int           { return len(em) }
func (em MessageList) Swap(i, j int)      { em[i], em[j] = em[j], em[i] }
func (em MessageList) Less(i, j int) bool { return em[i].Timestamp < em[j].Timestamp }

func (k *K8sJob) getLastMsg(ctx context.Context, namespace, name string) (lastMsg string, err error) {
	var ems MessageList

	eventList, err := k.client.ClientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	for i := range eventList.Items {
		e := &eventList.Items[i]
		if e.InvolvedObject.Kind != "Pod" || !strings.HasPrefix(e.InvolvedObject.Name, name) {
			continue
		}

		// One-stage analysis, the reasons are intuitively visible
		if cmt, ok := interestedEventCommentFirstMap[e.Reason]; ok {
			ems = append(ems, Message{
				Timestamp: e.LastTimestamp.Unix(),
				Reason:    e.Reason,
				Message:   e.Message,
				Comment:   cmt,
			})
			continue
		}

		// Two-stage analysis requires event.message analysis
		cmt, err := secondAnalyzePodEventComment(e.Reason, e.Message)
		if err == nil {
			ems = append(ems, Message{
				Timestamp: e.LastTimestamp.Unix(),
				Reason:    e.Reason,
				Message:   e.Message,
				Comment:   cmt,
			})
		}
	}

	sort.Sort(ems)
	if len(ems) > 0 {
		lastMsg = ems[len(ems)-1].Comment
	}
	return lastMsg, nil
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
