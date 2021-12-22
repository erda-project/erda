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
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
)

const (
	KeyWorkbenchItemType = "WorkbenchItemType"
	KeyFilterName        = "FilterName"
)

type GSHelper struct {
	gs *cptype.GlobalStateData
}

func NewGSHelper(gs *cptype.GlobalStateData) *GSHelper {
	return &GSHelper{gs: gs}
}

func (h *GSHelper) SetWorkbenchItemType(wbType apistructs.WorkbenchItemType) {
	if h.gs == nil {
		return
	}
	(*h.gs)[KeyWorkbenchItemType] = wbType
}

func (h *GSHelper) GetWorkbenchItemType() (apistructs.WorkbenchItemType, bool) {
	if h.gs == nil {
		return "", false
	}
	v, ok := (*h.gs)[KeyWorkbenchItemType]
	if !ok {
		return "", false
	}
	var wbType apistructs.WorkbenchItemType
	cputil.MustObjJSONTransfer(v, &wbType)
	return wbType, true
}

func (h *GSHelper) SetFilterName(name string) {
	if h.gs == nil {
		return
	}
	(*h.gs)[KeyFilterName] = name
}

func (h *GSHelper) GetFilterName() (string, bool) {
	if h.gs == nil {
		return "", false
	}
	v, ok := (*h.gs)[KeyFilterName]
	if !ok {
		return "", false
	}
	var name string
	cputil.MustObjJSONTransfer(v, &name)
	return name, true
}
