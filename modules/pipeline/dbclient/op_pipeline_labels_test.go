package dbclient

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestClient_GetPipelineIDsMatchLabels(t *testing.T) {
	ids, err := client.SelectPipelineIDsByLabels(apistructs.PipelineIDSelectByLabelRequest{
		AnyMatchLabels: map[string][]string{"k1": []string{"v1"}, "k2": {"v2"}, "k3": {"v3"}},
	})
	assert.NoError(t, err)
	spew.Dump(ids)
}

func TestFilter(t *testing.T) {
	ss := filter(
		[]uint64{1, 2},
		[]uint64{2},
		[]uint64{},
	)
	assert.True(t, len(ss) == 0)

	ss = filter(
		[]uint64{1, 2, 3},
		[]uint64{4, 5, 6},
		[]uint64{1},
	)
	assert.True(t, len(ss) == 0)

	ss = filter(
		[]uint64{1, 2, 3, 4},
		[]uint64{4, 5, 6},
		[]uint64{1, 4},
	)
	assert.True(t, len(ss) == 1)
	assert.True(t, ss[0] == 4)

	ss = filter()
	assert.True(t, len(ss) == 0)
}

func TestFilterAndOrder(t *testing.T) {
	ss := filterAndOrder(
		[]uint64{1, 2},
		[]uint64{2},
		[]uint64{},
	)
	assert.True(t, len(ss) == 0)

	ss = filterAndOrder(
		[]uint64{1, 2, 3},
		[]uint64{4, 5, 6},
		[]uint64{1},
	)
	assert.True(t, len(ss) == 0)

	ss = filterAndOrder(
		[]uint64{1, 2, 3, 4},
		[]uint64{4, 5, 6},
		[]uint64{1, 4},
	)
	assert.True(t, len(ss) == 1)
	assert.True(t, ss[0] == 4)

	ss = filterAndOrder()
	assert.True(t, len(ss) == 0)

	ss = filterAndOrder(
		[]uint64{1, 2, 3, 4, 5, 6},
		[]uint64{6, 5, 4, 2},
		[]uint64{3, 4, 8, 7, 9, 6, 2},
	)
	assert.True(t, len(ss) == 3)
	assert.Equal(t, []uint64{2, 4, 6}, ss)
}
