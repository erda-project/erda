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

package issueTable

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
)

type Column struct {
	Title     string `json:"title,omitempty"`
	DataIndex string `json:"dataIndex,omitempty"`
	Hidden    bool   `json:"hidden"`
}

type ColumnWithFixedWidth struct {
	Column
	Width int `json:"width,omitempty"`
}

func buildTableColumnProps(ctx context.Context, issueType string, customProperties []*pb.IssuePropertyExtraProperty) cptype.ComponentProps {
	id := Column{"ID", "id", false}
	name := ColumnWithFixedWidth{Column{cputil.I18n(ctx, "title"), "name", false}, 500}
	progress := Column{cputil.I18n(ctx, "progress"), "progress", false}
	severity := Column{cputil.I18n(ctx, "severity"), "severity", false}
	complexity := Column{cputil.I18n(ctx, "complexity"), "complexity", true}
	priority := Column{cputil.I18n(ctx, "priority"), "priority", false}
	state := Column{cputil.I18n(ctx, "state"), "state", false}
	assignee := Column{cputil.I18n(ctx, "assignee"), "assignee", false}
	deadline := Column{cputil.I18n(ctx, "deadline"), "deadline", false}
	closedAt := Column{cputil.I18n(ctx, "closed-at"), "closedAt", true}
	reopenCount := Column{cputil.I18n(ctx, "reopenCount"), "reopenCount", true}
	createdAt := Column{cputil.I18n(ctx, "created-at"), "createdAt", true}

	iteration := Column{cputil.I18n(ctx, "iteration"), "iteration", true}
	planStartedAt := Column{cputil.I18n(ctx, "started-at"), "planStartedAt", true}
	creator := Column{cputil.I18n(ctx, "creator"), "creator", true}
	owner := Column{cputil.I18n(ctx, "responsible-person"), "owner", true}

	var columns []interface{}
	switch issueType {
	case pb.IssueTypeEnum_REQUIREMENT.String():
		columns = []interface{}{id, name, progress, complexity, priority, iteration, state, assignee, planStartedAt, deadline, creator, createdAt}
	case pb.IssueTypeEnum_TASK.String():
		columns = []interface{}{id, name, complexity, priority, iteration, state, assignee, planStartedAt, deadline, creator, createdAt}
	case pb.IssueTypeEnum_BUG.String():
		columns = []interface{}{id, name, severity, complexity, priority, iteration, state, reopenCount, assignee, planStartedAt, deadline, owner, closedAt, creator, createdAt}
	case pb.IssueTypeEnum_TICKET.String():
		createdAt.Hidden = false
		columns = []interface{}{id, name, severity, complexity, priority, state, assignee, planStartedAt, deadline, creator, createdAt}
	default:
		columns = []interface{}{id, name, complexity, priority, iteration, state, assignee, planStartedAt, deadline, creator, createdAt}
	}

	// handle custom issue properties
	for _, property := range customProperties {
		columns = append(columns, Column{property.DisplayName, "property_" + property.PropertyName, true})
	}

	return map[string]interface{}{
		"columns":         columns,
		"rowKey":          "id",
		"pageSizeOptions": []string{"10", "20", "50", "100"},
		"batchOperations": []string{"delete"},
		"selectable":      true,
	}
}
