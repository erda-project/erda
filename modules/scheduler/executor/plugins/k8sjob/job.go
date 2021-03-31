package k8sjob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/k8serror"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/modules/scheduler/executor/plugins/k8s/toleration"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	executorKind          = "K8SJOB"
	defaultAPIPrefix      = "/apis/batch/v1/namespaces/"
	defaultNamespace      = "default"
	jobKind               = "Job"
	jobAPIVersion         = "batch/v1"
	initContainerName     = "pre-fetech-container"
	emptyDirVolumeName    = "pre-fetech-volume"
	aliyunImagePullSecret = "aliyun-registry"
)

var (
	defaultParallelism int32 = 1
	defaultCompletions int32 = 1
	// k8s job 默认是有 6 次重试的机会，
	// 基于现有业务，重试次数设置成 0 次，要么成功要么失败
	defaultBackoffLimit int32 = 0
)

// k8s job plugin's configure
//
// EXECUTOR_K8SJOB_K8SJOBFORTERMINUS_ADDR=http://127.0.0.1:8080
// EXECUTOR_K8SJOB_K8SJOBFORJOBTERMINUS_BASICAUTH=admin:1234
// EXECUTOR_K8S_K8SFORSERVICE_BASICAUTH=
// EXECUTOR_K8S_K8SFORSERVICE_CA_CRT=
// EXECUTOR_K8S_K8SFORSERVICE_CLIENT_CRT=
// EXECUTOR_K8S_K8SFORSERVICE_CLIENT_KEY=
// EXECUTOR_K8S_K8SFORSERVICE_BEARER_TOKEN=
func init() {
	executortypes.Register(executorKind, func(name executortypes.Name, clusterName string, options map[string]string, optionsPlus interface{}) (
		executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found k8s address in env variables")
		}

		if !strings.HasPrefix(addr, "inet://") {
			if !strings.HasPrefix(addr, "http") && !strings.HasPrefix(addr, "https") {
				addr = strutil.Concat("http://", addr)
			}
		}

		client := httpclient.New()
		if _, ok := options["CA_CRT"]; ok {
			logrus.Infof("k8s executor(%s) addr for https: %v", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
				[]byte(options["CLIENT_KEY"]),
				[]byte(options["CA_CRT"])))
			token, ok := options["BEARER_TOKEN"]
			if !ok {
				return nil, errors.Errorf("not found k8s bearer token")
			}
			// 默认开启了 RBAC, 需要通过 token 进行用户鉴权
			client.BearerTokenAuth(token)
		}

		basicAuth, ok := options["BASICAUTH"]
		if ok {
			userPasswd := strings.Split(basicAuth, ":")
			if len(userPasswd) == 2 {
				client.BasicAuth(userPasswd[0], userPasswd[1])
			}
		}

		clusterInfo, err := clusterinfo.New(clusterName, clusterinfo.WithCompleteParams(addr, client))
		if err != nil {
			return nil, errors.Errorf("failed to new cluster info, executorName: %s, clusterName: %s, (%v)",
				name, clusterName, err)
		}
		// 同步 cluster info（10m 一次）
		go clusterInfo.LoopLoadAndSync(context.Background(), false)

		return &k8sJob{
			name:        name,
			clusterName: clusterName,
			options:     options,
			addr:        addr,
			prefix:      defaultAPIPrefix,
			client:      client,
			pvc:         persistentvolumeclaim.New(persistentvolumeclaim.WithCompleteParams(addr, client)),
			namespace:   namespace.New(namespace.WithCompleteParams(addr, client)),
			secret:      secret.New(secret.WithCompleteParams(addr, client)),
			clusterInfo: clusterInfo,
			event:       event.New(event.WithCompleteParams(addr, client)),
		}, nil
	})
}

type k8sJob struct {
	name        executortypes.Name
	clusterName string
	options     map[string]string
	addr        string
	prefix      string
	client      *httpclient.HTTPClient
	pvc         *persistentvolumeclaim.PersistentVolumeClaim
	namespace   *namespace.Namespace
	secret      *secret.Secret
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

// Create 创建 k8s job
func (k *k8sJob) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	var b bytes.Buffer
	job := specObj.(apistructs.Job)

	if k.namespace.Exists(job.Namespace) != nil {
		if err := k.namespace.Create(job.Namespace, nil); err != nil {
			logrus.Errorf("failed to create namespace: %v", err)
		}
	}
	if err := k.createImageSecretIfNotExist(job.Namespace); err != nil {
		return nil, err
	}
	_, _, pvcs := GenerateK8SVolumes(&job)
	for _, pvc := range pvcs {
		if pvc == nil {
			continue
		}
		if err := k.pvc.Create(pvc); err != nil {
			return nil, err
		}
	}
	for i := range pvcs {
		if pvcs[i] == nil {
			continue
		}
		job.Volumes[i].ID = &(pvcs[i].Name)
	}

	kubeJob, err := k.generateKubeJob(specObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s job")
	}

	ns := kubeJob.Namespace
	name := kubeJob.Name
	path := strutil.Concat(k.prefix, ns, "/jobs")

	// POST /apis/batch/v1/namespaces/{namespace}/jobs
	resp, err := k.client.Post(k.addr).
		Path(path).
		Header("Content-Type", "application/json").
		JSONBody(kubeJob).
		Do().
		Body(&b)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s job, name: %s", name)
	}

	if resp == nil {
		return nil, errors.Errorf("resp is null")
	}

	if !resp.IsOK() {
		errMsg := fmt.Sprintf("failed to create k8s job, name: %s, statusCode: %d, resp body: %s",
			name, resp.StatusCode(), b.String())
		logrus.Errorf(errMsg)
		return nil, errors.Errorf(errMsg)
	}

	return job, nil
}

