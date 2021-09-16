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
	"sort"
	"strings"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type Filter struct {
	base.DefaultProvider
	CtxBdl     *bundle.Bundle
	SDK        *cptype.SDK
	Type       string                 `json:"type"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
}

var DefaultLabels = []string{"dice/workspace-dev=true", "dice/workspace-test=true", "dice/workspace-staging=true",
	"dice/workspace-prod=true", "dice/stateful-service=true", "dice/stateless-service=true",
	"dice/location-cluster-service=true", "dice/job=true", "dice/bigdata-job=true", "dice/lb", "dice/platform"}

func (f *Filter) GetFilterProps(labels map[string]struct{}) Props {
	fields := []Field{
		{
			Key:      "organization",
			Multiple: true,
			Label:    f.SDK.I18n("organization-label"),
			Type:     "select",
			Options:  []Option{},
		},
	}
	var customs []string
	var enterprise []string
	for l := range labels {
		if strings.HasPrefix(l, "dice/org-") && strings.HasSuffix(l, "=true") {
			enterprise = append(enterprise, l)
			continue
		}
		exist := false
		for _, dl := range DefaultLabels {
			if dl == l {
				exist = true
				break
			}
		}
		if !exist {
			customs = append(customs, l)
		}
	}
	sort.Slice(enterprise, func(i, j int) bool {
		return enterprise[i] < enterprise[j]
	})
	for _, l := range enterprise {
		i := strings.Index(l, "=true")
		fields[0].Options = append(fields[0].Options, Option{
			Label: l[9:i],
			Value: l,
		})
	}
	sort.Slice(customs, func(i, j int) bool {
		return customs[i] < customs[j]
	})
	var customOps []Option
	for _, l := range customs {
		customOps = append(customOps, Option{
			Label: l,
			Value: l,
		})
	}
	fields = append(fields, []Field{
		{
			Key:      "env",
			Label:    f.SDK.I18n("env-label"),
			Multiple: true,
			Type:     "select",
			Options: []Option{
				{Label: f.SDK.I18n("dev"), Value: "dice/workspace-dev"},
				{Label: f.SDK.I18n("test"), Value: "dice/workspace-test=true"},
				{Label: f.SDK.I18n("staging"), Value: "dice/workspace-staging=true"},
				{Label: f.SDK.I18n("prod"), Value: "dice/workspace-prod=true"},
			},
		},
		{
			Key:      "service",
			Label:    f.SDK.I18n("service-label"),
			Multiple: true,
			Type:     "select",
			Options: []Option{
				{Label: f.SDK.I18n("stateful-service"), Value: "dice/stateful-service=true"},
				{Label: f.SDK.I18n("stateless-service"), Value: "dice/stateless-service=true"},
				{Label: f.SDK.I18n("location-cluster-service"), Value: "dice/location-cluster-service=true"},
			},
		},
		{
			Key:      "job-label",
			Label:    f.SDK.I18n("job-label"),
			Multiple: true,
			Type:     "select",
			Options: []Option{
				{Label: f.SDK.I18n("cicd-job"), Value: "dice/job=true"},
				{Label: f.SDK.I18n("bigdata-job"), Value: "dice/bigdata-job=true"},
			},
		},
		{
			Key:      "other-label",
			Label:    f.SDK.I18n("other-label"),
			Multiple: true,
			Type:     "dropdown-select",
			Options: append([]Option{
				{Label: f.SDK.I18n("lb"), Value: "dice/lb"},
				{Label: f.SDK.I18n("platform"), Value: "dice/platform"},
			}, customOps...),
		},
		{Key: "Q", Type: "input", Placeholder: f.SDK.I18n("input node Name or IP")},
	}...,
	)
	p := Props{
		LabelWidth: 40,
		Fields:     fields,
	}
	return p
}

type Values map[string]interface{}

type State struct {
	Values      Values `json:"values"`
	ClusterName string `json:"clusterName"`
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
	Multiple    bool     `json:"multiple"`
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
