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

package sheet_iteration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func Test_getOrderedIterationsTitles(t *testing.T) {
	i1 := &dao.Iteration{BaseModel: dbengine.BaseModel{ID: 0}, Title: "i1"}
	i2 := &dao.Iteration{BaseModel: dbengine.BaseModel{ID: 1}, Title: "i2"}
	i3 := &dao.Iteration{BaseModel: dbengine.BaseModel{ID: 2}, Title: "i3"}
	i4 := &dao.Iteration{BaseModel: dbengine.BaseModel{ID: 0}, Title: "i4"}

	m := map[string]*dao.Iteration{
		i1.Title: i1,
		i2.Title: i2,
		i3.Title: i3,
		i4.Title: i4,
	}

	order := getOrderedIterationsTitles(m)
	assert.Equal(t, []string{"i2", "i3", "i1", "i4"}, order)
}
