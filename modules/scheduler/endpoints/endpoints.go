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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/impl/cap"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/instanceinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/job"
	"github.com/erda-project/erda/modules/scheduler/impl/labelmanager"
	"github.com/erda-project/erda/modules/scheduler/impl/resourceinfo"
	"github.com/erda-project/erda/modules/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/modules/scheduler/impl/terminal"
	"github.com/erda-project/erda/modules/scheduler/impl/volume"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type HTTPEndpoints struct {
	volumeImpl        volume.Volume
	serviceGroupImpl  servicegroup.ServiceGroup
	clusterImpl       cluster.Cluster
	job               job.Job
	labelManager      labelmanager.LabelManager
	instanceinfoImpl  instanceinfo.InstanceInfo
	clusterinfoImpl   clusterinfo.ClusterInfo
	componentinfoImpl instanceinfo.ComponentInfo
	resourceinfoImpl  resourceinfo.ResourceInfo
	cap               cap.Cap
	//metric            metric.Metric
	// TODO: add more impl here
}

func NewHTTPEndpoints(
	volume volume.Volume,
	servicegroup servicegroup.ServiceGroup,
	cluster cluster.Cluster,
	job job.Job,
	labelManager labelmanager.LabelManager,
	instanceinfo instanceinfo.InstanceInfo,
	clusterinfo clusterinfo.ClusterInfo,
	componentinfo instanceinfo.ComponentInfo,
	resourceinfo resourceinfo.ResourceInfo,
	cap cap.Cap) *HTTPEndpoints {
	return &HTTPEndpoints{
		volume,
		servicegroup,
		cluster,
		job,
		labelManager,
		instanceinfo,
		clusterinfo,
		componentinfo,
		resourceinfo,
		cap,
	}
}

const (
	ENABLE_SPECIFIED_K8S_NAMESPACE = "ENABLE_SPECIFIED_K8S_NAMESPACE"
	RetainNamespace                = "RETAIN_NAMESPACE"
)

// Routes scheduler
// TODO: Currently there are only servicegroup and volume APIs,，
//       There is also job API that needs to be migrated
func (h *HTTPEndpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/api/volumes", Method: http.MethodPost, Handler: h.VolumeCreate},
		{Path: "/api/volumes/{id}", Method: http.MethodDelete, Handler: h.VolumeDelete},
		{Path: "/api/volumes/{id}", Method: http.MethodGet, Handler: h.VolumeInfo},

		{Path: "/api/servicegroup", Method: http.MethodPost, Handler: h.ServiceGroupCreate},
		{Path: "/api/servicegroup", Method: http.MethodPut, Handler: h.ServiceGroupUpdate},
		{Path: "/api/servicegroup", Method: http.MethodDelete, Handler: h.ServiceGroupDelete},
		{Path: "/api/servicegroup", Method: http.MethodGet, Handler: h.ServiceGroupInfo},
		{Path: "/api/servicegroup/actions/restart", Method: http.MethodPost, Handler: h.ServiceGroupRestart},
		{Path: "/api/servicegroup/actions/cancel", Method: http.MethodPost, Handler: h.ServiceGroupCancel},
		{Path: "/api/clusterinfo/{clusterName}", Method: http.MethodGet, Handler: h.ClusterInfo},
		{Path: "/api/resourceinfo/{clusterName}", Method: http.MethodGet, Handler: h.ResourceInfo},
	}
}

