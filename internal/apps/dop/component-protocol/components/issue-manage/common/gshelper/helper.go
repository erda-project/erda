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

package gshelper

import (
	"github.com/mitchellh/mapstructure"

	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/standard-components/issueFilter/gshelper"
)

const (
	keyIssuePagingRequest = "IssuePagingRequest"
)

type GSHelper struct {
	gs *cptype.GlobalStateData
}

func NewGSHelper(gs *cptype.GlobalStateData) *GSHelper {
	return &GSHelper{gs: gs}
}

func assign(src, dst interface{}) error {
	if src == nil || dst == nil {
		return nil
	}

	return mapstructure.Decode(src, dst)
}

func (h *GSHelper) GetIssuePagingRequest() (*pb.PagingIssueRequest, bool) {
	if h.gs == nil {
		return nil, false
	}
	v, ok := (*h.gs)[keyIssuePagingRequest]
	if !ok {
		return nil, false
	}
	var req pb.PagingIssueRequest
	cputil.MustObjJSONTransfer(v, &req)
	return &req, true
}

func (h *GSHelper) GetIterationOptions() ([]model.SelectOption, bool) {
	if h.gs == nil {
		return nil, false
	}
	v, ok := (*h.gs)[gshelper.KeyIterationOptions]
	if !ok {
		return nil, false
	}
	var res []model.SelectOption
	cputil.MustObjJSONTransfer(v, &res)
	return res, true
}
