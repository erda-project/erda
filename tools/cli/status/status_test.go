package status

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteSessionInfoFromMap(t *testing.T) {
	sessions := map[string]StatusInfo{
		"https://openapi.erda.cloud":  {Token: "Bearer token"},
		"https://openapi.example.com": {SessionID: "legacy"},
	}

	result := deleteSessionInfoFromMap(sessions, "https://openapi.erda.cloud")
	require.Len(t, result, 1)
	_, exists := result["https://openapi.erda.cloud"]
	require.False(t, exists)
	require.Equal(t, "legacy", result["https://openapi.example.com"].SessionID)
}
