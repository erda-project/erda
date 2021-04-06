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

// TODO: refactor me, 把这个文件中的所有逻辑都去掉，只有 http 处理和 检查 参数合法性
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s *Server) epCreateRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	create := apistructs.ServiceGroupCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		return HTTPResponse{Status: http.StatusBadRequest}, err
	}
	runtime := apistructs.ServiceGroup(create)
	runtime.CreatedTime = time.Now().Unix()
	runtime.LastModifiedTime = runtime.CreatedTime
	if runtime.Type == "" {
		runtime.Type = "default"
	}
	runtime.Status = apistructs.StatusCreated

	// TODO: Authentication, authorization, admission control
	if !validateRuntimeName(runtime.ID) || !validateRuntimeNamespace(runtime.Type) {
		return HTTPResponse{
			Status: http.StatusUnprocessableEntity,
			Content: apistructs.ServiceGroupCreateResponse{
				Name:  runtime.ID,
				Error: "invalid Name or Namespace",
			},
		}, nil
	}

	if err := clusterutil.SetRuntimeExecutorByCluster(&runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusUnprocessableEntity,
			Content: apistructs.ServiceGroupCreateResponse{
				Name:  runtime.ID,
				Error: err.Error(),
			},
		}, nil
	}

	if err := s.store.Put(ctx, makeRuntimeKey(runtime.Type, runtime.ID), runtime); err != nil {
		return nil, err
	}

	// build match tags and exclude tags
	runtime.Labels = appendServiceTags(runtime.Labels, runtime.Executor)

	if _, err := s.handleRuntime(ctx, &runtime, task.TaskCreate); err != nil {
		return nil, err
	}

	// 不轮询最新状态，etcd只存储相应元数据信息及期望值
	// runtime最新状态是希望上层通过GET接口来查询

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupCreateResponse{
			Name: runtime.ID,
		},
	}, nil

	//version := "TODO"
	// TODO: update version to etcd if version needed
}

func (s *Server) epUpdateRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	// 根据全量更新来实现
	// upRuntime 从上层request body解析出，反映上层期望的runtime
	upRuntime := apistructs.ServiceGroup{}
	if err := json.NewDecoder(r.Body).Decode(&upRuntime); err != nil {
		return HTTPResponse{Status: http.StatusBadRequest}, err
	}

	name := vars["name"]
	namespace := vars["namespace"]

	if len(upRuntime.Services) == 0 || name != upRuntime.ID || namespace != upRuntime.Type {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupUpdateResponse{
				Error: fmt.Sprintf("update request body services empty, or namespace/name not matched"),
			},
		}, nil
	}

	// oldRuntime 从etcd中获取，反映当前runtime
	oldRuntime := apistructs.ServiceGroup{}
	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &oldRuntime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupUpdateResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	// 清理 cache, 会触发一次新的部署
	s.l.Lock()
	_, ok := s.runtimeMap[oldRuntime.Type+"--"+oldRuntime.ID]
	if ok {
		delete(s.runtimeMap, oldRuntime.Type+"--"+oldRuntime.ID)
	}
	s.l.Unlock()

	diffAndPatchRuntime(&upRuntime, &oldRuntime)

	if err := s.store.Put(ctx, makeRuntimeKey(oldRuntime.Type, oldRuntime.ID), oldRuntime); err != nil {
		return nil, err
	}

	// build match tags and exclude tags
	oldRuntime.Labels = appendServiceTags(oldRuntime.Labels, oldRuntime.Executor)

	if _, err := s.handleRuntime(ctx, &oldRuntime, task.TaskUpdate); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupUpdateResponse{
			Name: oldRuntime.ID,
		},
	}, nil
}

// 暂时约定上层通过epUpdateRuntime来实现
func (s *Server) epUpdateService(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	// TODO
	return nil, nil
}

func (s *Server) epRestartRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	// imagine the request's body is empty
	name := vars["name"]
	namespace := vars["namespace"]
	runtime := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupRestartResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	// 清理 cache, 会触发一次新的部署
	s.l.Lock()
	_, ok := s.runtimeMap[runtime.Type+"--"+runtime.ID]
	if ok {
		delete(s.runtimeMap, runtime.Type+"--"+runtime.ID)
	}
	s.l.Unlock()

	if runtime.Extra == nil {
		runtime.Extra = make(map[string]string)
	}
	runtime.Extra[LastRestartTime] = time.Now().String()
	runtime.LastModifiedTime = time.Now().Unix()

	if err := s.store.Put(ctx, makeRuntimeKey(runtime.Type, runtime.ID), runtime); err != nil {
		return nil, err
	}

	// build match tags and exclude tags
	runtime.Labels = appendServiceTags(runtime.Labels, runtime.Executor)

	if _, err := s.handleRuntime(ctx, &runtime, task.TaskUpdate); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupRestartResponse{
			Name: runtime.ID,
		},
	}, nil
}

