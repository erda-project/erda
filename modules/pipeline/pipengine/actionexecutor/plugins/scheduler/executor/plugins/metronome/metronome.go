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

package metronome

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

func init() {
	types.MustRegister(Kind, func(name types.Name, clusterName string, cluster apistructs.ClusterInfo) (types.TaskExecutor, error) {
		if cluster.SchedConfig == nil {
			return nil, errors.Errorf("missing option configs for kind : %s, cluster: %+V", Kind, cluster)
		}

		client := httpclient.New()
		addr := cluster.SchedConfig.MasterURL
		if addr == "" {
			return nil, errors.Errorf("missing connect address for %s, addr: %s", Kind, addr)
		}

		if cluster.SchedConfig.CACrt != "" {
			logrus.Infof("metronome executor(%s) addr for https: %s", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(cluster.SchedConfig.ClientCrt),
				[]byte(cluster.SchedConfig.ClientKey),
				[]byte(cluster.SchedConfig.CACrt)))
		}

		//TODO add basic auth method for metronome

		go logic.GetAndSetTokenAuth(client, string(name))

		return &Metronome{
			name:        name,
			clusterName: clusterName,
			addr:        addr,
			client:      client,
			cluster:     cluster,
			enableTag:   cluster.SchedConfig.EnableTag,
		}, nil
	})
}

func (c *Metronome) Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	jobFromUser, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	_, scheduleInfo, err := logic.GetScheduleInfo(c.cluster, string(c.name), string(Kind), jobFromUser)
	if err != nil {
		return nil, err
	}

	job := apistructs.Job{
		JobFromUser:  jobFromUser,
		ScheduleInfo: scheduleInfo,
	}

	mJob, err := c.generateMetronomeJob(&job)
	if err != nil {
		logrus.Errorf("generateMetronomeJob err: %v, task: %+v", err, task)
		return nil, err
	}

	var respMetronomeJob MetronomeJob

	if bs, e := json.Marshal(mJob); e == nil {
		logrus.Debugf("mjob json: %v", string(bs))
	} else {
		logrus.Errorf("marshal mjob error: %v", err)
		return nil, errors.Errorf("Create: marshal mjob error: %v", err)
	}

	// create that job
	var b bytes.Buffer
	resp, err := c.client.Post(c.addr).
		Path("/v1/jobs").
		Header("Content-Type", "application/json").
		JSONBody(mJob).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("post metronome job(%s) error: %v", mJob.Id, err)
	}

	if !resp.IsOK() {
		return nil, errors.Errorf("failed to create metronome job: %s, statusCode=%d, body: %v", mJob.Id, resp.StatusCode(), b.String())
	}

	r := bytes.NewReader(b.Bytes())
	if err := json.NewDecoder(r).Decode(&respMetronomeJob); err != nil {
		return nil, err
	}

	b.Reset()
	// post a run to run that job
	resp, err = c.client.Post(c.addr).
		Path("/v1/jobs/"+respMetronomeJob.Id+"/runs").
		Header("Content-Type", "application/json").
		JSONBody(nil).
		Do().
		Body(&b)

	if err != nil {
		return nil, errors.Errorf("post metronome job(%s) runs error: %v", mJob.Id, err)
	}

	if !resp.IsOK() {
		return nil, errors.Errorf("failed to create a run for metronome job: %s, statusCode=%d, body: %v", mJob.Id, resp.StatusCode(), b.String())
	}

	return job, nil
}

func (c *Metronome) Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	jobName := task.Extra.Namespace + "." + task.Name
	runs := make([]Run, 0)

	// first step, kill all runs for a job
	resp, err := c.client.Get(c.addr).
		Path("/v1/jobs/" + url.PathEscape(jobName) + "/runs").
		Do().JSON(&runs)
	if err != nil {
		return nil, errors.Wrapf(err, "metronome get runs for job(%s) failed, err: %v", jobName, err)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to get runs of job, name: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	for _, run := range runs {
		resp, err = c.client.Post(c.addr).
			Path("/v1/jobs/" + url.PathEscape(jobName) + "/runs/" + run.Id + "/actions/stop").
			Do().
			DiscardBody()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to stop metronome run for job, runID: %v, name: %s", run.Id, jobName)
		}
		if !resp.IsOK() {
			return nil, errors.Errorf("failed to stop metronome run(%s) for the job(%s), statusCode=%d", run.Id, jobName, resp.StatusCode())
		}
	}

	return c.removejob(ctx, task)
}

