package filters

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/erda-project/erda/modules/eventbox/constant"
// 	"github.com/erda-project/erda/modules/eventbox/register"
// 	"github.com/erda-project/erda/modules/eventbox/types"

// 	"github.com/stretchr/testify/assert"
// )

// func TestRegisterFilter(t *testing.T) {
// 	r, err := register.New()
// 	assert.Nil(t, err)
// 	filter := NewRegisterFilter(r)
// 	m := types.Message{
// 		Sender:  "self",
// 		Content: "2333",
// 		Labels: map[types.LabelKey]interface{}{
// 			types.LabelKey(constant.RegisterLabelKey): []string{"aaa"},
// 			"other": "value",
// 		},
// 		Time: 0,
// 	}
// 	assert.Nil(t, r.Put("aaa", map[types.LabelKey]interface{}{
// 		"bbb": "1",
// 		"ccc": "2",
// 	}))
// 	derr := filter.Filter(&m)

// 	assert.True(t, derr.IsOK())
// 	if !derr.IsOK() {
// 		fmt.Printf("%+v\n", derr) // debug print

// 	}

// 	assert.Equal(t, "1", m.Labels["/bbb"])
// 	assert.Equal(t, "value", m.Labels["other"])
// 	assert.Equal(t, []string{"aaa"}, m.Labels[types.LabelKey(constant.RegisterLabelKey)])

// }
