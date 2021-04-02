package filters

import (
	"testing"

	"github.com/erda-project/erda/modules/eventbox/types"

	"github.com/stretchr/testify/assert"
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
