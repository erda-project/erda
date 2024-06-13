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

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/actionagent"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/actionexecutor/logic"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

type ContainerStatus string

// docker container status
var (
	ContainerStatusRunning    ContainerStatus = "running"
	ContainerStatusRestarting ContainerStatus = "restarting"
	ContainerStatusCreated    ContainerStatus = "created"
	ContainerStatusExited     ContainerStatus = "exited"
	ContainerStatusPaused     ContainerStatus = "paused"
	ContainerStatusRemoving   ContainerStatus = "removing"
	ContainerStatusDead       ContainerStatus = "dead"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindDocker)

func init() {
	types.MustRegister(Kind, func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		return New(name)
	})
}

// TODO: create docker client by executor-enviroment
// the remote docker client should be created by host, tls-veiry like client.WithHost("tcp://xxx:2375")
// the local docker  client could be created by docker-sock env directly like clinet.FromEnv
func New(name types.Name) (*DockerJob, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create docker client")
	}
	return &DockerJob{
		name:       name,
		client:     cli,
		errWrapper: logic.NewErrorWrapper(name.String()),
	}, nil
}

type DockerJob struct {
	client     *client.Client
	name       types.Name
	errWrapper *logic.ErrorWrapper
}

func (d *DockerJob) Kind() types.Kind {
	return Kind
}

func (d *DockerJob) Name() types.Name {
	return d.name
}

func (d *DockerJob) Status(ctx context.Context, task *spec.PipelineTask) (desc apistructs.PipelineStatusDesc, err error) {
	if err := logic.ValidateAction(task); err != nil {
		return desc, err
	}
	jobName := logic.MakeJobName(task)
	container, err := d.client.ContainerInspect(ctx, jobName)
	if err != nil {
		return desc, err
	}
	switch container.State.Status {
	case string(ContainerStatusRunning), string(ContainerStatusRestarting), string(ContainerStatusCreated):
		desc.Status = apistructs.PipelineStatusRunning
	case string(ContainerStatusExited):
		desc.Status = apistructs.PipelineStatusFailed
		if container.State.ExitCode == 0 {
			desc.Status = apistructs.PipelineStatusSuccess
		}
	case string(ContainerStatusPaused):
		desc.Status = apistructs.PipelineStatusPaused
	case string(ContainerStatusRemoving):
		desc.Status = apistructs.PipelineStatusStopByUser
	case string(ContainerStatusDead):
		desc.Status = apistructs.PipelineStatusFailed
	default:
		desc.Status = apistructs.PipelineStatusUnknown
	}
	return desc, nil
}

func (d *DockerJob) Exist(ctx context.Context, task *spec.PipelineTask) (created, started bool, err error) {
	statusDesc, err := d.Status(ctx, task)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, false, nil
		}
		return
	}
	created = true
	switch statusDesc.Status {
	case apistructs.PipelineStatusUnknown:
		err = errors.Errorf("failed to judge job exist or not, detail: %v", statusDesc)
	case apistructs.PipelineStatusRunning, apistructs.PipelineStatusSuccess, apistructs.PipelineStatusPaused,
		apistructs.PipelineStatusStopByUser, apistructs.PipelineStatusFailed:
		started = true
	default:
		started = true
	}
	return
}

