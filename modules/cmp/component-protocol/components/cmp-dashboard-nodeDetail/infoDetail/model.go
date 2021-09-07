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

package infoDetail

import (
	"context"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type InfoDetail struct {
	CtxBdl *bundle.Bundle
	base.DefaultProvider
	SDK   *cptype.SDK
	Ctx   context.Context
	Type  string          `json:"type"`
	Props Props           `json:"props"`
	Data  map[string]Data `json:"data"`
}

type Data struct {
	Survive          string  `json:"survive"`
	NodeIp           string  `json:"nodeIp"`
	Version          string  `json:"version"`
	Os               string  `json:"os"`
	ContainerRuntime string  `json:"containerRuntime"`
	PodNum           string  `json:"podNum"`
	Tags             []Field `json:"tag"`
	Desc             []Field `json:"desc"`
}

type Props struct {
	ColumnNum int     `json:"columnNum"`
	Fields    []Field `json:"fields"`
}

type Field struct {
	Label      string               `json:"label"`
	Group      string               `json:"group"`
	ValueKey   string               `json:"valueKey"`
	RenderType string               `json:"renderType"`
	SpaceNum   int                  `json:"spaceNum"`
	Operations map[string]Operation `json:"operations"`
}

type Operation struct {
	Key           string                 `json:"key"`
	Reload        bool                   `json:"reload"`
	FillMeta      string                 `json:"fillMeta,omitempty"`
	Target        string                 `json:"target,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
	ClickableKeys interface{}            `json:"clickableKeys,omitempty"`
	Text          string                 `json:"text"`
	Command       Command                `json:"command,omitempty"`
}

type Command struct {
	Key          string `json:"key"`
	Target       string `json:"target"`
	JumpOut      bool
	CommandState CommandState `json:"state"`
}

type CommandState struct {
	Visible  bool     `json:"visible"`
	FormData FormData `json:"formData"`
}

type FormData struct {
	RecordId string `json:"recordId"`
}
