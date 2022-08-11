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

package issueFilter

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/standard-components/issueFilter"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/standard-components/issueFilter/gshelper"
)

type provider struct {
	issueFilter.IssueFilter
}

func init() {
	cpregister.RegisterComponent("issue-manage", "issueFilter", func() cptype.IComponent { return &provider{} })
}

func (p *provider) BeforeHandleOp(sdk *cptype.SDK) {
	p.Initial(sdk)
	p.State.WithStateCondition = true
	p.State.IssueRequestKey = gshelper.KeyIssuePagingRequest
}
