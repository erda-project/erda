package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/uuid"
)

const (
	NotFoundSuffix = "not found"
)

func prepareJob(job *apistructs.JobFromUser) {
	// job ID may be specified by job owner, e.g. jobflow
	if job.ID == "" {
		job.ID = uuid.Generate()
	}

	// Do not overwrite environment
	if job.Env != nil && job.Env[conf.TraceLogEnv()] == "" && job.ID != "" {
		job.Env[conf.TraceLogEnv()] = job.ID
	}
}

func (s *Server) epCreateJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	create := apistructs.JobCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		return HTTPResponse{Status: http.StatusBadRequest}, err
	}

	// check job kind
	if create.Kind == "" {
		create.Kind = string(apistructs.Metronome) // FIXME 兼容现有job kind未传的情况，后续须强制业务方传kind
	} else {
		if create.Kind != string(apistructs.Metronome) && create.Kind != string(apistructs.Flink) && create.Kind != string(apistructs.Spark) &&
			create.Kind != string(apistructs.LocalDocker) && create.Kind != string(apistructs.Kubernetes) &&
			create.Kind != string(apistructs.Swarm) && create.Kind != string(apistructs.LocalJob) {
			return HTTPResponse{Status: http.StatusBadRequest}, errors.Errorf("param[kind] is invalid")
		}

		// check job resource
		if create.Kind != string(apistructs.Metronome) && create.Resource == "" {
			return HTTPResponse{Status: http.StatusBadRequest}, errors.Errorf("param[resource] is invalid")
		}
	}

	// TODO 后续须增加clusterName强制校验
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
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobCreateResponse{
				Name:  job.Name,
				Error: "param[name] or param[namespace] is invalid",
			},
		}, nil
	}

	// 获取jobStatus，判断是否处于Running
	// 处于Running的话，不更新job到store
	update, err := s.fetchJobStatus(ctx, &job)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch job status: %s", job.Name)
	}
	if update {
		if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), &job); err != nil {
			return nil, err
		}
	}
	if job.Status == apistructs.StatusRunning {
		return nil, apistructs.ErrJobIsRunning
	}

	if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobCreateResponse{
			Job:  job,
			Name: job.Name,
		},
	}, nil
}

func (s *Server) epStartJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	job := apistructs.Job{}
	if err := s.store.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strings.HasSuffix(err.Error(), NotFoundSuffix) {
			return HTTPResponse{
				Status: http.StatusNotFound,
			}, err
		}
		return nil, err
	}

	// update job status
	job.LastModify = time.Now().String()
	job.LastStartTime = time.Now().Unix()
	job.Status = apistructs.StatusUnschedulable
	if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
		return nil, err
	}

	// update job request content(currently only update env)
	jobTmp := apistructs.Job{}
	if err := json.NewDecoder(r.Body).Decode(&jobTmp); err != nil {
		if r.ContentLength != 0 {
			logrus.Errorf("failed to decode the start input json. job=%s", job.Name)
			return HTTPResponse{Status: http.StatusBadRequest}, err
		}
	} else {
		logrus.Debugf("job start input body: %+v", jobTmp)
		if job.Env == nil {
			job.Env = make(map[string]string)
		}
		for k, v := range jobTmp.Env {
			job.Env[k] = v
		}
	}

	// build job match tags & exclude tags
	job.Labels = appendJobTags(job.Labels)

	result, err := s.handleJobTask(ctx, &job, task.TaskCreate)
	if err != nil {
		return nil, err
	}
	// get job id from flink/spark server
	if job.Kind != string(apistructs.Metronome) {
		if result.Extra != nil {
			job.ID = result.Extra.(string)
			// update job id to etcd
			if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
				return nil, err
			}
		} else {
			logrus.Errorf("[alert] can't get job id, %+v", job)
		}
	}

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobStartResponse{
			Name: name,
			Job:  job,
		},
	}, nil
}