// Destroy 等同于 Remove
func (k *k8sJob) Destroy(ctx context.Context, specObj interface{}) error {
	return k.Remove(ctx, specObj)
}

// Status 查询 k8s job 状态
func (k *k8sJob) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var (
		statusDesc apistructs.StatusDesc
		job        batchv1.Job
		jobpods    corev1.PodList
	)

	kubeJob, err := k.generateKubeJob(specObj)
	if err != nil {
		return statusDesc, errors.Wrapf(err, "failed to get k8s job Status")
	}

	ns := kubeJob.Namespace
	name := kubeJob.Name
	path := strutil.Concat(k.prefix, ns, "/jobs/", name)

	// GET /apis/batch/v1/namespaces/{namespace}/jobs/{name}
	var b bytes.Buffer
	resp, err := k.client.Get(k.addr).
		Path(path).
		Header("Content-Type", "application/json").
		Do().
		Body(&b)

	if err != nil {
		return statusDesc, errors.Wrapf(err, "failed to get k8s job status, name: %s", name)
	}

	if resp == nil {
		return statusDesc, errors.Errorf("response is null")
	}

	if !resp.IsOK() {
		errMsg := fmt.Sprintf("failed to get the status of job, name: %s, statusCode: %d, body: %v",
			name, resp.StatusCode(), b.String())
		if resp.StatusCode() == http.StatusNotFound {
			statusDesc.Status = apistructs.StatusNotFoundInCluster
			return statusDesc, nil
		}
		logrus.Errorf(errMsg)
		return statusDesc, errors.Errorf(errMsg)
	}

	if err := json.NewDecoder(&b).Decode(&job); err != nil {
		return statusDesc, err
	}

	if job.Spec.Selector != nil {
		matchlabels := []string{}
		for k, v := range job.Spec.Selector.MatchLabels {
			matchlabels = append(matchlabels, fmt.Sprintf("%s=%v", k, v))
		}
		selector := strutil.Join(matchlabels, ",", true)
		var b bytes.Buffer
		resp, err := k.client.Get(k.addr).
			Path(fmt.Sprintf("/api/v1/namespaces/%s/pods", job.Namespace)).
			Param("labelSelector", selector).
			Header("Content-Type", "application/json").
			Do().
			Body(&b)
		if err != nil {
			return statusDesc, errors.Wrapf(err, "failed to get job pods, name: %s, err: %v", name, err)
		}
		if resp == nil || !resp.IsOK() {
			return statusDesc, errors.Errorf("failed to get job pods, name: %s, resp: %v, body: %v", name, *resp, b.String())
		}
		if err := json.NewDecoder(&b).Decode(&jobpods); err != nil {
			return statusDesc, err
		}

	}

	msgList, err := k.event.AnalyzePodEvents(ns, name)
	if err != nil {
		logrus.Errorf("failed to analyze job events, namespace: %s, name: %s, (%v)",
			ns, name, err)
	}

	var lastMsg string
	if len(msgList) > 0 {
		lastMsg = msgList[len(msgList)-1].Comment
	}

	statusDesc = generateKubeJobStatus(&job, &jobpods, lastMsg)

	return statusDesc, nil
}

func (k *k8sJob) removePipelineJobs(namespace string) error {
	return k.namespace.Delete(namespace)
}

// Remove 删除 k8s job
func (k *k8sJob) Remove(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}
	if job.Name == "" {
		return k.removePipelineJobs(job.Namespace)
	}
	var b bytes.Buffer

	kubeJob, err := k.generateKubeJob(specObj)
	if err != nil {
		return errors.Wrapf(err, "failed to remove k8s job")
	}

	ns := kubeJob.Namespace
	name := kubeJob.Name
	path := strutil.Concat(k.prefix, ns, "/jobs/", name)

	// DELETE /apis/batch/v1/namespaces/{namespace}/jobs/{name}
	resp, err := k.client.Delete(k.addr).
		Path(path).
		Header("Content-Type", "application/json").
		JSONBody(deleteOptions).
		Do().
		Body(&b)
	if err != nil {
		return errors.Wrapf(err, "failed to remove k8s job, name: %s", name)
	}

	if resp == nil {
		return errors.Errorf("response is null")
	}

	if !resp.IsOK() && resp.StatusCode() != 404 {
		errMsg := fmt.Sprintf("failed to remove k8s job, name: %s, statusCode: %d, resp body: %s",
			name, resp.StatusCode(), b.String())
		logrus.Errorf(errMsg)
		return errors.Errorf(errMsg)
	}

	return nil
}

