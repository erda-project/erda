package filehelper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateFile(t *testing.T) {
	err := CreateFile("/tmp/ci/test.sh", "echo hello\necho world\necho nice", 0700)
	require.NoError(t, err)
}

func TestCreateFile2(t *testing.T) {
	err := CreateFile2("/tmp/ci/test.sh", bytes.NewBufferString("ssssxfdsfs\nfs"), 0700)
	require.NoError(t, err)
}
