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

package filter

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (f *Filter) GetGroups() map[string]struct{} {
	props := f.GetFilterProps()
	groups := make(map[string]struct{})
	for _, f := range props.Fields {
		for _, l := range f.Options {
			groups[l.Value] = struct{}{}
		}
	}
	return groups
}

type Filter struct {
	base.DefaultProvider
	CtxBdl     *bundle.Bundle
	SDK        *cptype.SDK
	Type       string                 `json:"type"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
}

func (f *Filter) GetFilterProps() Props {
	p := Props{
		LabelWidth: 40,
		Fields: []Field{
			{
				Key:   "env",
				Label: f.SDK.I18n("env"),
				Type:  "select",
				Options: []Option{
					{Label: f.SDK.I18n("dev"), Value: "dev"},
					{Label: f.SDK.I18n("test"), Value: "test"},
					{Label: f.SDK.I18n("staging"), Value: "staging"},
					{Label: f.SDK.I18n("prod"), Value: "prod"},
				},
			},
			{
				Key:   "service",
				Label: f.SDK.I18n("service"),
				Type:  "select",
				Options: []Option{
					{Label: f.SDK.I18n("stateful"), Value: "stateful"},
					{Label: f.SDK.I18n("stateless"), Value: "stateless"},
				},
			},
			{
				Key:   "packjob",
				Label: f.SDK.I18n("packjob"),
				Type:  "select",
				Options: []Option{
					{Label: "pack-job", Value: "packJob"},
				},
			},
			{
				Key:   "other",
				Label: f.SDK.I18n("other"),
				Type:  "select",
				Options: []Option{
					{Label: f.SDK.I18n("cluster-service"), Value: "cluster-service"},
					{Label: f.SDK.I18n("custom"), Value: "custom"},
					{Label: f.SDK.I18n("mono"), Value: "mono"},
					{Label: f.SDK.I18n("cordon"), Value: "cordon"},
					{Label: f.SDK.I18n("drain"), Value: "drain"},
					{Label: f.SDK.I18n("platform"), Value: "platform"},
				},
			},
			{Key: "Q", Type: "input", Placeholder: "请输入"},
		},
	}
	return p
}

type Values map[string]string
type State struct {
	Values Values `json:"values"`
}

type Props struct {
	LabelWidth int     `json:"labelWidth"`
	Fields     []Field `json:"fields"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload,omitempty"`
}

type Field struct {
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Options     []Option `json:"options"`
	Key         string   `json:"key"`
	Placeholder string   `json:"placeholder"`
}

type Option struct {
	Label    string   `json:"label"`
	Value    string   `json:"value"`
	Children []Option `json:"children"`
}
