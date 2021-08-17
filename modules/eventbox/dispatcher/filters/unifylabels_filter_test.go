// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/eventbox/types"
)

func TestUnifyLabels(t *testing.T) {
	// defer elapsed("page")()
	f := NewUnifyLabelsFilter()
	m := types.Message{Labels: map[types.LabelKey]interface{}{
		"aaa":  "bbb",
		"/ccc": "dede",
	}}
	assert.Nil(t, f.Filter(&m))
	assert.NotNil(t, m.Labels[types.LabelKey("/aaa")])
}

// func elapsed(what string) func() {
// 	start := time.Now()
// 	return func() {
// 		fmt.Printf("%s took %v\n", what, time.Since(start))
// 	}
// }
