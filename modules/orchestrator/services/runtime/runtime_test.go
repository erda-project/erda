// Package runtime 应用实例相关操作
package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestModifyStatusIfNotForDisplay(t *testing.T) {
	runtime := apistructs.RuntimeInspectDTO{
		Status: "Unknown",
		Services: map[string]*apistructs.RuntimeInspectServiceDTO{
			"test": {
				Status: "Stopped",
			},
		},
	}
	updateStatusToDisplay(&runtime)
	assert.Equal(t, "Unknown", runtime.Status)
	for _, s := range runtime.Services {
		assert.Equal(t, "Stopped", s.Status)
	}
}