func (s *Server) epStopJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	job := apistructs.Job{}

	if err := s.store.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strings.HasSuffix(err.Error(), NotFoundSuffix) {
			return HTTPResponse{
				Status: http.StatusNotFound,
			}, err
		}
		return nil, err
	}

	if _, err := s.handleJobTask(ctx, &job, task.TaskDestroy); err != nil {
		return nil, err
	}

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobStopResponse{
			Name: name,
		},
	}, nil
}

func (s *Server) epDeleteJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	job := apistructs.Job{}

	// 多次删除后job为空结构体, jsonstore的remove接口可以再添加一个返回值判断job是否被填充
	if err := s.store.Remove(ctx, makeJobKey(namespace, name), &job); err != nil {
		return nil, err
	}

	if len(job.Name) == 0 {
		return HTTPResponse{
			Status: http.StatusOK,
			Content: apistructs.JobDeleteResponse{
				Name:  name,
				Error: "that name not found",
			},
		}, nil
	}

	if err := s.deleteJob(ctx, &job); err != nil {
		return nil, err
	}

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobDeleteResponse{
			Name: name,
		},
	}, nil
}

func (s *Server) epJobPipeline(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	namespace := vars["namespace"]

	var names []string
	if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
		return nil, err
	}
	// max job number is 10 by supported.
	if len(names) > 10 {
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Names: names,
				Error: "the jobs are too many, max number is 10.",
			},
		}, nil
	}

	// read all jobs from store.
	jobs := make([]apistructs.Job, len(names))

	for i, name := range names {
		if err := s.store.Get(ctx, makeJobKey(namespace, name), &jobs[i]); err != nil {
			if strings.HasSuffix(err.Error(), NotFoundSuffix) {
				return HTTPResponse{
					Status: http.StatusNotFound,
				}, err
			}
			return nil, err
		}
	}

	// start all jobs one by one.
	for i := range jobs {
		job := &jobs[i]
		if err := s.startOneJob(ctx, job); err != nil {
			job.LastMessage = err.Error()

			return HTTPResponse{
				Status: http.StatusInternalServerError,
				Content: apistructs.JobBatchResponse{
					Names: names,
					Jobs:  jobs,
					Error: "failed to pipeline jobs",
				},
			}, nil
		}
	}

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobBatchResponse{
			Names: names,
			Jobs:  jobs,
		},
	}, nil
}

func (s *Server) epJobConcurrent(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	namespace := vars["namespace"]

	var names []string
	if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
		return nil, err
	}
	// max job number is 10 by supported.
	if len(names) > 10 {
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Names: names,
				Error: "the jobs are too many, max number is 10.",
			},
		}, nil
	}

	// read all jobs from store.
	jobs := make([]apistructs.Job, len(names))

	for i, name := range names {
		if err := s.store.Get(ctx, makeJobKey(namespace, name), &jobs[i]); err != nil {
			if strings.HasSuffix(err.Error(), NotFoundSuffix) {
				return HTTPResponse{
					Status: http.StatusNotFound,
				}, err
			}
			return nil, err
		}
	}

	var wg sync.WaitGroup

	for i := range jobs {
		wg.Add(1)

		go func(j int) {
			defer wg.Done()

			job := &jobs[j]
			if err := s.startOneJob(ctx, job); err != nil {
				job.LastMessage = err.Error()
			}
		}(i)
	}

	wg.Wait()

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.JobBatchResponse{
			Names: names,
			Jobs:  jobs,
		},
	}, nil
}

func (s *Server) epListJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	namespace := vars["namespace"]

	jobs := make([]*apistructs.Job, 0, 100)

	err := s.store.ForEach(ctx, makeJobKey(namespace, ""), apistructs.Job{}, func(key string, j interface{}) error {
		job := j.(*apistructs.Job)
		update, err := s.fetchJobStatus(ctx, job)
		if err != nil {
			return err
		}
		if update {
			if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
				return err
			}
		}
		jobs = append(jobs, j.(*apistructs.Job))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status:  http.StatusOK,
		Content: jobs,
	}, nil
}

