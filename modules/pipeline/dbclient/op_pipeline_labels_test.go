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

package dbclient

//import (
//	"testing"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestClient_GetPipelineIDsMatchLabels(t *testing.T) {
//	ids, err := client.SelectPipelineIDsByLabels(apistructs.PipelineIDSelectByLabelRequest{
//		AnyMatchLabels: map[string][]string{"k1": []string{"v1"}, "k2": {"v2"}, "k3": {"v3"}},
//	})
//	assert.NoError(t, err)
//	spew.Dump(ids)
//}
//
//func TestFilter(t *testing.T) {
//	ss := filter(
//		[]uint64{1, 2},
//		[]uint64{2},
//		[]uint64{},
//	)
//	assert.True(t, len(ss) == 0)
//
//	ss = filter(
//		[]uint64{1, 2, 3},
//		[]uint64{4, 5, 6},
//		[]uint64{1},
//	)
//	assert.True(t, len(ss) == 0)
//
//	ss = filter(
//		[]uint64{1, 2, 3, 4},
//		[]uint64{4, 5, 6},
//		[]uint64{1, 4},
//	)
//	assert.True(t, len(ss) == 1)
//	assert.True(t, ss[0] == 4)
//
//	ss = filter()
//	assert.True(t, len(ss) == 0)
//}
//
//func TestFilterAndOrder(t *testing.T) {
//	ss := filterAndOrder(
//		[]uint64{1, 2},
//		[]uint64{2},
//		[]uint64{},
//	)
//	assert.True(t, len(ss) == 0)
//
//	ss = filterAndOrder(
//		[]uint64{1, 2, 3},
//		[]uint64{4, 5, 6},
//		[]uint64{1},
//	)
//	assert.True(t, len(ss) == 0)
//
//	ss = filterAndOrder(
//		[]uint64{1, 2, 3, 4},
//		[]uint64{4, 5, 6},
//		[]uint64{1, 4},
//	)
//	assert.True(t, len(ss) == 1)
//	assert.True(t, ss[0] == 4)
//
//	ss = filterAndOrder()
//	assert.True(t, len(ss) == 0)
//
//	ss = filterAndOrder(
//		[]uint64{1, 2, 3, 4, 5, 6},
//		[]uint64{6, 5, 4, 2},
//		[]uint64{3, 4, 8, 7, 9, 6, 2},
//	)
//	assert.True(t, len(ss) == 3)
//	assert.Equal(t, []uint64{2, 4, 6}, ss)
//}
