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

package spark

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
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

		//TODO get spark version and executor image

		return &Spark{
			name:      name,
			addr:      addr,
			client:    client,
			enableTag: cluster.SchedConfig.EnableTag,
			cluster:   cluster,
		}, nil
	})
}

func (s *Spark) Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	job, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	_, scheduleInfo, err := logic.GetScheduleInfo(s.cluster, string(s.name), string(Kind), job)
	if err != nil {
		return nil, err
	}

	sparkProperties := make(map[string]string)
	sparkProperties["spark.app.name"] = job.Name
	sparkProperties["spark.driver.supervise"] = "true"
	sparkProperties["spark.eventLog.enabled"] = "false"
	sparkProperties["spark.submit.deployMode"] = "cluster"
	sparkProperties["spark.mesos.executor.docker.image"] = s.executorImage
	constrains := constructSparkConstraints(&scheduleInfo)
	if constrains != "dice_tags" { // When the user does not pass the label, the constraints are dice_tags, and the spark task does not need to be restricted, otherwise it will cause spark to exit
		// Measured, the tag configuration takes effect in spark.mesos.driver.constraints instead of spark.mesos.constraints
		sparkProperties["spark.mesos.driver.constraints"] = constrains
	}

	if job.CPU > 0 {
		sparkProperties["spark.driver.cores"] = fmt.Sprintf("%d", int64(job.CPU))
	}
	if job.Memory > 0 {
		sparkProperties["spark.driver.memory"] = fmt.Sprintf("%d", int64(job.Memory)) + "m"
		sparkProperties["spark.executor.memory"] = fmt.Sprintf("%d", int64(job.Memory)) + "m"
	}

	// Set the necessary environment variables
	if len(job.Env) == 0 {
		job.Env = make(map[string]string)
	}
	job.Env["SPARK_ENV_LOADED"] = "1"

	sparkRequest := &SparkCreateRequest{
		AppResource:          job.Resource,
		Action:               "CreateSubmissionRequest",
		ClientSparkVersion:   s.sparkVersion,
		EnvironmentVariables: job.Env,
		SparkProperties:      sparkProperties,
		AppArgs:              job.MainArgs,
	}
	if job.MainClass != "" {
		sparkRequest.MainClass = job.MainClass
	}

	// If the AppArgs user does not specify, it also needs to be specified as: []
	if len(job.MainArgs) > 0 {
		sparkRequest.AppArgs = job.MainArgs
	} else {
		sparkRequest.AppArgs = make([]string, 0)
	}
	logrus.Infof("job: %s, spark request: %+v", job.Name, sparkRequest)

	// Send creation request to Spark Server
	var sparkResp SparkResponse
	resp, err := s.client.Post(s.addr).
		Path("/v1/submissions/create").
		Header("Content-Type", "application/json").
		JSONBody(sparkRequest).
		Do().JSON(&sparkResp)

	if err != nil {
		logrus.Infof("run spark job(%s) error: %v", job.Name, err)
		return nil, errors.Errorf("run spark job(%s) error: %v", job.Name, err)
	}
	if !resp.IsOK() {
		logrus.Infof("run spark job(%s) error, statusCode: %d, response: %+v", job.Name, resp.StatusCode(), sparkResp)
		return nil, errors.Errorf("run spark job(%s) error, statusCode: %d", job.Name, resp.StatusCode())
	}

	logrus.Infof("job: %s, spark response: %+v", job.Name, sparkResp)
	return sparkResp.SubmissionId, nil
}

func (s *Spark) Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	if task.Extra.JobID == "" {
		return nil, nil
	}

	var sparkResp SparkResponse
	resp, err := s.client.Post(s.addr).
		Path(fmt.Sprintf("/v1/submissions/kill/%s", task.Extra.JobID)).
		Header("Content-Type", "application/json").
		Do().JSON(&sparkResp)

	logrus.Infof("jobId: %s delete spark job response: %+v", task.Extra.JobID, sparkResp)
	if err != nil {
		logrus.Errorf("delete spark job(%s) error: %v", task.Extra.JobID, err)
		return nil, errors.Errorf("delete spark job(%s) error: %v", task.Extra.JobID, err)
	}
	if resp.IsNotfound() {
		logrus.Infof("can not find spark job(%s)", task.Extra.JobID)
		return nil, nil
	}
	if !resp.IsOK() {
		logrus.Errorf("delete spark job(%s) error, statusCode: %d", task.Extra.JobID, resp.StatusCode())
		return nil, errors.Errorf("delete spark job(%s) error, statusCode: %d", task.Extra.JobID, resp.StatusCode())
	}

	return nil, nil
}

func (s *Spark) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error) {
	jobStatus := apistructs.StatusDesc{Status: apistructs.StatusUnknown}

	if task.Extra.JobID == "" {
		jobStatus.Status = apistructs.StatusCreated
		return jobStatus, nil
	}

	var sparkResp SparkResponse
	resp, err := s.client.Get(s.addr).
		Path(fmt.Sprintf("/v1/submissions/status/%s", task.Extra.JobID)).
		Header("Content-Type", "application/json").
		Do().JSON(&sparkResp)

	if err != nil {
		return jobStatus, errors.Errorf("get spark job(%+v) error: %v", task, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			logrus.Warnf("can't find job: %+v", task)
			return jobStatus, nil
		}
		return jobStatus, errors.Errorf("get flink job(%+v) error, statusCode: %d", task, resp.StatusCode())
	}

	jobStatus.Status = convertStatus(sparkResp.DriverState)

	return jobStatus, nil
}

func (s *Spark) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = s.Remove(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func constructSparkConstraints(r *apistructs.ScheduleInfo) string {
	// eg: dice_tags:bigdata,spark
	constrains := labelconfig.DCOS_ATTRIBUTE + ":"
	for _, like := range r.ExclusiveLikes {
		constrains = constrains + like + ","
	}
	last := len(constrains) - 1 // remove last comma
	return constrains[:last]
}

func convertStatus(sparkStatus string) apistructs.StatusCode {
	switch sparkStatus {
	case "RUNNING", "QUEUED":
		return apistructs.StatusRunning
	case "FINISHED":
		return apistructs.StatusFinished
	case "FAILED":
		return apistructs.StatusFailed
	default:
		logrus.Infof("sparkStatus: %s", sparkStatus)
		return apistructs.StatusUnknown
	}
}
