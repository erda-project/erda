package settings

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {
	ttl := ttl{}
	ttl.TTL = 7
	ttl.HotTTL = 3

	byt, err := json.Marshal(&ttl)
	require.Equal(t, `{"ttl":"168h0m0s","hot_ttl":"72h0m0s"}`, string(byt))
	require.NoError(t, err)
}
