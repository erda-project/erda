package agenttool

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTar(t *testing.T) {
	err := tar("b.tar", "/tmp/gittar/telegraf")
	require.NoError(t, err)

	p, err := filepath.Abs(".")
	require.NoError(t, err)

	err = unTar("b.tar", p)
	require.NoError(t, err)
}