// Update 更新 k8s job
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

// Inspect 查看 k8s job 详细信息
func (k *k8sJob) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.New("job(k8s) not support inspect action")
}

// Cancel 停止 k8s job
func (k *k8sJob) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	name := strutil.Concat(job.Namespace, ".", job.Name)

	// 通过设置 job.spec.parallelism = 0 来停止 job
	return nil, k.setJobParallelism(defaultNamespace, name, 0)
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

	// 1核=1000m
	cpu := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.CPU*1000)), "m"))
	// 1Mi=1024K=1024x1024字节
	memory := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.Memory)), "Mi"))

	vols, volMounts, _ := GenerateK8SVolumes(&job)

	backofflimit := int32(job.BackoffLimit)
	kubeJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       jobKind,
			APIVersion: jobAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      strutil.Concat(job.Namespace, ".", job.Name),
			Namespace: job.Namespace,
			// TODO: 现在无法直接使用job.Labels，不符合k8s labels的规则
			//Labels:    job.Labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &defaultParallelism,
			// Completions = nil, 即表示只要有一次成功，就算completed
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
							ImagePullPolicy: corev1.PullAlways,
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
	// 按当前业务，只支持一个 Pod 一个 Container
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
			ImagePullPolicy: corev1.PullAlways,
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

	// 加上 K8S 标识
	env = append(env, corev1.EnvVar{
		Name:  "IS_K8S",
		Value: "true",
	})

	// 加上 namespace 标识
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

	// Add POD_NAME
	env = append(env, corev1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	// Add POD_UUID
	env = append(env, corev1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.uid",
			},
		},
	})

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

	// job controller 都还未处理该 Job
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
		// TODO: 如何判断 job 是被 stopped?
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
	var kubeJob batchv1.Job

	path := strutil.Concat(k.prefix, ns, "/jobs/", name)

	// GET /apis/batch/v1/namespaces/{namespace}/jobs/{name}
	resp, err := k.client.Get(k.addr).
		Path(path).
		Header("Content-Type", "application/json").
		Do().
		JSON(&kubeJob)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to inspect k8s job, name: %s", name)
	}

	if resp == nil {
		return nil, errors.Errorf("response is null")
	}

	if !resp.IsOK() {
		errMsg := fmt.Sprintf("failed to inspect k8s job, name: %s, statusCode: %d",
			name, resp.StatusCode())
		logrus.Errorf(errMsg)
		return nil, errors.Errorf(errMsg)
	}

	return &kubeJob, nil
}

func (k *k8sJob) updateK8SJob(job batchv1.Job) error {
	ns := job.Namespace
	name := job.Name
	path := strutil.Concat(k.prefix, ns, "/jobs/", name)

	// PUT /apis/batch/v1/namespaces/{namespace}/jobs/{name}
	resp, err := k.client.Put(k.addr).
		Path(path).
		Header("Content-Type", "application/json").
		JSONBody(job).
		Do().
		DiscardBody()
	if err != nil {
		return errors.Wrapf(err, "failed to update k8s job, name: %s", name)
	}

	if resp == nil {
		return errors.Errorf("response is null")
	}

	if !resp.IsOK() {
		errMsg := fmt.Sprintf("failed to update k8s job, name: %s, statusCode: %d",
			name, resp.StatusCode())
		logrus.Errorf(errMsg)
		return errors.Errorf(errMsg)
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

// GenerateK8SVolumes 根据 job 配置，生产 volume 相关配置
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
	if _, err := k.secret.Get(namespace, k8s.AliyunRegistry); err == nil {
		return nil
	}

	// 集群初始化的时候会在 default namespace 下创建一个拉镜像的 secret
	s, err := k.secret.Get("default", k8s.AliyunRegistry)
	if err != nil {
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

	if err := k.secret.Create(mysecret); err != nil {
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
	if err := k.namespace.Exists(jobvolume.Namespace); err != nil {
		if err == k8serror.ErrNotFound {
			if err := k.namespace.Create(jobvolume.Namespace, nil); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	sc := whichStorageClass(jobvolume.Type)
	id := fmt.Sprintf("%s-%s", jobvolume.Namespace, jobvolume.Name)
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: jobvolume.Namespace,
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
	if err := k.pvc.CreateIfNotExists(&pvc); err != nil {
		return "", err
	}
	return id, nil
}
func (*k8sJob) KillPod(podname string) error {
	return fmt.Errorf("not support for k8sJob")
}
