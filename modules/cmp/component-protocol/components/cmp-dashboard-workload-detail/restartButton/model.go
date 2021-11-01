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

package restartButton

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentRestartButton struct {
	base.DefaultProvider

	ctx        context.Context
	sdk        *cptype.SDK
	bdl        *bundle.Bundle
	server     cmp.SteveServer
	Type       string                 `json:"type"`
	State      State                  `json:"state"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName string `json:"clusterName,omitempty"`
	WorkloadID  string `json:"workloadId,omitempty"`
}

type Props struct {
	Type     string   `json:"type"`
	Text     string   `json:"text"`
	TipProps TipProps `json:"tipProps"`
}

type TipProps struct {
	Placement string `json:"placement,omitempty"`
}

type Operation struct {
	Key         string `json:"key,omitempty"`
	Reload      bool   `json:"reload"`
	Confirm     string `json:"confirm,omitempty"`
	Disabled    bool   `json:"disabled"`
	DisabledTip string `json:"disabledTip,omitempty"`
}
