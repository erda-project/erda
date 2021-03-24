package conf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyToken(t *testing.T) {
	token, err := applyOpenAPIToken("openapi.marathon.l4lb.thisdcos.directory:9529")
	require.NoError(t, err)
	fmt.Println(token)
}