func (s *Server) epCancelAction(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]
	runtime := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	result, err := s.handleRuntime(ctx, &runtime, task.TaskCancel)
	if err != nil {
		return nil, err
	}

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: result.Extra,
	}, nil
}

func (s *Server) epDeleteRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]

	runtime := apistructs.ServiceGroup{}
	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupDeleteResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}
	if _, err := s.handleRuntime(ctx, &runtime, task.TaskDestroy); err != nil {
		return nil, err
	}
	if err := s.store.Remove(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupDeleteResponse{
			Name: name,
		},
	}, nil
}

func (s *Server) epGetRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]
	runtime := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	s.l.Lock()
	run, ok := s.runtimeMap[runtime.Type+"--"+runtime.ID]
	s.l.Unlock()
	if ok {
		return HTTPResponse{
			Status:  http.StatusOK,
			Content: *run,
		}, nil
	}

	result, err := s.handleRuntime(ctx, &runtime, task.TaskInspect)
	if err != nil {
		return nil, err
	}

	if result.Extra == nil {
		err = errors.Errorf("Cannot get runtime(%v/%v) info from TaskInspect", runtime.Type, runtime.ID)
		logrus.Errorf(err.Error())

		// return empty response for testing
		return HTTPResponse{
			Status:  http.StatusOK,
			Content: apistructs.ServiceGroup{},
		}, nil
	}
	newRuntime := result.Extra.(*apistructs.ServiceGroup)

	s.l.Lock()
	s.runtimeMap[runtime.Type+"--"+runtime.ID] = newRuntime
	s.l.Unlock()

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: *newRuntime,
	}, nil
}

func (s *Server) epGetRuntimeNoCache(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]
	runtime := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	result, err := s.handleRuntime(ctx, &runtime, task.TaskInspect)
	if err != nil {
		return nil, err
	}

	if result.Extra == nil {
		err = errors.Errorf("Cannot get runtime(%v/%v) info from TaskInspect", runtime.Type, runtime.ID)
		logrus.Errorf(err.Error())

		// return empty response for testing
		return HTTPResponse{
			Status:  http.StatusOK,
			Content: apistructs.ServiceGroup{},
		}, nil
	}
	newRuntime := result.Extra.(*apistructs.ServiceGroup)

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: *newRuntime,
	}, nil
}

func (s *Server) handleRuntime(ctx context.Context, runtime *apistructs.ServiceGroup, taskAction task.Action) (task.TaskResponse, error) {
	var result task.TaskResponse

	if err := clusterutil.SetRuntimeExecutorByCluster(runtime); err != nil {
		return result, err
	}

	// put the task in scheduler's buffered channel
	// handle the task in schduler's loop
	task, err := s.sched.Send(ctx, task.TaskRequest{
		ExecutorKind: getServiceExecutorKindByName(runtime.Executor),
		ExecutorName: runtime.Executor,
		Action:       taskAction,
		ID:           runtime.ID,
		Spec:         *runtime,
	})
	if err != nil {
		return result, err
	}

	// get response from task's channel
	if result = task.Wait(ctx); result.Err() != nil {
		return result, result.Err()
	}

	return result, nil
}

// upRuntime may not be complete
func diffAndPatchRuntime(upRuntime *apistructs.ServiceGroup, oldRuntime *apistructs.ServiceGroup) {
	// generate LastModifiedTime according to current time
	oldRuntime.LastModifiedTime = time.Now().Unix()

	oldRuntime.Labels = upRuntime.Labels
	oldRuntime.ServiceDiscoveryKind = upRuntime.ServiceDiscoveryKind

	// TODO: refactor it, separate data and status into different etcd key
	// 全量更新
	oldRuntime.Services = upRuntime.Services
}

