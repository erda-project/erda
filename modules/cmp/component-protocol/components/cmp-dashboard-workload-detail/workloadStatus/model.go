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

package workloadStatus

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentWorkloadStatus struct {
	base.DefaultProvider
	sdk *cptype.SDK
	bdl *bundle.Bundle

	Type  string `json:"type,omitempty"`
	Data  Data   `json:"data,omitempty"`
	Props Props  `json:"props,omitempty"`
	State State  `json:"state,omitempty"`
}

type Props struct {
	Size string `json:"size,omitempty"`
}

type Data struct {
	Labels Labels `json:"labels,omitempty"`
}

type Labels struct {
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
}

type State struct {
	ClusterName string `json:"clusterName,omitempty"`
	WorkloadID  string `json:"workloadId,omitempty"`
}
