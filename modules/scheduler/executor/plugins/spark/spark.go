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

package spark

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/schedule/schedulepolicy/labelconfig"
)

const (
	kind = "SPARK"
)

// Spark API目前没有官方文档，具体请查看Spark源码: RestSubmissionClient
type Spark struct {
	name          executortypes.Name
	addr          string
	options       map[string]string
	client        *httpclient.HTTPClient
	enableTag     bool   // Whether to enable label scheduling
	sparkVersion  string // Spark version
	executorImage string // Spark Executor image eg: mesosphere/spark:2.3.1-2.2.1-2-hadoop-2.6
}

func init() {
	executortypes.Register(kind, func(name executortypes.Name, clustername string, options map[string]string, optionsPlus interface{}) (executortypes.Executor, error) {
		addr, ok := options["ADDR"]
		if !ok {
			return nil, errors.Errorf("not found spark address in env variables")
		}

		client := httpclient.New()

		if _, ok := options["CA_CRT"]; ok {
			logrus.Infof("flink executor(%s) addr for https: %v", name, addr)
			client = httpclient.New(httpclient.WithHttpsCertFromJSON([]byte(options["CLIENT_CRT"]),
				[]byte(options["CLIENT_KEY"]),
				[]byte(options["CA_CRT"])))
		}

		enableTag, err := util.ParseEnableTagOption(options, "ENABLETAG", true)
		if err != nil {
			return nil, err
		}

		sparkVersion, ok := options["SPARK_VERSION"]
		if !ok {
			return nil, errors.Errorf("not found spark version in env variables")
		}

		executorImage, ok := options["SPARK_EXECUTOR_IMAGE"]
		if !ok {
			return nil, errors.Errorf("not found spark executor image in env variables")
		}

		return &Spark{
			name:          name,
			addr:          addr,
			options:       options,
			client:        client,
			enableTag:     enableTag,
			sparkVersion:  sparkVersion,
			executorImage: executorImage,
		}, nil
	})
}

func (s *Spark) Kind() executortypes.Kind {
	return kind
}

func (s *Spark) Name() executortypes.Name {
	return s.name
}

func (s *Spark) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	// Construct Spark task request structure, configuration can refer to: https://spark.apache.org/docs/latest/running-on-mesos.html#configuration
	sparkProperties := make(map[string]string)
	sparkProperties["spark.app.name"] = job.Name
	sparkProperties["spark.driver.supervise"] = "true"
	sparkProperties["spark.eventLog.enabled"] = "false"
	sparkProperties["spark.submit.deployMode"] = "cluster"
	sparkProperties["spark.mesos.executor.docker.image"] = s.executorImage
	constrains := constructSparkConstraints(&job.ScheduleInfo)
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

func (s *Spark) Destroy(ctx context.Context, specObj interface{}) error {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return errors.New("invalid job spec")
	}

	if job.ID == "" {
		return nil
	}

	var sparkResp SparkResponse
	resp, err := s.client.Post(s.addr).
		Path(fmt.Sprintf("/v1/submissions/kill/%s", job.ID)).
		Header("Content-Type", "application/json").
		Do().JSON(&sparkResp)

	logrus.Infof("jobId: %s delete spark job response: %+v", job.ID, sparkResp)
	if err != nil {
		logrus.Errorf("delete spark job(%s) error: %v", job.ID, err)
		return errors.Errorf("delete spark job(%s) error: %v", job.ID, err)
	}
	if resp.IsNotfound() {
		logrus.Infof("can not find spark job(%s)", job.ID)
		return nil
	}
	if !resp.IsOK() {
		logrus.Errorf("delete spark job(%s) error, statusCode: %d", job.ID, resp.StatusCode())
		return errors.Errorf("delete spark job(%s) error, statusCode: %d", job.ID, resp.StatusCode())
	}

	return nil
}

