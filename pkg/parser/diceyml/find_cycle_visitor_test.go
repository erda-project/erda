package diceyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractChain(t *testing.T) {
	nodes := map[string]map[string]struct{}{
		"a": {"b": {}, "c": {}},
		"b": {"c": {}},
		"c": {"a": {}},
	}
	chain := extractCycle(nodes)
	assert.Equal(t, 4, len(chain), "%v", chain)
}