func (d *DockerJob) Create(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer d.errWrapper.WrapTaskError(&err, "create job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	created, _, err := d.Exist(ctx, task)
	if err != nil {
		return nil, err
	}
	if created {
		logrus.Warnf("%s: task already created, taskInfo: %s", d.Kind().String(), logic.PrintTaskInfo(task))
	}
	return nil, nil
}

func (d *DockerJob) Start(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer d.errWrapper.WrapTaskError(&err, "start job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	created, started, err := d.Exist(ctx, task)
	if err != nil {
		return nil, err
	}
	if !created {
		logrus.Warnf("%s: task not created(auto try to create), taskInfo: %s", d.Kind().String(), logic.PrintTaskInfo(task))
		_, err = d.Create(ctx, task)
		if err != nil {
			return nil, err
		}
		logrus.Warnf("dockerjob: action created, continue to start, taskInfo: %s", logic.PrintTaskInfo(task))
	}
	if started {
		logrus.Warnf("%s: task already started, taskInfo: %s", d.Kind().String(), logic.PrintTaskInfo(task))
		return nil, nil
	}
	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}
	if len(job.Volumes) > 0 {
		mounts := d.GenerateDockerVolumes(&job)
		for _, mt := range mounts {
			_, err := d.client.VolumeCreate(ctx, volume.CreateOptions{
				Name:   mt.Source,
				Driver: "local",
			})
			if err != nil {
				return nil, err
			}
		}

		for i := range mounts {
			job.Volumes[i].ID = &(mounts[i].Source)
		}
	}
	if job.PreFetcher != nil {
		preFetcherJob, err := d.generatePreFetcherContainer(ctx, &job)
		if err != nil {
			return nil, errors.Errorf("failed to created prefetcher container: %v", err)
		}
		if err := d.client.ContainerStart(ctx, preFetcherJob.ID, dockertypes.ContainerStartOptions{}); err != nil {
			return nil, errors.Errorf("failed to start prefetcher container: %v", err)
		}
	}
	jobContainer, err := d.createContainerByJob(ctx, &job)
	if err != nil {
		return nil, err
	}
	if err := d.client.ContainerStart(ctx, jobContainer.ID, dockertypes.ContainerStartOptions{}); err != nil {
		return nil, errors.Errorf("failed to start job container, err: %v", err)
	}
	return apistructs.Job{
		JobFromUser: job,
	}, nil
}

func (d *DockerJob) Update(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, errors.Errorf("%s not support update operation", d.Kind().String())
}

func (d *DockerJob) Cancel(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer d.errWrapper.WrapTaskError(&err, "cancel job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	oldUUID := task.Extra.UUID
	task.Extra.UUID = task_uuid.MakeJobID(task)
	data, err = d.Delete(ctx, task)
	task.Extra.UUID = oldUUID
	return data, err
}

func (d *DockerJob) Remove(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer d.errWrapper.WrapTaskError(&err, "remove job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	task.Extra.UUID = task_uuid.MakeJobID(task)
	return d.Delete(ctx, task)
}

func (d *DockerJob) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	task := tasks[0]
	defer d.errWrapper.WrapTaskError(&err, "batch delete job", task)
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = d.Delete(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (d *DockerJob) Delete(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}
	name := d.makeJobName(task.Extra.Namespace, task.Extra.UUID)
	jobContainer, err := d.client.ContainerInspect(ctx, name)
	if err != nil {
		if client.IsErrNotFound(err) {
			logrus.Warnf("%s: task not exist, taskInfo: %s", d.Kind().String(), logic.PrintTaskInfo(task))
			return nil, nil
		}
		return nil, err
	}
	if err := d.client.ContainerRemove(ctx, jobContainer.ID, dockertypes.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return nil, err
	}
	for _, mountVol := range jobContainer.Mounts {
		err = d.client.VolumeRemove(ctx, mountVol.Name, true)
		if err != nil {
			if !client.IsErrNotFound(err) {
				return nil, errors.Errorf("failed to remove docker volume, name: %s", mountVol.Name)
			}
			logrus.Warningf("the docker job %s's volume %s in namespace %s is not found", name, mountVol.Name, job.Namespace)
		}
	}
	return task.Extra.UUID, nil
}

func (d *DockerJob) Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error) {
	jobName := logic.MakeJobName(task)
	inspect, err := d.client.ContainerInspect(ctx, jobName)
	if err != nil {
		return apistructs.TaskInspect{}, err
	}
	inspectDesc, err := json.Marshal(inspect)
	if err != nil {
		return apistructs.TaskInspect{}, err
	}
	return apistructs.TaskInspect{Desc: string(inspectDesc)}, nil
}

func (d *DockerJob) createContainerByJob(ctx context.Context, job *apistructs.JobFromUser) (container.CreateResponse, error) {
	jobImage, err := d.client.ImagePull(ctx, job.Image, dockertypes.ImagePullOptions{})
	if err != nil {
		return container.CreateResponse{}, errors.Errorf("failed to pull image, err: %v", err)
	}
	defer jobImage.Close()
	cpu := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.CPU*1000)), "m"))
	memory := resource.MustParse(strutil.Concat(strconv.Itoa(int(job.Memory)), "Mi"))

	cfg := container.Config{
		Image: job.Image,
		Env:   d.generateContainerEnvs(job),
	}
	if job.Cmd != "" {
		cfg.Cmd = append(cfg.Cmd, []string{"sh", "-c", job.Cmd}...)
	}
	memoryInt64, ok := memory.AsInt64()
	if !ok {
		return container.CreateResponse{}, errors.Errorf("failed to convert memory to int64")
	}
	hostCfg := container.HostConfig{
		Resources: container.Resources{
			Memory:     memoryInt64,
			MemorySwap: memoryInt64,
			NanoCPUs:   cpu.MilliValue(),
		},
		Privileged: true,
	}
	volMounts := d.GenerateDockerVolumes(job)
	for _, mt := range volMounts {
		hostCfg.Mounts = append(hostCfg.Mounts, mount.Mount{Type: mt.Type, Source: mt.Source, Target: mt.Target})
	}
	if job.PreFetcher != nil {
		hostCfg.Mounts = append(hostCfg.Mounts, mount.Mount{Type: mount.TypeVolume, Source: d.makePreFetcherVolumeName(job), Target: job.PreFetcher.ContainerPath})
	}
	for _, bind := range job.Binds {
		hostCfg.Mounts = append(hostCfg.Mounts, mount.Mount{Type: mount.TypeBind, Source: bind.HostPath, Target: bind.ContainerPath, ReadOnly: bind.ReadOnly})
	}
	con, err := d.client.ContainerCreate(ctx, &cfg, &hostCfg, nil, nil, d.makeJobName(job.Namespace, job.Name))
	if err != nil {
		return container.CreateResponse{}, err
	}
	return con, nil
}