func (c *Metronome) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error) {
	jobStatus := apistructs.StatusDesc{Status: apistructs.StatusUnknown}
	jobName := task.Extra.Namespace + "." + task.Name

	var getJob MetronomeJobResult
	resp, err := c.client.Get(c.addr).
		Path("/v1/jobs/"+jobName).
		Param("embed", "activeRuns").
		Param("embed", "history").
		Do().
		JSON(&getJob)

	if err != nil {
		return jobStatus, errors.Wrapf(err, "metronome get job(%s) failed", jobName)
	}
	if !resp.IsOK() {
		if resp.StatusCode() == http.StatusNotFound {
			// 1. If the create interface comes in, it will ignore this state and go on
			// 2. The real get interface comes in, and the normal scene can be summarized into StatusStoppedByKilled state, that is, the job has been deleted
			jobStatus.Status = apistructs.StatusNotFoundInCluster
			logrus.Debugf("not found metronome job, name: %s", jobName)
			return jobStatus, nil
		}
		return jobStatus, errors.Errorf("failed to get metronome job(%s), statusCode=%d", jobName, resp.StatusCode())
	}

	// The current setting, there will only be one runtime in a job
	runs := make([]RunResult, 0)
	resp, err = c.client.Get(c.addr).
		Path("/v1/jobs/" + url.PathEscape(jobName) + "/runs").
		Do().
		JSON(&runs)
	// there is no active runs
	if len(runs) == 0 {
		if len(getJob.ActiveRuns) > 0 {
			jobStatus.Status = apistructs.StatusRunning
		} else if len(getJob.History.SuccessfulFinishedRuns) > 0 {
			jobStatus.Status = apistructs.StatusStoppedOnOK
		} else if len(getJob.History.FailedFinishedRuns) > 0 {
			jobStatus.Status = apistructs.StatusStoppedOnFailed
		} else {
			jobStatus.Status = apistructs.StatusUnknown
			jobStatus.LastMessage = string(apistructs.StatusUnknown)
			logrus.Warningf("get metronome job, body: %+v", getJob)
		}
	} else {
		run := runs[0]
		// The currently observed state includes
		// INITIAL, The resource is not in place, "INITIAL" is the status obtained from the api, and the corresponding display on the dcos page is "Starting"
		// ACTIVE, Running, "Running" is displayed on the corresponding dcos page
		// STARTING
		// SUCCESS
		// FAILED
		switch run.Status {
		case "INITIAL":
			jobStatus.Status = apistructs.StatusUnschedulable
		case "ACTIVE", "STARTING":
			jobStatus.Status = apistructs.StatusRunning
		case "SUCCESS":
			jobStatus.Status = apistructs.StatusStoppedOnOK
		case "FAILED":
			jobStatus.Status = apistructs.StatusStoppedOnFailed
		default:
			jobStatus.Status = apistructs.StatusUnknown
			logrus.Warningf("metronome job unknown status: %s, jobName: %s", run.Status, jobName)
		}
	}

	logrus.Debugf("metronome job(%s) status: %+v", jobName, jobStatus)
	return jobStatus, nil
}

