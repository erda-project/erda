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

package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/jsonstore"
)

func (j *JobImpl) Create(create apistructs.JobCreateRequest) (apistructs.Job, error) {
	// check job kind
	if create.Kind == "" {
		create.Kind = string(apistructs.Metronome) // FIXME Compatible with the untransmitted situation of the existing job kind, the business side must be forced to transmit the kind
	} else {
		if create.Kind != string(apistructs.Metronome) && create.Kind != string(apistructs.Flink) &&
			create.Kind != string(apistructs.K8SFlink) && create.Kind != string(apistructs.Spark) &&
			create.Kind != string(apistructs.K8SSpark) && create.Kind != string(apistructs.LocalDocker) &&
			create.Kind != string(apistructs.Kubernetes) && create.Kind != string(apistructs.Swarm) &&
			create.Kind != string(apistructs.LocalJob) {
			return apistructs.Job{}, errors.Errorf("param [kind:%s] is invalid", create.Kind)
		}

		// check job resource
		if create.Kind != string(apistructs.Metronome) && create.Resource == "" {
			return apistructs.Job{}, errors.Errorf("param[resource] is invalid")
		}
	}

	// TODO: Mandatory verification of clusterName must be added in the follow-up
	logrus.Infof("epCreateJob job: %+v", create)
	job := apistructs.Job{
		JobFromUser: apistructs.JobFromUser(create),
		CreatedTime: time.Now().Unix(),
		LastModify:  time.Now().String(),
	}

	// Flink & Spark doesn't need generate jobId, it will be generated on flink/spark server
	if job.Kind == string(apistructs.Metronome) {
		prepareJob(&job.JobFromUser)
	}

	if job.Namespace == "" {
		job.Namespace = "default"
	}
	job.Status = apistructs.StatusCreated

	if !validateJobName(job.Name) && !validateJobNamespace(job.Namespace) {
		return apistructs.Job{}, errors.New("param[name] or param[namespace] is invalid")
	}

	// transform job kind to k8sflink
	if value, ok := job.Params["bigDataConf"]; ok {
		job.BigdataConf = apistructs.BigdataConf{
			BigdataMetadata: apistructs.BigdataMetadata{
				Name:      job.Name,
				Namespace: job.Namespace,
			},
			Spec: apistructs.BigdataSpec{},
		}
		err := json.Unmarshal([]byte(value.(string)), &job.BigdataConf.Spec)
		if err != nil {
			return apistructs.Job{}, fmt.Errorf("unmarshal bigdata config error: %s", err.Error())
		}
		if job.BigdataConf.Spec.FlinkConf != nil {
			job.Kind = string(apistructs.K8SFlink)
		}
		if job.BigdataConf.Spec.SparkConf != nil {
			job.Kind = string(apistructs.K8SSpark)
		}
	}

	// Get jobStatus, determine whether it is Running
	// If you are in Running, do not update the job to the store
	ctx := context.Background()
	update, err := j.fetchJobStatus(ctx, &job)
	if err != nil {
		return apistructs.Job{}, errors.Wrapf(err, "failed to fetch job status: %s", job.Name)
	}
	if update {
		if err := j.js.Put(ctx, makeJobKey(job.Namespace, job.Name), &job); err != nil {
			return apistructs.Job{}, err
		}
	}
	if job.Status == apistructs.StatusRunning {
		return apistructs.Job{}, apistructs.ErrJobIsRunning
	}

	if err := j.js.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
		return apistructs.Job{}, err
	}
	jobsinfo := apistructs.Job{}
	if err := j.js.Get(ctx, makeJobKey(job.Namespace, ""), &jobsinfo); err != nil {
		if err == jsonstore.NotFoundErr {
			if err := j.js.Put(ctx, makeJobKey(job.Namespace, ""), job); err != nil {
				return apistructs.Job{}, err
			}
		} else {
			return apistructs.Job{}, err
		}
	}

	return job, nil
}