func (d *DockerJob) generatePreFetcherContainer(ctx context.Context, job *apistructs.JobFromUser) (container.CreateResponse, error) {
	jobImage, err := d.client.ImagePull(ctx, job.PreFetcher.FileFromImage, dockertypes.ImagePullOptions{})
	if err != nil {
		return container.CreateResponse{}, errors.Errorf("failed to pull image, err: %v", err)
	}
	defer jobImage.Close()
	preFetcherVolume, err := d.client.VolumeCreate(context.Background(), volume.CreateOptions{
		Name:   d.makePreFetcherVolumeName(job),
		Driver: "local",
	})
	if err != nil {
		return container.CreateResponse{}, err
	}
	cfg := container.Config{
		Image: job.PreFetcher.FileFromImage,
		Env:   d.generateContainerEnvs(job),
	}
	hostCfg := container.HostConfig{
		AutoRemove: true,
		Privileged: true,
		Mounts:     []mount.Mount{{Type: mount.TypeVolume, Source: preFetcherVolume.Name, Target: job.PreFetcher.ContainerPath}},
	}
	preFetcherContainer, err := d.client.ContainerCreate(context.Background(), &cfg, &hostCfg, nil, nil, d.makePreFetcherJobName(job))
	if err != nil {
		return container.CreateResponse{}, err
	}
	return preFetcherContainer, nil
}

func (d *DockerJob) generateContainerEnvs(job *apistructs.JobFromUser) []string {
	env := make([]string, 0)
	envMap := job.Env

	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// add docker label
	env = append(env, fmt.Sprintf("%s=true", apistructs.JobEnvIsDocker))

	// add namespace label
	env = append(env, fmt.Sprintf("%s=%s", apistructs.JobEnvNamespace, job.Namespace))
	env = append(env,
		fmt.Sprintf("%s=%f", apistructs.JobEnvOriginCPU, job.CPU),
		fmt.Sprintf("%s=%f", apistructs.JobEnvOriginMEM, job.Memory),
		fmt.Sprintf("%s=%f", apistructs.JobEnvRequestCPU, job.CPU),
		fmt.Sprintf("%s=%f", apistructs.JobEnvRequestMEM, job.Memory),
		fmt.Sprintf("%s=%f", apistructs.JobEnvLimitCPU, job.CPU),
		fmt.Sprintf("%s=%f", apistructs.JobENvLimitMEM, job.Memory),
	)

	// add container TerminusDefineTag env
	if len(job.TaskContainers) > 0 {
		env = append(env, fmt.Sprintf("%s=%s", apistructs.TerminusDefineTag, job.TaskContainers[0].ContainerID))
	}

	// todo: move this to executor enviroment, after impl the environment provider
	// active upload logs
	env = append(env, fmt.Sprintf("%s=true", actionagent.EnvEnablePushLog2Collector))

	return env
}

func (d *DockerJob) GenerateDockerVolumes(job *apistructs.JobFromUser) []mount.Mount {
	mounts := make([]mount.Mount, 0)
	for i, v := range job.Volumes {
		var volID string
		if v.ID == nil {
			volID = fmt.Sprintf("%s-%s-%d", job.Namespace, job.Name, i)
		} else {
			volID = *v.ID
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volID,
			Target: v.Path,
		})
	}
	return mounts
}

func (d *DockerJob) makePreFetcherJobName(job *apistructs.JobFromUser) string {
	return strutil.Concat(job.Namespace, ".", job.Name, "-prefetcher")
}

func (d *DockerJob) makePreFetcherVolumeName(job *apistructs.JobFromUser) string {
	return strutil.Concat(job.Namespace, ".", job.Name, "-prefetcher-volume")
}

func (d *DockerJob) makeJobName(namespace string, taskUUID string) string {
	return strutil.Concat(namespace, ".", taskUUID)
}