func (c *Metronome) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = c.Remove(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (c *Metronome) removejob(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	jobName := task.Extra.Namespace + "." + task.Name

	// remove job
	resp, err := c.client.Delete(c.addr).
		Path("/v1/jobs/" + url.PathEscape(jobName)).
		Do().
		DiscardBody()
	if err != nil {
		return nil, errors.Wrapf(err, "metronome delete job: %s", jobName)
	}
	if !resp.IsOK() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.Errorf("failed to delete metronome job: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	return nil, nil
}

func (c *Metronome) generateMetronomeJob(job *apistructs.Job) (*MetronomeJob, error) {
	var placement *Placement

	constrains := constructMetronomeConstrains(&job.ScheduleInfo)

	if v := os.Getenv("FORCE_OLD_LABEL_SCHEDULE"); v == "true" {
		placement = c.buildMetronomePlacement(job.Labels)
	} else {
		if constrains == nil {
			placement = nil
		} else {
			placement = constrains2Placement(constrains)
		}
	}
	envs := job.Env
	envs["DICE_CPU_ORIGIN"] = fmt.Sprintf("%f", job.CPU)
	envs["DICE_CPU_REQUEST"] = fmt.Sprintf("%f", job.CPU)
	envs["DICE_CPU_LIMIT"] = fmt.Sprintf("%f", job.CPU)
	envs["DICE_MEM_ORIGIN"] = fmt.Sprintf("%f", job.Memory)
	envs["DICE_MEM_REQUEST"] = fmt.Sprintf("%f", job.Memory)
	envs["DICE_MEM_LIMIT"] = fmt.Sprintf("%f", job.Memory)
	envs["IS_K8S"] = "false"

	ciEnvs, err := logic.GetCLusterInfo(c.clusterName)
	if err != nil {
		return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)",
			c.clusterName, err)
	}

	mJob := &MetronomeJob{
		Id: job.Namespace + "." + job.Name,
		// TODO: following undefied fields(Description, Label)
		Description: "hello-metronome",
		Labels:      job.Labels,
		Run: Run{
			Artifacts: make([]Artifact, 0),
			Cmd:       job.Cmd,
			Cpus:      job.CPU,
			Mem:       job.Memory,
			Env:       job.Env,
			Restart: Restart{
				Policy: "NEVER",
			},
			Docker: Docker{
				Image: job.Image,
				// TODO: following undefied fields
				ForcePullImage: true,
				//Parameters:     parameters,
			},
			Volumes: make([]Volume, 0),
			// TODO: following undefied fields
			MaxLaunchDelay: 3600,
			Disk:           0,
			Placement:      placement,
		},
	}

	for k, v := range ciEnvs {
		mJob.Run.Env[string(k)] = v
	}

	for _, bind := range job.Binds {
		var mode string

		if bind.ReadOnly {
			mode = "RO"
		} else {
			mode = "RW"
		}

		hostPath, err := logic.ParseJobHostBindTemplate(bind.HostPath, ciEnvs)
		if err != nil {
			return nil, err
		}

		mJob.Run.Volumes = append(mJob.Run.Volumes,
			Volume{
				ContainerPath: bind.ContainerPath,
				HostPath:      hostPath,
				Mode:          mode,
			})
	}

	for i, vol := range job.Volumes {
		volid := uuid.Generate()
		if vol.ID != nil {
			volid = *vol.ID
		}

		mp, ok := ciEnvs["DICE_STORAGE_MOUNTPOINT"]
		if !ok {
			return nil, errors.New("not found DICE_STORAGE_MOUNTPOINT from clusterInfo")
		}

		pipelineHostPath := strutil.JoinPath(mp, "/devops/ci/pipelines", volid)
		mJob.Run.Volumes = append(mJob.Run.Volumes, Volume{
			ContainerPath: vol.Path,
			HostPath:      pipelineHostPath,
			Mode:          "RW",
		})
		job.Volumes[i].ID = &volid
	}

	// pre fetch from host path
	if job.PreFetcher != nil && job.PreFetcher.FileFromHost != "" {
		hostPath, err := logic.ParseJobHostBindTemplate(job.PreFetcher.FileFromHost, ciEnvs)
		if err != nil {
			return nil, err
		}
		mJob.Run.Volumes = append(mJob.Run.Volumes, Volume{
			ContainerPath: job.PreFetcher.ContainerPath,
			HostPath:      hostPath,
			Mode:          "RO",
		})
	}

	// contruct artifact
	if artifacts := generateFetcherFromEnv(job.Env); len(artifacts) != 0 {
		mJob.Run.Artifacts = artifacts
	}

	return mJob, nil
}

