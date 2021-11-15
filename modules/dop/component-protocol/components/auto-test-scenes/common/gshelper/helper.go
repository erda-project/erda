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
)

const (
	SceneConfigKey    string = "SceneConfigKey"
	SceneSetConfigKey string = "SceneSetConfigKey"
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

func (h *GSHelper) SetGlobalSelectedSetID(id uint64) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalSelectedSetID"] = id
}

func (h *GSHelper) GetGlobalSelectedSetID() uint64 {
	if h.gs == nil {
		return 0
	}
	v, ok := (*h.gs)["GlobalSelectedSetID"]
	if !ok {
		return 0
	}
	if _, ok = v.(uint64); ok {
		return v.(uint64)
	}
	return uint64(v.(float64))
}

func (h *GSHelper) SetGlobalActiveConfig(key string) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalActiveConfig"] = key
}

func (h *GSHelper) GetGlobalActiveConfig() string {
	if h.gs == nil {
		return ""
	}
	if v, ok := (*h.gs)["GlobalActiveConfig"].(string); ok {
		return v
	}
	return SceneSetConfigKey
}