func (s *Spark) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	jobStatus := apistructs.StatusDesc{Status: apistructs.StatusUnknown}

	job, ok := specObj.(apistructs.Job)
	if !ok {
		return jobStatus, errors.New("invalid job spec")
	}

	if job.ID == "" {
		jobStatus.Status = apistructs.StatusCreated
		return jobStatus, nil
	}

	var sparkResp SparkResponse
	resp, err := s.client.Get(s.addr).
		Path(fmt.Sprintf("/v1/submissions/status/%s", job.ID)).
		Header("Content-Type", "application/json").
		Do().JSON(&sparkResp)

	if err != nil {
		return jobStatus, errors.Errorf("get spark job(%+v) error: %v", job, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			logrus.Warnf("can't find job: %+v", job)
			return jobStatus, nil
		}
		return jobStatus, errors.Errorf("get flink job(%+v) error, statusCode: %d", job, resp.StatusCode())
	}

	jobStatus.Status = convertStatus(sparkResp.DriverState)

	return jobStatus, nil
}

func (s *Spark) Remove(ctx context.Context, specObj interface{}) error {
	return s.Destroy(ctx, specObj)
}

func (s *Spark) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	return nil, errors.Errorf("job(spark) not support update action")
}

func (s *Spark) Inspect(ctx context.Context, specObj interface{}) (interface{}, error) {
	job, ok := specObj.(apistructs.Job)
	if !ok {
		return nil, errors.New("invalid job spec")
	}

	if job.ID == "" {
		return job, nil
	}

	var sparkResp SparkResponse
	resp, err := s.client.Get(s.addr).
		Path(fmt.Sprintf("/v1/submissions/status/%s", job.ID)).
		Header("Content-Type", "application/json").
		Do().JSON(&sparkResp)

	logrus.Infof("get spark response: %+v", sparkResp)
	if err != nil {
		logrus.Errorf("get spark job(%s) error: %+v", job.ID, err)
		return nil, errors.Errorf("get spark job(%s) error: %+v", job.ID, err)
	}
	if !resp.IsOK() {
		logrus.Errorf("get spark job(%s) error, stautsCode: %d", job.ID, resp.StatusCode())
		return nil, errors.Errorf("get spark job(%s) error, stautsCode: %d", job.ID, resp.StatusCode())
	}
	job.Status = convertStatus(sparkResp.DriverState)

	return job, nil
}

func (s *Spark) Cancel(ctx context.Context, spec interface{}) (interface{}, error) {
	return nil, errors.New("job(spark) not support Cancel action")
}
func (s *Spark) Precheck(ctx context.Context, specObj interface{}) (apistructs.ServiceGroupPrecheckData, error) {
	return apistructs.ServiceGroupPrecheckData{Status: "ok"}, nil
}

func (s *Spark) CapacityInfo() apistructs.CapacityInfoData {
	return apistructs.CapacityInfoData{}
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

func constructSparkConstraints(r *apistructs.ScheduleInfo) string {
	// eg: dice_tags:bigdata,spark
	constrains := labelconfig.DCOS_ATTRIBUTE + ":"
	for _, like := range r.ExclusiveLikes {
		constrains = constrains + like + ","
	}
	last := len(constrains) - 1 // remove last comma
	return constrains[:last]
}

func (m *Spark) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return errors.New("SetNodeLabels not implemented in Spark")
}
func (m *Spark) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	return apistructs.ClusterResourceInfoData{}, fmt.Errorf("resourceinfo not support for spark")
}
func (*Spark) CleanUpBeforeDelete() {}
func (*Spark) JobVolumeCreate(ctx context.Context, spec interface{}) (string, error) {
	return "", fmt.Errorf("not support for spark")
}
func (*Spark) KillPod(podname string) error {
	return fmt.Errorf("not support for spark")
}

func (*Spark) Scale(ctx context.Context, spec interface{}) (interface{}, error) {
	return apistructs.ServiceGroup{}, fmt.Errorf("scale not support for spark")
}
