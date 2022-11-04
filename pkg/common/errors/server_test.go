package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalServerErrorFormatErr(t *testing.T) {
	err := NewInternalServerErrorMessage("test")
	require.Equal(t, err.Error(), "internal error: test")
}