func (h *HTTPEndpoints) VolumeCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.VolumeCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode create volume request fail: %v", err)
		return mkResponse(apistructs.VolumeCreateResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	config, err := volume.VolumeCreateConfigFrom(req)
	if err != nil {
		errstr := fmt.Sprintf("create volume: convert create config fail: %v", err)
		return mkResponse(apistructs.VolumeCreateResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}

	info, err := h.volumeImpl.Create(config)
	if err != nil {
		errstr := fmt.Sprintf("create volume fail: %v", err)
		return mkResponse(apistructs.VolumeCreateResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.VolumeCreateResponse{
		apistructs.Header{Success: true},
		info,
	})
}

func (h *HTTPEndpoints) VolumeDelete(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	id := vars["id"]
	if id == "" {
		errstr := fmt.Sprintf("empty volume id")
		return mkResponse(apistructs.VolumeDeleteResponse{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}

	if err := h.volumeImpl.Delete(volume.VolumeIdentity(id), true); err != nil {
		errstr := fmt.Sprintf("delete volume fail: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.VolumeDeleteResponse{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.VolumeDeleteResponse{
		apistructs.Header{Success: true},
	})

}

func (h *HTTPEndpoints) VolumeInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	id := vars["id"]
	if id == "" {
		errstr := fmt.Sprintf("empty volume id")
		return mkResponse(apistructs.VolumeDeleteResponse{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	info, err := h.volumeImpl.Info(volume.VolumeIdentity(id))
	if err != nil {
		errstr := fmt.Sprintf("get volume info fail: %v", err)
		return mkResponse(apistructs.VolumeInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.VolumeInfoResponse{
		apistructs.Header{Success: true},
		info,
	})
}

func (h *HTTPEndpoints) ServiceGroupCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	//h.metric.TotalCounter.WithLabelValues(metric.ServiceCreateTotal).Add(1)
	req := apistructs.ServiceGroupCreateV2Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode create servicegroup request fail: %v", err)
		return mkResponse(apistructs.ServiceGroupCreateV2Response{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	sg, err := h.serviceGroupImpl.Create(req)
	if err != nil {
		errstr := fmt.Sprintf("create servicegroup fail: %v", err)
		//h.metric.ErrorCounter.WithLabelValues(metric.ServiceCreateError).Add(1)
		return mkResponse(apistructs.ServiceGroupCreateV2Response{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.ServiceGroupCreateV2Response{
		Header: apistructs.Header{Success: true},
		Data: apistructs.ServiceGroupCreateV2Data{
			ID:   sg.ID,
			Type: sg.Type,
		},
	})
}

func (h *HTTPEndpoints) ServiceGroupUpdate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.ServiceGroupUpdateV2Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode update servicegroup request fail: %v", err)
		return mkResponse(apistructs.ServiceGroupUpdateV2Response{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	sg, err := h.serviceGroupImpl.Update(req)
	if err != nil {
		errstr := fmt.Sprintf("update servicegroup fail: %v", err)
		return mkResponse(apistructs.ServiceGroupUpdateV2Response{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.ServiceGroupUpdateV2Response{
		Header: apistructs.Header{Success: true},
		Data: apistructs.ServiceGroupCreateV2Data{
			Type: sg.Type,
			ID:   sg.ID,
		},
	})
}

func (h *HTTPEndpoints) ServiceGroupDelete(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	//h.metric.TotalCounter.WithLabelValues(metric.ServiceRemoveTotal).Add(1)

	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	force := r.URL.Query().Get("force")
	if namespace == "" || name == "" {
		errstr := fmt.Sprintf("empty namespace or name")
		return mkResponse(apistructs.ServiceGroupDeleteV2Response{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}

	if err := h.serviceGroupImpl.Delete(namespace, name, force); err != nil {
		errstr := fmt.Sprintf("delete servicegroup fail: %v", err)
		//h.metric.ErrorCounter.WithLabelValues(metric.ServiceRemoveError).Add(1)
		if err == servicegroup.DeleteNotFound {
			return mkResponse(apistructs.ServiceGroupDeleteV2Response{
				apistructs.Header{
					Success: false,
					Error: apistructs.ErrorResponse{
						Code: "404",
						Msg:  errstr,
					},
				}})
		}

		return mkResponse(apistructs.ServiceGroupDeleteV2Response{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.ServiceGroupDeleteV2Response{
		apistructs.Header{Success: true}})
}

func (h *HTTPEndpoints) ServiceGroupInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		errstr := fmt.Sprintf("empty namespace or name")
		return mkResponse(apistructs.ServiceGroupInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	sg, err := h.serviceGroupImpl.Info(ctx, namespace, name)
	if err != nil {
		errstr := fmt.Sprintf("get servicegroup info fail: %v", err)
		return mkResponse(apistructs.ServiceGroupInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupInfoResponse{
		apistructs.Header{Success: true},
		sg,
	})
}

func (h *HTTPEndpoints) ServiceGroupRestart(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		errstr := fmt.Sprintf("empty namespace or name")
		return mkResponse(apistructs.ServiceGroupRestartV2Response{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	if err := h.serviceGroupImpl.Restart(namespace, name); err != nil {
		errstr := fmt.Sprintf("restart servicegroup fail: %v", err)
		return mkResponse(apistructs.ServiceGroupRestartV2Response{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupRestartV2Response{
		apistructs.Header{Success: true},
	})
}

func (h *HTTPEndpoints) ServiceGroupCancel(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		errstr := fmt.Sprintf("empty namespace or name")
		return mkResponse(apistructs.ServiceGroupInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	if err := h.serviceGroupImpl.Restart(namespace, name); err != nil {
		errstr := fmt.Sprintf("restart servicegroup fail: %v", err)
		return mkResponse(apistructs.ServiceGroupRestartV2Response{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupRestartV2Response{
		apistructs.Header{Success: true},
	})

}

func (h *HTTPEndpoints) ServiceGroupPrecheck(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.ServiceGroupPrecheckRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode precheck servicegroup request fail: %v", err)
		return mkResponse(apistructs.ServiceGroupPrecheckResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	res, err := h.serviceGroupImpl.Precheck(req)
	if err != nil {
		errstr := fmt.Sprintf("precheck servicegroup fail: %v", err)
		return mkResponse(apistructs.ServiceGroupPrecheckResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupPrecheckResponse{
		Header: apistructs.Header{Success: true},
		Data:   res,
	})

}

func (h *HTTPEndpoints) ServiceGroupKillPod(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.ServiceGroupKillPodRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode killpod request fail: %v", err)
		return mkResponse(apistructs.ServiceGroupKillPodResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	if req.Namespace == "" || req.Name == "" || req.PodName == "" {
		errstr := fmt.Sprintf("empty namespace or name or containerID")
		return mkResponse(apistructs.ServiceGroupKillPodResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	err := h.serviceGroupImpl.KillPod(ctx, req.Namespace, req.Name, req.PodName)
	if err != nil {
		return mkResponse(apistructs.ServiceGroupKillPodResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupKillPodResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}

func (h *HTTPEndpoints) ServiceGroupConfigUpdate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.ServiceGroup{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("configupdate decode servicegroup request fail: %v", err)
		return mkResponse(apistructs.ServiceGroupConfigUpdateResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	namespace := req.Type
	name := req.ID
	if namespace == "" || name == "" {
		errstr := fmt.Sprintf("empty namespace or name")
		return mkResponse(apistructs.ServiceGroupConfigUpdateResponse{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	if err := h.serviceGroupImpl.ConfigUpdate(req); err != nil {
		errstr := fmt.Sprintf("configupdate servicegroup fail: %v", err)
		return mkResponse(apistructs.ServiceGroupConfigUpdateResponse{
			apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.ServiceGroupConfigUpdateResponse{
		apistructs.Header{Success: true},
	})
}

func (h *HTTPEndpoints) ClusterHook(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.ClusterEvent{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode clusterhook request fail: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{Status: http.StatusBadRequest, Content: errstr}, nil
	}
	if err := h.clusterImpl.Hook(&req); err != nil {
		errstr := fmt.Sprintf("failed to handle cluster event: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError, Content: errstr}, nil
	}
	return httpserver.HTTPResponse{Status: http.StatusOK}, nil
}

func (h *HTTPEndpoints) ClusterCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return httpserver.HTTPResponse{Status: http.StatusGone}, nil
}

func (h *HTTPEndpoints) JobCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	//h.metric.TotalCounter.WithLabelValues(metric.JobCreateTotal).Add(1)
	req := apistructs.JobCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode create volume request fail: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobCreateResponse{
				Error: errstr,
			},
		}, nil
	}

	// specify namespace from scheduler ENV 'ENABLE_SPECIFIED_K8S_NAMESPACE'
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		req.Namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	job, err := h.job.Create(req)
	if err != nil {
		//h.metric.ErrorCounter.WithLabelValues(metric.JobCreateError).Add(1)
		errstr := fmt.Sprintf("failed to create job: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobCreateResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobCreateResponse{
		Name: job.Name,
		Job:  job,
	})
}

func (h *HTTPEndpoints) JobStart(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]
	if name == "" || namespace == "" {
		errstr := "failed to start job, empty name or namespace"
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobStartResponse{
				Error: errstr,
			},
		}, nil
	}
	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	resultJob, err := h.job.Start(namespace, name, map[string]string{})
	if err != nil {
		errstr := fmt.Sprintf("failed to start job, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusInternalServerError,
			Content: apistructs.JobStartResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobStartResponse{
		Name: name,
		Job:  resultJob,
	})
}

func (h *HTTPEndpoints) JobVolumeCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	req := apistructs.JobVolume{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("decode create jobvolume request fail: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobVolumeCreateResponse{
				Error: errstr,
			},
		}, nil
	}
	id, err := h.job.CreateJobVolume(req)
	if err != nil {
		errstr := fmt.Sprintf("failed to create job volume: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobVolumeCreateResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobVolumeCreateResponse{ID: id})

}
func (h *HTTPEndpoints) JobStop(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	if name == "" || namespace == "" {
		errstr := "failed to stop job, empty name or namespace"
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobStopResponse{
				Error: errstr,
			},
		}, nil
	}

	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	if err := h.job.Stop(namespace, name); err != nil {
		errstr := fmt.Sprintf("failed to stop job, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobStopResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobStopResponse{Name: name})
}

func (h *HTTPEndpoints) JobDelete(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var job apistructs.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil && r.ContentLength != 0 {
		errstr := fmt.Sprintf("failed to decode jobStart body, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobDeleteResponse{
				Error: errstr,
			},
		}, nil
	}

	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		job.Namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	if job.Env == nil {
		job.Env = make(map[string]string, 0)
	}
	job.Env[RetainNamespace] = "true"

	if err := h.job.Delete(job); err != nil {
		errstr := fmt.Sprintf("failed to delete job, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobDeleteResponse{
				Name:      job.Name,
				Namespace: job.Namespace,
				Error:     errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobDeleteResponse{Name: job.Name, Namespace: job.Namespace})
}

// batch Delete Jobs will set retainNamespace is false
// so that the namespace will be deleted when the
// job count is zero
func (h *HTTPEndpoints) DeleteJobs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var jobs []apistructs.Job
	if err := json.NewDecoder(r.Body).Decode(&jobs); err != nil && r.ContentLength != 0 {
		errstr := fmt.Sprintf("failed to decode jobStart body, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: []apistructs.JobDeleteResponse{
				{
					Error: errstr,
				},
			},
		}, nil
	}

	deleteResponseList := apistructs.JobsDeleteResponse{}
	logrus.Infof("batch delete %d jobs", len(jobs))

	for _, job := range jobs {
		if job.Env == nil {
			job.Env = make(map[string]string, 0)
		}
		job.Env[RetainNamespace] = "false"

		if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
			job.Namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
		}

		if err := h.job.Delete(job); err != nil {
			errstr := fmt.Sprintf("failed to delete job %s in ns %s, err: %v", job.Name, job.Namespace, err)
			logrus.Error(errstr)
			deleteResponseList = append(deleteResponseList, apistructs.JobDeleteResponse{
				Name:      job.Name,
				Namespace: job.Namespace,
				Error:     errstr,
			})
		}
	}

	if len(deleteResponseList) != 0 {
		return httpserver.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: deleteResponseList,
		}, nil
	}
	return mkResponse(apistructs.JobsDeleteResponse{})
}

func (h *HTTPEndpoints) JobInspect(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	if name == "" || namespace == "" {
		errstr := "failed to inspect job, empty name or namespace"
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: errstr,
		}, nil
	}

	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE)
	}

	job, err := h.job.Inspect(namespace, name)
	if err != nil {
		errstr := fmt.Sprintf("failed to inspect job, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Content: errstr,
		}, nil
	}
	return mkResponse(job)
}
func (h *HTTPEndpoints) JobList(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := vars["namespace"]

	if namespace == "" {
		errstr := "failed to list job, empty namespace"
		return httpserver.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: errstr,
		}, nil
	}

	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = ENABLE_SPECIFIED_K8S_NAMESPACE
	}

	jobs, err := h.job.List(namespace)
	if err != nil {
		errstr := fmt.Sprintf("failed to list jobs, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Content: errstr,
		}, nil
	}
	return mkResponse(jobs)
}

func (h *HTTPEndpoints) JobPipeline(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := vars["namespace"]

	if namespace == "" {
		errstr := "failed to pipeline jobs, empty namespace"
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}

	var names []string
	if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
		errstr := fmt.Sprintf("failed to decode jobPipeline request body, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}

	jobs, err := h.job.Pipeline(namespace, names)
	if err != nil {
		errstr := fmt.Sprintf("failed to pipeline jobs, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusInternalServerError,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobBatchResponse{
		Names: names,
		Jobs:  jobs,
	})
}

func (h *HTTPEndpoints) JobConcurrent(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	namespace := vars["namespace"]
	if namespace == "" {
		errstr := "failed to concurrently run jobs, empty namespace"
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}

	var names []string
	if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
		errstr := fmt.Sprintf("failed to decode JobConcurrent request body, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}

	if os.Getenv(ENABLE_SPECIFIED_K8S_NAMESPACE) != "" {
		namespace = ENABLE_SPECIFIED_K8S_NAMESPACE
	}

	jobs, err := h.job.Concurrent(namespace, names)
	if err != nil {
		errstr := fmt.Sprintf("failed to concurrently run jobs, err: %v", err)
		logrus.Error(errstr)
		return httpserver.HTTPResponse{
			Status: http.StatusInternalServerError,
			Content: apistructs.JobBatchResponse{
				Error: errstr,
			},
		}, nil
	}
	return mkResponse(apistructs.JobBatchResponse{
		Names: names,
		Jobs:  jobs,
	})

}

func (h *HTTPEndpoints) LabelList(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return mkResponse(apistructs.ScheduleLabelListResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.ScheduleLabelListData{Labels: h.labelManager.List()},
	})
}

// {"force":false,"tag":"any,workspace-dev,workspace-test,workspace-staging,workspace-prod,job,service-stateful,org-terminus","hosts":["10.168.0.100"]}
func (h *HTTPEndpoints) SetNodeLabels(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	//h.metric.TotalCounter.WithLabelValues(metric.LableTotal).Add(1)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errstr := "Invalid request body"
		logrus.Error(errstr)
		return mkResponse(apistructs.ScheduleLabelSetResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				}},
		})
	}
	var reqBody apistructs.ScheduleLabelSetRequest
	if err := json.Unmarshal(body, &reqBody); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal request body: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.ScheduleLabelSetResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}

	if err := h.labelManager.SetNodeLabel(labelmanager.Cluster{
		ClusterName: reqBody.ClusterName,
		ClusterType: reqBody.ClusterType,
		SoldierURL:  reqBody.SoldierURL,
	}, reqBody.Hosts, reqBody.Tags); err != nil {
		//h.metric.ErrorCounter.WithLabelValues(metric.LableError).Add(1)
		errstr := fmt.Sprintf("failed to set node labels: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.ScheduleLabelSetResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}
	return mkResponse(apistructs.ScheduleLabelSetResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}

func (h *HTTPEndpoints) PodInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	cond := instanceinfo.QueryPodConditions{}
	cond.Cluster = r.URL.Query().Get("cluster")
	cond.OrgName = r.URL.Query().Get("orgName")
	cond.OrgID = r.URL.Query().Get("orgID")
	cond.ProjectName = r.URL.Query().Get("projectName")
	cond.ProjectID = r.URL.Query().Get("projectID")
	cond.ApplicationName = r.URL.Query().Get("applicationName")
	cond.ApplicationID = r.URL.Query().Get("applicationID")
	cond.RuntimeName = r.URL.Query().Get("runtimeName")
	cond.RuntimeID = r.URL.Query().Get("runtimeID")
	cond.ServiceName = r.URL.Query().Get("serviceName")
	cond.Workspace = r.URL.Query().Get("workspace")
	cond.ServiceType = r.URL.Query().Get("serviceType")
	cond.AddonID = r.URL.Query().Get("addonID")
	phases := r.URL.Query().Get("phases")
	if phases != "" {
		cond.Phases = strutil.Map(strutil.Split(phases, ",", true), strutil.ToLower, strutil.Title)
	}
	limitstr := r.URL.Query().Get("limit")
	if limitstr == "" {
		cond.Limit = 0
	} else {
		limit, err := strconv.Atoi(limitstr)
		if err != nil {
			errstr := fmt.Sprintf("failed to parse limit query condition: %v", err)
			logrus.Error(errstr)
			return mkResponse(apistructs.PodInfoResponse{
				Header: apistructs.Header{
					Success: false,
					Error: apistructs.ErrorResponse{
						Msg: errstr,
					},
				},
			})
		}
		cond.Limit = limit
	}
	pods, err := h.instanceinfoImpl.QueryPod(cond)
	if err != nil {
		errstr := fmt.Sprintf("failed to query instance info: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.PodInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}
	return mkResponse(apistructs.PodInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   pods,
	})
}

func (h *HTTPEndpoints) InstanceInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	cond := instanceinfo.QueryInstanceConditions{}
	cond.Cluster = r.URL.Query().Get("cluster")
	cond.OrgName = r.URL.Query().Get("orgName")
	cond.OrgID = r.URL.Query().Get("orgID")
	cond.ProjectName = r.URL.Query().Get("projectName")
	cond.ProjectID = r.URL.Query().Get("projectID")
	cond.ApplicationName = r.URL.Query().Get("applicationName")
	cond.EdgeApplicationName = r.URL.Query().Get("edgeApplicationName")
	cond.EdgeSite = r.URL.Query().Get("edgeSite")
	cond.ApplicationID = r.URL.Query().Get("applicationID")
	cond.RuntimeName = r.URL.Query().Get("runtimeName")
	cond.RuntimeID = r.URL.Query().Get("runtimeID")
	cond.ServiceName = r.URL.Query().Get("serviceName")
	cond.Workspace = r.URL.Query().Get("workspace")
	cond.ContainerID = r.URL.Query().Get("containerID")
	cond.InstanceIP = r.URL.Query().Get("instanceIP")
	cond.HostIP = r.URL.Query().Get("hostIP")
	cond.ServiceType = r.URL.Query().Get("serviceType")
	cond.AddonID = r.URL.Query().Get("addonID")
	phases := r.URL.Query().Get("phases")
	if phases != "" {
		cond.Phases = strutil.Map(strutil.Split(phases, ",", true), strutil.ToLower, strutil.Title)
	}
	limitstr := r.URL.Query().Get("limit")
	if limitstr == "" {
		cond.Limit = 0
	} else {
		limit, err := strconv.Atoi(limitstr)
		if err != nil {
			errstr := fmt.Sprintf("failed to parse limit query condition: %v", err)
			logrus.Error(errstr)
			return mkResponse(apistructs.InstanceInfoResponse{
				Header: apistructs.Header{
					Success: false,
					Error: apistructs.ErrorResponse{
						Msg: errstr,
					},
				},
			})
		}
		cond.Limit = limit
	}
	instances, err := h.instanceinfoImpl.QueryInstance(cond)
	if err != nil {
		errstr := fmt.Sprintf("failed to query instance info: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.InstanceInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}

	return mkResponse(apistructs.InstanceInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   instances,
	})
}

func (h *HTTPEndpoints) ServiceInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	cond := instanceinfo.QueryServiceConditions{}
	cond.OrgName = r.URL.Query().Get("orgName")
	cond.OrgID = r.URL.Query().Get("orgID")
	cond.ProjectName = r.URL.Query().Get("projectName")
	cond.ProjectID = r.URL.Query().Get("projectID")
	cond.ApplicationName = r.URL.Query().Get("applicationName")
	cond.ApplicationID = r.URL.Query().Get("applicationID")
	cond.RuntimeName = r.URL.Query().Get("runtimeName")
	cond.RuntimeID = r.URL.Query().Get("runtimeID")
	cond.ServiceName = r.URL.Query().Get("serviceName")
	cond.Workspace = r.URL.Query().Get("workspace")
	cond.ServiceType = r.URL.Query().Get("serviceType")
	services, err := h.instanceinfoImpl.QueryService(cond)
	if err != nil {
		errstr := fmt.Sprintf("failed to query service info: %v", err)
		logrus.Error(errstr)
		return mkResponse(apistructs.ServiceInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}
	return mkResponse(apistructs.ServiceInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   services,
	})
}

func (h *HTTPEndpoints) ComponentInfo(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	info, err := h.componentinfoImpl.Get()
	if err != nil {
		return mkResponse(apistructs.ComponentInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}
	return mkResponse(apistructs.ComponentInfoResponse{
		apistructs.Header{Success: true},
		info,
	})
}

// ClusterInfo 获取集群信息
func (h *HTTPEndpoints) ClusterInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	name := vars["clusterName"]
	if name == "" {
		errstr := fmt.Sprintf("empty cluster name")
		return mkResponse(apistructs.ClusterInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}

	info, err := h.clusterinfoImpl.Info(name)
	if err != nil {
		errstr := fmt.Sprintf("failed to get cluster info, clusterName: %s, (%v)", name, err)
		return mkResponse(apistructs.ClusterInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	return mkResponse(apistructs.ClusterInfoResponse{
		apistructs.Header{Success: true},
		info,
	})
}

func (h *HTTPEndpoints) ResourceInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	name := vars["clusterName"]
	if name == "" {
		errstr := fmt.Sprintf("empty cluster name")
		return mkResponse(apistructs.ClusterResourceInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}
	brief := false
	briefArg := r.URL.Query().Get("brief")
	if briefArg != "" {
		brief = true
	}
	data, err := h.resourceinfoImpl.Info(name, brief)
	if err != nil {
		errstr := fmt.Sprintf("failed to get resourceinfo: %v", err)
		return mkResponse(apistructs.ClusterResourceInfoResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			}})
	}

	return mkResponse(apistructs.ClusterResourceInfoResponse{
		Header: apistructs.Header{
			Success: true,
		},
		Data: data,
	})
}

// func (h *HTTPEndpoints) ServiceGroupStatusInfo(ctx context.Context, r *http.Request, vars map[string]string) (
// 	httpserver.Responser, error) {
// 	namespace := r.URL.Query().Get("namespace")
// 	name := r.URL.Query().Get("name")
// 	if namespace == "" || name == "" {
// 		errstr := fmt.Sprintf("empty namespace or name")
// 		return mkResponse(apistructs.ServiceGroupInfoResponse{
// 			Header: apistructs.Header{
// 				Success: false,
// 				Error:   apistructs.ErrorResponse{Msg: errstr},
// 			},
// 		})
// 	}
// 	h.instanceinfoImpl.QueryServiceGroup(namespace, name)
// }

func (h *HTTPEndpoints) Terminal(w http.ResponseWriter, r *http.Request) {
	terminal.Terminal(w, r)
}

func (h *HTTPEndpoints) CapacityInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	clustername := r.URL.Query().Get("clusterName")
	data := h.cap.CapacityInfo(clustername)
	return mkResponse(apistructs.CapacityInfoResponse{
		Header: apistructs.Header{Success: true},
		Data:   data,
	})
}

func (h *HTTPEndpoints) ServiceScaling(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	sg := apistructs.ServiceGroup{}
	err := json.NewDecoder(r.Body).Decode(&sg)
	if err != nil {
		return mkResponse(apistructs.ScheduleLabelSetResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: fmt.Sprintf("unmarshall to decoder the service err: %v", err),
				}},
		})
	}
	if _, err = h.serviceGroupImpl.Scale(&sg); err != nil {
		return mkResponse(apistructs.ScheduleLabelSetResponse{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: fmt.Sprintf("scale service %s error: %v", sg.Services[0].Name, err),
				}},
		})
	}
	return mkResponse(apistructs.ScheduleLabelSetResponse{
		Header: apistructs.Header{
			Success: true,
		},
	})
}

func mkResponse(content interface{}) (httpserver.Responser, error) {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: content,
	}, nil
}
