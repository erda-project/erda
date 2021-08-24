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

package register

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	stypes "github.com/erda-project/erda/modules/eventbox/server/types"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type PutRequest struct {
	Key    string                         `json:"key"`
	Labels map[types.LabelKey]interface{} `json:"labels"`
}

type DelRequest struct {
	Key string `json:"key"`
}

type GetResponseContent map[types.LabelKey]interface{}

type RegisterHTTP struct {
	register Register
}

func NewHTTP(register Register) *RegisterHTTP {
	return &RegisterHTTP{
		register: register,
	}
}

func (r *RegisterHTTP) Put(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	var m PutRequest
	if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
		logrus.Errorf("RegisterHTTP Put: %v", err)
		return stypes.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: "unmarshal message failed",
		}, err
	}
	if err := r.register.Put(m.Key, m.Labels); err != nil {
		err := errors.Errorf("RegisterHTTP Put: %v", err)
		logrus.Error(err)
		return stypes.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Content: err.Error(),
		}, err
	}
	return stypes.HTTPResponse{
		Status:  http.StatusOK,
		Content: "",
	}, nil
}

func (r *RegisterHTTP) PrefixGet(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	key := req.URL.Query().Get("key")
	if key == "" {
		logrus.Infof("RegisterHTTP Get: request not provide key")
		return stypes.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: "request not provide key",
		}, nil
	}
	labels := r.register.PrefixGet(key)
	if labels == nil {
		logrus.Infof("RegisterHTTP Get (not found): %v", key)
		return stypes.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: "",
		}, nil
	}
	return stypes.HTTPResponse{
		Status:  http.StatusOK,
		Content: labels,
	}, nil
}

func (r *RegisterHTTP) Del(ctx context.Context, req *http.Request, vars map[string]string) (stypes.Responser, error) {
	var m DelRequest
	if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
		logrus.Errorf("RegisterHTTP Del: %v", err)
		return stypes.HTTPResponse{
			Status:  http.StatusBadRequest,
			Content: "unmarshal message failed",
		}, err
	}
	if err := r.register.Del(m.Key); err != nil {
		return stypes.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Content: err.Error(),
		}, err
	}
	return stypes.HTTPResponse{
		Status:  http.StatusOK,
		Content: "",
	}, nil
}

func (r *RegisterHTTP) GetHTTPEndPoints() []stypes.Endpoint {
	return []stypes.Endpoint{
		{"/register", http.MethodGet, r.PrefixGet},
		{"/register", http.MethodPut, r.Put},
		{"/register", http.MethodDelete, r.Del},
	}
}
