package clickhouse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	p := provider{}
	require.True(t, p.Select([]string{"metric"}))
}

func TestNewWrite(t *testing.T) {
	p := provider{}
	write, err := p.NewWriter(context.Background())
	require.NoError(t, err)
	require.Nil(t, write)
}
