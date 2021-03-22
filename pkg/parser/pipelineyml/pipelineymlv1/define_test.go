package pipelineymlv1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	y :=
		`
version: '1.0'

triggers:
- schedule:
    cron: "* * * * *"
    filters:
    - type: git-branch
      onlys:
      - master
`
	err := New([]byte(y)).Parse()
	require.NoError(t, err)
}
