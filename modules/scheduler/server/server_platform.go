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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s *Server) epCreatePlatform(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	sg := apistructs.ServiceGroup{}
	if err := json.NewDecoder(r.Body).Decode(&sg); err != nil {
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.ServiceGroupUpdateResponse{
				Error: err.Error(),
			},
		}, nil
	}
	sg.CreatedTime = time.Now().Unix()
	sg.LastModifiedTime = sg.CreatedTime
	sg.Status = apistructs.StatusCreated

	if !validateRuntimeName(sg.ID) {
		return HTTPResponse{
			Status: http.StatusUnprocessableEntity,
			Content: apistructs.ServiceGroupCreateResponse{
				Name:  sg.ID,
				Error: "invalid Name",
			},
		}, nil
	}

	if err := s.store.Put(ctx, makePlatformKey(sg.ID), sg); err != nil {
		return nil, err
	}

	// add "MATCH_TAGS":"platform"
	sg.Labels[apistructs.LabelMatchTags] = apistructs.TagPlatform

	rt := apistructs.ServiceGroup(sg)
	if _, err := s.handleRuntime(ctx, &rt, task.TaskCreate); err != nil {
		return nil, err
	}

	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupCreateResponse{
			Name: sg.ID,
		},
	}, nil
}

func (s *Server) epDeletePlatform(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]

	sg := apistructs.ServiceGroup{}
	if err := s.store.Get(ctx, makePlatformKey(name), &sg); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupDeleteResponse{
				Error: fmt.Sprintf("Cannot get serviceGroup(%s) from etcd, err: %v", name, err),
			},
		}, nil
	}
	rt := apistructs.ServiceGroup(sg)
	if _, err := s.handleRuntime(ctx, &rt, task.TaskDestroy); err != nil {
		logrus.Errorf("delete platform component(%s) failed", name)
		return nil, err
	}

	if err := s.store.Remove(ctx, makePlatformKey(name), &sg); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupDeleteResponse{
			Name: name,
		},
	}, nil
}

func (s *Server) epGetPlatform(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	name := vars["name"]
	sg := apistructs.ServiceGroup{}

	if err := s.store.Get(ctx, makePlatformKey(name), &sg); err != nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%s) from etcd, err: %v", name, err),
			},
		}, nil
	}

	rt := apistructs.ServiceGroup(sg)
	result, err := s.handleRuntime(ctx, &rt, task.TaskInspect)
	if err != nil {
		return nil, err
	}

	if result.Extra == nil {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupGetErrorResponse{
				Error: fmt.Sprintf("Cannot get runtime(%v) info from TaskInspect", name),
			},
		}, nil
	}
	sg = *(result.Extra.(*apistructs.ServiceGroup))
	return HTTPResponse{
		Status:  http.StatusOK,
		Content: sg,
	}, nil

}

func (s *Server) epUpdatePlatform(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	sg := apistructs.ServiceGroup{}
	if err := json.NewDecoder(r.Body).Decode(&sg); err != nil {
		return HTTPResponse{
			Status: http.StatusBadRequest,
			Content: apistructs.ServiceGroupUpdateResponse{
				Error: err.Error(),
			},
		}, nil
	}

	name := vars["name"]
	if name != sg.ID {
		return HTTPResponse{
			Status: http.StatusNotFound,
			Content: apistructs.ServiceGroupUpdateResponse{
				Error: fmt.Sprintf("name not matched, http id(%s), req body name(%s)", name, sg.ID),
			},
		}, nil
	}

	// add "MATCH_TAGS":"platform"
	sg.Labels[apistructs.LabelMatchTags] = apistructs.TagPlatform
	sg.LastModifiedTime = time.Now().Unix()

	rt := apistructs.ServiceGroup(sg)
	if err := s.store.Put(ctx, makePlatformKey(name), rt); err != nil {
		return nil, err
	}

	if _, err := s.handleRuntime(ctx, &rt, task.TaskUpdate); err != nil {
		return nil, err
	}
	return HTTPResponse{
		Status: http.StatusOK,
		Content: apistructs.ServiceGroupUpdateResponse{
			Name: name,
		},
	}, nil

	return nil, nil
}

func (s *Server) epRestartPlatform(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
	return nil, nil
}
