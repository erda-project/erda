package gittarutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepo_GetCommit(t *testing.T) {
	commit, err := r.GetCommit("feature/init_sql")
	require.NoError(t, err)
	fmt.Printf("%+v\n", commit)
}