func appendServiceTags(labels map[string]string, executor string) map[string]string {
	matchTags := make([]string, 0)
	if labels["SERVICE_TYPE"] == "STATELESS" {
		matchTags = append(matchTags, apistructs.TagServiceStateless)
	} else if executor == "MARATHONFORTERMINUSY" && labels["SERVICE_TYPE"] == "ADDONS" {
		matchTags = append(matchTags, apistructs.TagServiceStateful)
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[apistructs.LabelMatchTags] = strings.Join(matchTags, ",")
	labels[apistructs.LabelExcludeTags] = apistructs.TagLocked + "," + apistructs.TagPlatform
	return labels
}

// to suppress the error, to be the same with origin semantic
func getServiceExecutorKindByName(name string) string {
	e, err := executor.GetManager().Get(executortypes.Name(name))
	if err != nil {
		return conf.DefaultRuntimeExecutor()
	}
	return string(e.Kind())
}

func (s *Server) epPrefixGetRuntime(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	// e.g. "94-test-2478"
	prefix := vars["prefix"]
	var prefixKey string

	// TODO: FIX ME. 默认用services这个namespace
	if !strings.HasPrefix(prefix, "/dice/service/") {
		prefixKey = "/dice/service/services/" + prefix
	}

	return s.prefixGetRuntime(ctx, prefixKey)
}

func (s *Server) epPrefixGetRuntimeWithNamespace(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	// e.g. "services/94-test-2478"
	namespace := vars["namespace"]
	prefix := vars["prefix"]
	var prefixKey string

	if !strings.HasPrefix(prefix, "/dice/service/") {
		prefixKey = "/dice/service/" + namespace + "/" + prefix
	}

	return s.prefixGetRuntime(ctx, prefixKey)
}

func (s *Server) prefixGetRuntime(ctx context.Context, prefixKey string) (Responser, error) {
	kvs := make(map[string]interface{})
	runtime := &apistructs.ServiceGroup{}

	saveKV := func(k string, v interface{}) error {
		kvs[k] = v
		return nil
	}

	if err := s.store.ForEach(ctx, prefixKey, apistructs.ServiceGroup{}, saveKV); err != nil {
		return HTTPResponse{
			Status: http.StatusInternalServerError,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get prefix(%s) from etcd, err: %v", prefixKey, err),
			},
		}, err
	}

	logrus.Debugf("prefixKey(%s) get kvs len: %v", prefixKey, len(kvs))

	if len(kvs) == 0 {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("not found for prefix(%s) from etcd", prefixKey),
			},
		}, errors.Errorf("not found for prefix(%s) from etcd", prefixKey)
	}

	if len(kvs) > 1 {
		var keys []string
		for k := range kvs {
			keys = append(keys, k)
		}
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("more than one key matched for prefix(%s) from etcd, matched keys: %v", prefixKey, keys),
			},
		}, errors.Errorf("more than one key matched for prefix(%s) from etcd, matched keys: %v", prefixKey, keys)
	}

	// only one pair
	for _, v := range kvs {
		runtime = v.(*apistructs.ServiceGroup)
	}

	//logrus.Debugf("get runtime: %+v", runtime)
	result, err := s.handleRuntime(ctx, runtime, task.TaskInspect)
	if err != nil {
		return nil, err
	}

	if result.Extra == nil {
		err = errors.Errorf("get EMPTY runtimeinfo(%v/%v) from TaskInspect", runtime.Type, runtime.ID)
		logrus.Errorf(err.Error())

		return HTTPResponse{
			Status:  http.StatusOK,
			Content: apistructs.ServiceGroup{},
		}, err
	}
	newRuntime := result.Extra.(*apistructs.ServiceGroup)

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: *newRuntime,
	}, nil
}

func (s *Server) epGetRuntimeStatus(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	namespace := vars["namespace"]
	runtime := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makeRuntimeKey(namespace, name), &runtime); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s/%s) from etcd, err: %v", namespace, name, err),
			},
		}, nil
	}

	result, err := s.handleRuntime(ctx, &runtime, task.TaskInspect)
	if err != nil {
		return nil, err
	}

	multiStatus := apistructs.MultiLevelStatus{
		Namespace: namespace,
		Name:      name,
	}

	// 返回空的runtime
	if result.Extra == nil {
		logrus.Errorf("got runtime(%v/%v) empty, executor: %s", runtime.Type, runtime.ID, runtime.Executor)
		return nil, errors.Errorf("got runtime(%v/%v) but found it empty", runtime.Type, runtime.ID)
	}

	newRuntime := result.Extra.(*apistructs.ServiceGroup)
	multiStatus.Status = convertServiceStatus(newRuntime.Status)
	multiStatus.More = make(map[string]string)
	for _, service := range newRuntime.Services {
		multiStatus.More[service.Name] = convertServiceStatus(service.Status)
	}

	return HTTPResponse{
		Status:  http.StatusOK,
		Content: multiStatus,
	}, nil
}

func convertServiceStatus(serviceStatus apistructs.StatusCode) string {
	switch serviceStatus {
	case apistructs.StatusReady:
		return string(apistructs.StatusHealthy)

	case apistructs.StatusProgressing:
		return string(apistructs.StatusUnHealthy)

	default:
		return string(apistructs.StatusUnknown)
	}
}
