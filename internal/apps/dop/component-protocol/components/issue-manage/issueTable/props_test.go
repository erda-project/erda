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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func Test_buildTableColumnProps(t *testing.T) {
	ctx := context.WithValue(context.Background(), cptype.GlobalInnerKeyCtxSDK, &cptype.SDK{Tran: &MockTran{}})
	got := buildTableColumnProps(ctx, "REQUIREMENT", nil)
	c := got["columns"].([]interface{})
	assert.Equal(t, c[2].(Column).DataIndex, "progress")
	assert.Equal(t, c[1].(ColumnWithFixedWidth).Width, 500)
	got = buildTableColumnProps(ctx, "TASK", nil)
	c = got["columns"].([]interface{})
	assert.Equal(t, c[2].(Column).DataIndex, "complexity")
	got = buildTableColumnProps(ctx, "BUG", nil)
	c = got["columns"].([]interface{})
	assert.Equal(t, c[2].(Column).DataIndex, "severity")
	got = buildTableColumnProps(ctx, "ALL", nil)
	c = got["columns"].([]interface{})
	assert.Equal(t, c[2].(Column).DataIndex, "complexity")
}
