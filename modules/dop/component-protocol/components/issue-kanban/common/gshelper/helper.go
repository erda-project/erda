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

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
)

const (
	keyIssuePagingRequestKanBan = "IssuePagingRequestKanBan"
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

func (h *GSHelper) SetIssuePagingRequest(req apistructs.IssuePagingRequest) {
	if h.gs == nil {
		return
	}
	(*h.gs)[keyIssuePagingRequestKanBan] = req
}

func (h *GSHelper) GetIssuePagingRequest() (*apistructs.IssuePagingRequest, bool) {
	if h.gs == nil {
		return nil, false
	}
	v, ok := (*h.gs)[keyIssuePagingRequestKanBan]
	if !ok {
		return nil, false
	}
	var req apistructs.IssuePagingRequest
	cputil.MustObjJSONTransfer(v, &req)
	return &req, true
}
