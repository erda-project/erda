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
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	kind = "METRONOME"
)

func init() {
	executortypes.Register(kind, func(name executortypes.Name, clusterName string, options map[string]string, optionsPlus interface{}) (executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found metronome address in env variables")
		}

		client := httpclient.New()

		if _, ok := options["CA_CRT"]; ok {
			logrus.Infof("metronome executor(%s) addr for https: %v", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
				[]byte(options["CLIENT_KEY"]),
				[]byte(options["CA_CRT"])))
		}

		if basicAuth, ok := options["BASICAUTH"]; ok {
			ba := strings.Split(basicAuth, ":")
			if len(ba) == 2 {
				client = client.BasicAuth(ba[0], ba[1])
			}
		}

		go util.GetAndSetTokenAuth(client, string(name))

		enableTag, err := util.ParseEnableTagOption(options, "ENABLETAG", false)
		if err != nil {
			return nil, err
		}

		// cluster info
		js, err := jsonstore.New()
		if err != nil {
			return nil, errors.Errorf("failed to new json store for clusterinfo, executor: %s, (%v)",
				name, err)
		}
		ci := clusterinfo.NewClusterInfoImpl(js)

		return &Metronome{
			name:        name,
			clusterName: clusterName,
			options:     options,
			addr:        addr,
			client:      client,
			enableTag:   enableTag,
			clusterInfo: ci,
		}, nil
	})
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

type Metronome struct {
	name        executortypes.Name
	clusterName string
	options     map[string]string
	addr        string
	client      *httpclient.HTTPClient
	enableTag   bool
	clusterInfo clusterinfo.ClusterInfo
}

func (c *Metronome) Kind() executortypes.Kind {
	return kind
}

func (c *Metronome) Name() executortypes.Name {
	return c.name
}

func (c *Metronome) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	mJob, err := c.generateMetronomeJob(&job)
	if err != nil {
		logrus.Errorf("generateMetronomeJob error: %v, %+v", err, specObj)
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

func (c *Metronome) Destroy(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	if job.Name == "" {
		return errors.New("metronome not support delete pipeline jobs")
	}

	jobName := job.Namespace + "." + job.Name
	runs := make([]Run, 0)

	// first step, kill all runs for a job
	resp, err := c.client.Get(c.addr).
		Path("/v1/jobs/" + url.PathEscape(jobName) + "/runs").
		Do().JSON(&runs)
	if err != nil {
		return errors.Wrapf(err, "metronome get runs for job(%s) failed, err: %v", jobName, err)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to get runs of job, name: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	for _, run := range runs {
		resp, err = c.client.Post(c.addr).
			Path("/v1/jobs/" + url.PathEscape(jobName) + "/runs/" + run.Id + "/actions/stop").
			Do().
			DiscardBody()
		if err != nil {
			return errors.Wrapf(err, "failed to stop metronome run for job, runID: %v, name: %s", run.Id, jobName)
		}
		if !resp.IsOK() {
			return errors.Errorf("failed to stop metronome run(%s) for the job(%s), statusCode=%d", run.Id, jobName, resp.StatusCode())
		}
	}

	// second step, delete that job
	return c.removeJob(ctx, specObj)
}

func (c *Metronome) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	jobStatus := apistructs.StatusDesc{Status: apistructs.StatusUnknown}

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return jobStatus, errors.New("invalid job spec")
	}

	jobName := job.Namespace + "." + job.Name
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

// Attention: DO NOT delete Metronome job that had active runs in it,
// if DELETE a metronome with active runs, there would be a 409 Error, with following info:
// "message":"There are active job runs. Override with stopCurrentJobRuns=true",
// Use DESTROY method instead
func (c *Metronome) Remove(ctx context.Context, specObj interface{}) error {
	return c.Destroy(ctx, specObj)
}
func (c *Metronome) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

func (c *Metronome) removeJob(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	jobName := job.Namespace + "." + job.Name

	// remove job
	resp, err := c.client.Delete(c.addr).
		Path("/v1/jobs/" + url.PathEscape(jobName)).
		Do().
		DiscardBody()
	if err != nil {
		return errors.Wrapf(err, "metronome delete job: %s", jobName)
	}
	if !resp.IsOK() {
		if resp.StatusCode() == http.StatusNotFound {
			// 按照调度层的接口，如果一create就来delete是有问题的，
			// 因为create只会在etcd存一下jobid, metronome看来job就是不存在的
			// 其他的情况也可认为删除成功
			return nil
		}
		return errors.Errorf("failed to delete metronome job: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	return nil
}

func (c *Metronome) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.New("job(metronome) not support update action")
}

func (c *Metronome) Inspect(ctx context.Context, spec interface{}) (interface{}, error) {
	return nil, errors.New("job(metronome) not support inspect action")
}

func (c *Metronome) Cancel(ctx context.Context, spec interface{}) (interface{}, error) {
	return nil, errors.New("job(metronome) not support Cancel action")
}
func (m *Metronome) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

func (c *Metronome) generateMetronomeJob(job *apistructs.Job) (*MetronomeJob, error) {
	var placement *Placement

	constrains := constructMetronomeConstrains(&job.ScheduleInfo)

	// emergency exit
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

	// get cluster info
	ciEnvs, err := c.clusterInfo.Info(c.clusterName)
	if err != nil {
		return nil, errors.Errorf("failed to get cluster info, clusterName: %s, (%v)",
			c.clusterName, err)
	}

	//parameters := setDockerLabelParameters(job.Labels)
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

	// container volumes
	for _, bind := range job.Binds {
		var mode string

		if bind.ReadOnly {
			mode = "RO"
		} else {
			mode = "RW"
		}

		hostPath, err := clusterinfo.ParseJobHostBindTemplate(bind.HostPath, ciEnvs)
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
		hostPath, err := clusterinfo.ParseJobHostBindTemplate(job.PreFetcher.FileFromHost, ciEnvs)
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
	dcosCons := util.BuildDcosConstraints(c.enableTag, labels, nil, nil)
	return constrains2Placement(dcosCons)
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

func setDockerLabelParameters(labels map[string]string) []DockerParameter {
	parameters := make([]DockerParameter, 0)
	for k, v := range labels {
		// The label directly inserted into the metronome is the mesos label, which is not related to the docker label
		// The way to pass to docker label is detailed in
		// https://jira.mesosphere.com/browse/MARATHON-4738
		// https://issues.apache.org/jira/browse/MESOS-4446
		parameters = append(parameters, DockerParameter{"label", k + "=" + v})
	}
	return parameters
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

func (m *Metronome) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in Metronome")
}
func (m *Metronome) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for metronome")
}
func (*Metronome) CleanUpBeforeDelete() {}
func (*Metronome) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for metronome")
}

func (*Metronome) KillPod(podname string) error {
	return fmt.Errorf("not support for metronome")
}

func (*Metronome) Scale(ctx context.Context, spec interface{}) (interface{}, error) {
	return apistructs.ServiceGroup{}, fmt.Errorf("scale not support for metronome")
}