func (c *Metronome) buildMetronomePlacement(labels map[string]string) *Placement {
	dcosCons := logic.BuildDcosConstraints(c.enableTag, labels, nil, nil)
	return constrains2Placement(dcosCons)
}

func generateFetcherFromEnv(env map[string]string) []Artifact {
	artifacts := []Artifact{}
	for key, value := range env {
		if strings.HasPrefix(key, "MESOS_FETCHER_URI") {
			artifact := Artifact{
				Uri:        value,
				Executable: false,
				Extract:    false,
				Cache:      false,
			}
			artifacts = append(artifacts, artifact)
		}
	}
	return artifacts
}

func constrains2Placement(constrains [][]string) *Placement {
	if constrains == nil || len(constrains) == 0 {
		return nil
	}
	var metroCons []Constraints
	for _, one := range constrains {
		if len(one) != 3 {
			continue
		}
		metroCons = append(metroCons, Constraints{Attribute: one[0], Operator: one[1], Value: one[2]})
	}
	if len(metroCons) == 0 {
		return nil
	}
	return &Placement{metroCons}
}

// TODO: This function needs to be refactored
func constructMetronomeConstrains(r *apistructs.ScheduleInfo) [][]string {
	var constrains [][]string
	if r.IsPlatform {
		constrains = append(constrains,
			[]string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + apistructs.TagPlatform + `\b.*`})
		if r.IsUnLocked {
			constrains = append(constrains,
				[]string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + apistructs.TagLocked + `\b.*`})
		}
	}

	// Do not schedule to the node marked with the prefix of the label
	for _, unlikePrefix := range r.UnLikePrefixs {
		constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + unlikePrefix + `[^,]+\b.*`})
	}
	// Not scheduled to the node with this label
	unlikes := []string{}
	copy(unlikes, r.UnLikes)
	if !r.IsPlatform {
		unlikes = append(unlikes, apistructs.TagPlatform)
	}
	if r.IsUnLocked {
		unlikes = append(unlikes, apistructs.TagLocked)
	}
	for _, unlike := range unlikes {
		constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "UNLIKE", `.*\b` + unlike + `\b.*`})
	}
	// Specify scheduling to the node labeled with the prefix
	// Currently no such label
	for _, likePrefix := range r.LikePrefixs {
		constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + likePrefix + `\b.*`})
	}
	// Specify to be scheduled to the node with this label, not coexisting with any
	for _, exclusiveLike := range r.ExclusiveLikes {
		constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + exclusiveLike + `\b.*`})
	}
	// Specify to be scheduled to the node with this label, if the any label is enabled, the any label is attached
	for _, like := range r.Likes {
		if r.Flag {
			constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + apistructs.TagAny + `\b.*|.*\b` + like + `\b.*`})
		} else {
			constrains = append(constrains, []string{labelconfig.DCOS_ATTRIBUTE, "LIKE", `.*\b` + like + `\b.*`})
		}
	}

	// Specify scheduling to the node with this label, allowing multiple OR operations
	if len(r.InclusiveLikes) > 0 {
		constrain := []string{labelconfig.DCOS_ATTRIBUTE, "LIKE"}
		var sentence string
		for i, inclusiveLike := range r.InclusiveLikes {
			if i == len(r.InclusiveLikes)-1 {
				sentence = sentence + `.*\b` + inclusiveLike + `\b.*`
				constrain = append(constrain, sentence)
				constrains = append(constrains, constrain)
			} else {
				sentence = sentence + `.*\b` + inclusiveLike + `\b.*|`
			}
		}
	}

	if len(r.SpecificHost) != 0 {
		constrains = append(constrains, []string{"hostname", "LIKE", strings.Join(r.SpecificHost, "|")})
	}

	return constrains
}