func (s *Server) epGetJob(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	job := apistructs.Job{}

	if err := s.store.Get(ctx, makeJobKey(namespace, name), &job); err != nil {
		if strings.HasSuffix(err.Error(), NotFoundSuffix) {
			return HTTPResponse{
				Status: http.StatusNotFound,
			}, err
		}
		return nil, err
	}

	update, err := s.fetchJobStatus(ctx, &job)
	if err != nil {
		return nil, err
	}
	if update {
		if err := s.store.Put(ctx, makeJobKey(namespace, name), &job); err != nil {
			return nil, err
		}
	}

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: job,
	}, nil
}

func (s *Server) startOneJob(ctx context.Context, job *apistructs.Job) error {
	job.LastStartTime = time.Now().Unix()
	job.Status = apistructs.StatusRunning

	if err := s.store.Put(ctx, makeJobKey(job.Namespace, job.Name), job); err != nil {
		logrus.Warnf("failed to update job status: %s (%v)", job.Name, err)
	}

	// build job match tags & exclude tags
	job.Labels = appendJobTags(job.Labels)

	_, err := s.handleJobTask(ctx, job, task.TaskCreate)
	return err
}

func (s *Server) fetchJobStatus(ctx context.Context, job *apistructs.Job) (bool, error) {
	result, err := s.handleJobTask(ctx, job, task.TaskStatus)
	if err != nil {
		return false, err
	}

	if result.Status().Status == apistructs.StatusCode("") {
		return false, nil
	}

	if job.StatusDesc.Status != result.Status().Status || job.StatusDesc.LastMessage != result.Status().LastMessage {
		job.LastModify = time.Now().String()
		job.StatusDesc = result.Status()
		return true, nil
	}
	return false, nil
}

func (s *Server) deleteJob(ctx context.Context, job *apistructs.Job) error {
	_, err := s.handleJobTask(ctx, job, task.TaskRemove)
	return err
}

// add job kind to label based on which, scheduler can schedule different kind of job resources.
func modifyLabelsWithJobKind(labels map[string]string) {
	if kind, ok := labels[apistructs.LabelJobKind]; ok {
		matchTagsStr := labels[apistructs.LabelMatchTags]
		matchTagsSlice := strings.Split(matchTagsStr, ",")
		matchTagsSlice = append(matchTagsSlice, kind)
		matchTags := strings.Join(matchTagsSlice, ",")
		labels[apistructs.LabelMatchTags] = matchTags
	}
}

func appendJobTags(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	matchTags := make([]string, 0)
	if kind, ok := labels[apistructs.LabelJobKind]; !ok || kind != apistructs.TagBigdata {
		matchTags = append(matchTags, apistructs.TagJob)
	}
	if labels[apistructs.LabelPack] == "true" {
		matchTags = append(matchTags, apistructs.TagPack)
	}

	labels[apistructs.LabelMatchTags] = strings.Join(matchTags, ",")
	if _, ok := labels[apistructs.LabelExcludeTags]; ok {
		labels[apistructs.LabelExcludeTags] = labels[apistructs.LabelExcludeTags] + "," + apistructs.TagLocked + "," + apistructs.TagPlatform
	} else {
		labels[apistructs.LabelExcludeTags] = apistructs.TagLocked + "," + apistructs.TagPlatform
	}

	modifyLabelsWithJobKind(labels)
	return labels
}

func (s *Server) handleJobTask(ctx context.Context, job *apistructs.Job, action task.Action) (task.TaskResponse, error) {
	if err := clusterutil.SetJobExecutorByCluster(job); err != nil {
		return task.TaskResponse{}, err
	}

	tsk, err := s.sched.Send(ctx, task.TaskRequest{
		ExecutorKind: executor.GetJobExecutorKindByName(job.Executor),
		ExecutorName: job.Executor,
		Action:       action,
		ID:           job.Name,
		Spec:         *job,
	})
	if err != nil {
		return task.TaskResponse{}, err
	}

	result := tsk.Wait(ctx)
	return result, result.Err()
}
