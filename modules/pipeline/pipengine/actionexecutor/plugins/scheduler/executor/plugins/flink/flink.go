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

package flink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/http/httpclient"
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

		return &Flink{
			name:      name,
			addr:      addr,
			client:    client,
			enableTag: cluster.SchedConfig.EnableTag,
			cluster:   cluster,
		}, nil
	})
}

func (f *Flink) Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	jobFromUser, err := logic.TransferToSchedulerJob(task)
	if err != nil {
		return nil, err
	}

	flinkCreateRequest := FlinkCreateRequest{
		EntryClass:  jobFromUser.MainClass,
		ProgramArgs: strings.Join(jobFromUser.MainArgs, " "),
	}

	var buffer bytes.Buffer
	resp, err := f.client.Post(f.addr).Path(fmt.Sprintf("/jars/%s/run", jobFromUser.Resource)).
		Header("Content-Type", "application/json").
		JSONBody(flinkCreateRequest).Do().Body(&buffer)
	if err != nil {
		return nil, errors.Errorf("run flink job(%s) error: %v", jobFromUser.Resource, err)
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("run flink job(%s) error, statusCode: %d, body: %v", jobFromUser.Resource, resp.StatusCode(), buffer.String())
	}
	// Return JobId
	var flinkCreateResponse FlinkCreateResponse
	r := bytes.NewReader(buffer.Bytes())
	if err := json.NewDecoder(r).Decode(&flinkCreateResponse); err != nil {
		return nil, err
	}
	logrus.Debugf("job: %s, flink response: %+v", jobFromUser.Name, flinkCreateResponse)
	return flinkCreateResponse.JobId, nil
}

func (f *Flink) Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	if task.Extra.JobID == "" {
		return nil, nil
	}

	resp, err := f.client.Get(f.addr).
		Path(fmt.Sprintf("/jobs/%s/yarn-cancel", task.Extra.JobID)).
		Header("Content-Type", "application/json; charset=UTF-8").
		Do().
		DiscardBody()

	if err != nil {
		return nil, errors.Errorf("delete flink job(%s) error: %v", task.Extra.JobID, err)
	}
	if resp.IsNotfound() {
		return nil, nil
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("run flink job(%s) error, statusCode: %d", task.Extra.JobID, resp.StatusCode())
	}
	return nil, nil
}

func (f *Flink) Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error) {
	jobStatus := apistructs.StatusDesc{Status: apistructs.StatusUnknown}

	if task.Extra.JobID == "" {
		jobStatus.Status = apistructs.StatusCreated
		return jobStatus, nil
	}

	var flinkResp FlinkGetResponse
	resp, err := f.client.Get(f.addr).
		Path(fmt.Sprintf("/jobs/%s", task.Extra.JobID)).
		Header("Content-Type", "application/json; charset=UTF-8").
		Do().
		JSON(&flinkResp)

	if err != nil {
		return jobStatus, errors.Errorf("get flink job(%+v) error: %v", task, err)
	}
	if !resp.IsOK() {
		if resp.IsNotfound() {
			logrus.Warnf("can't find job: %+v", task)
			return jobStatus, nil
		}
		return jobStatus, errors.Errorf("get flink job(%+v) error, statusCode: %d", task, resp.StatusCode())
	}

	jobStatus.Status = convertStatus(flinkResp.State)

	return jobStatus, nil
}

func (f *Flink) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = f.Remove(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func convertStatus(flinkStatus string) apistructs.StatusCode {
	switch flinkStatus {
	case "RUNNING":
		return apistructs.StatusRunning
	case "FINISHED":
		return apistructs.StatusFinished
	case "FAILED", "FAILING", "RESTARTING":
		return apistructs.StatusFailed
	case "CANCELED":
		return apistructs.StatusStopped
	default:
		logrus.Infof("flinkStatus: %s", flinkStatus)
		return apistructs.StatusUnknown
	}
}
