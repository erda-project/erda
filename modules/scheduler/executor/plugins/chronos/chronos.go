package chronos

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/httpclient"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	kind               = "CHRONOS"
	defaultNetwork     = "USER"
	defaultNetworkName = "dcos"
	defaultSchedule    = "R1//PT1M"
)

// chronos plugin's configure
//
// EXECUTOR_CHRONOS_CHRONOSFORJOB_ADDR="{host}:{port}"
// EXECUTOR_CHRONOS_CHRONOSFORJOB_NETWORK="USER"
// EXECUTOR_CHRONOS_CHRONOSFORJOB_ENABLETAG=true
//
func init() {
	executortypes.Register(kind, func(name executortypes.Name, clustername string, options map[string]string, optionsPlus interface{}) (executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found chronos address in env variables")
		}

		// TODO: check http scheme

		network, ok := options["NETWORK"]
		if !ok {
			network = defaultNetwork
		}
		enableTag, err := util.ParseEnableTagOption(options, "ENABLETAG", false)
		if err != nil {
			return nil, err
		}
		return &Chronos{
			name:      name,
			options:   options,
			addr:      addr,
			network:   network,
			client:    httpclient.New(),
			enableTag: enableTag,
		}, nil
	})
}

type Chronos struct {
	name      executortypes.Name
	options   map[string]string
	addr      string
	network   string
	client    *httpclient.HTTPClient
	enableTag bool
}

func (c *Chronos) Kind() executortypes.Kind {
	return kind
}

func (c *Chronos) Name() executortypes.Name {
	return c.name
}

func (c *Chronos) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	var envs []NVField

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	for k, v := range job.Env {
		envs = append(envs, NVField{Name: k, Value: v})
	}

	cJob := Job{
		// regexp: /^[w.-]+$/
		Name:                 job.Namespace + "." + job.Name,
		Command:              job.Cmd,
		Shell:                true,
		Cpus:                 job.CPU,
		Mem:                  job.Memory,
		Description:          job.Namespace + "/" + job.Name,
		Retries:              0,
		Schedule:             defaultSchedule,
		Constraints:          util.BuildDcosConstraints(c.enableTag, job.Labels, nil, nil),
		EnvironmentVariables: envs,
		Async:                false,
		Disabled:             false,
		Container: &JobContainer{
			Image:   job.Image,
			Type:    "DOCKER",
			Network: c.network,
		},
		ScheduleTimeZone: "UTC",
		// If Chronos misses the scheduled run time for any reason,
		// it will still run the job if the time is within this interval.
		Epsilon: "PT2H",
		//Schedule:         scheduleTime(0, time.Now().UTC().Unix(), int64(12*time.Hour)),
	}

	// container network name
	if cJob.Container.Network == defaultNetwork {
		cJob.Container.NetworkName = defaultNetworkName
	}

	// container volumes
	for _, bind := range job.Binds {
		var mode string

		if bind.ReadOnly {
			mode = "RO"
		} else {
			mode = "RW"
		}
		cJob.Container.Volumes = append(cJob.Container.Volumes,
			JobContainerVolume{
				ContainerPath: bind.ContainerPath,
				HostPath:      bind.HostPath,
				Mode:          mode,
			})
	}

	resp, err := c.client.Post(c.addr).
		Path("/v1/scheduler/iso8601").
		Header("Content-Type", "application/json").
		JSONBody(cJob).Do().DiscardBody()
	if err != nil {
		return nil, errors.Wrapf(err, "chronos schedule job: %s", cJob.Name)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to create chronos job: %s, statusCode=%d", cJob.Name, resp.StatusCode())
	}

	return nil, nil
}

func (c *Chronos) Destroy(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	jobName := job.Namespace + "." + job.Name

	// kill all tasks for a job
	resp, err := httpclient.New().Delete(c.addr).
		Path("/v1/scheduler/task/kill/" + url.PathEscape(jobName)).
		Do().DiscardBody()
	if err != nil {
		return errors.Wrapf(err, "chronos kill all tasks by job: %s", jobName)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to kill all chronos task by job: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	// remove job
	resp, err = c.client.Delete(c.addr).
		Path("/v1/scheduler/job/" + url.PathEscape(jobName)).
		Do().DiscardBody()
	if err != nil {
		return errors.Wrapf(err, "chronos delete job: %s", jobName)
	}
	if !resp.IsOK() {
		return errors.Errorf("failed to delete chronos job: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	return nil
}

func (c *Chronos) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	var (
		jobStatus apistructs.StatusDesc
		summary   jobSummary
	)

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return jobStatus, errors.New("invalid job spec")
	}

	jobName := job.Namespace + "." + job.Name

	resp, err := c.client.Get(c.addr).Path("/v1/scheduler/jobs/summary").
		Do().JSON(&summary)
	if err != nil {
		return jobStatus, errors.Wrapf(err, "chronos get status: %s", jobName)
	}
	if !resp.IsOK() {
		return jobStatus, errors.Errorf("failed to get chronos jobs summary: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	jobStatus.Status = getChronosJobStatus(jobName, summary)
	return jobStatus, nil
}

func (c *Chronos) Remove(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	jobName := job.Namespace + "." + job.Name

	// remove job
	resp, err := c.client.Delete(c.addr).
		Path("/v1/scheduler/job/" + url.PathEscape(jobName)).
		Do().DiscardBody()
	if err != nil || !resp.IsOK() {
		logrus.Warnf("failed to remove chronos job: %s, statusCode=%d", jobName, resp.StatusCode())
	}

	return nil
}

func (c *Chronos) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.New("job(chronos) not support update action")
}

func (c *Chronos) Inspect(ctx context.Context, spec interface{}) (interface{}, error) {
	return nil, errors.New("job(chronos) not support inspect action")
}

func (c *Chronos) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in Chronos")
}

func makeChronosJobKey(namespace, name string) string {
	return "/dice/plugins/chronos/job/" + namespace + "/" + name
}

func getChronosJobStatus(jName string, sum jobSummary) apistructs.StatusCode {
	var status apistructs.StatusCode = apistructs.StatusStoppedByKilled

	for _, record := range sum.Jobs {
		if record.Name == jName {
			logrus.Debugf("chronos job status: %v", record)

			switch record.Status {
			case "success":
				status = apistructs.StatusStoppedOnOK
			case "failure":
				status = apistructs.StatusStoppedOnFailed
			case "fresh":
				// state是idle的情况下，需要判断是否已经调度
				// 没有调度的话：StatusUnschedulable
				// 已经调度过的话：StatusStoppedByKilled
				if strings.Contains(record.State, "running") {
					status = apistructs.StatusRunning
				} else if record.State == "queued" {
					status = apistructs.StatusUnschedulable
				} else if record.State == "idle" {
					if record.Schedule == defaultSchedule {
						status = apistructs.StatusUnschedulable
					} else {
						status = apistructs.StatusStoppedByKilled
					}
				} else {
					status = apistructs.StatusUnknown
				}
			default:
				status = apistructs.StatusUnknown
			}
			break
		}
	}

	return status
}

func (c *Chronos) Cancel(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.Errorf("job(metronome) not support Cancel action")
}
func (c *Chronos) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}
func (c *Chronos) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
}

func (c *Chronos) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for chronos")
}

func (k *Chronos) CleanUpBeforeDelete() {}
func (k *Chronos) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for chronos")
}

func (*Chronos) KillPod(podname string) error {
	return fmt.Errorf("not support for chronos")
}
