package dbclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetPipelineStatus(t *testing.T) {
	status, err := client.GetPipelineStatus(10001654)
	assert.NoError(t, err)
	fmt.Println(status)
}
