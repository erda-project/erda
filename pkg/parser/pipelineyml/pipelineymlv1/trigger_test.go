package pipelineymlv1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelineYml_GetTriggerScheduleCron(t *testing.T) {
	y := New([]byte(
		`
version: '1.0'

triggers:
- schedule:
    cron: "* * * * *"
    filters:
    - type: git-branch
      onlys:
      - master
- schedule:
    cron: "*/5 * * * *"
    filters:
    - type: git-branch
      excepts:
      - test
`))
	err := y.Parse(WithBranch("master"))
	require.Error(t, err)

	err = y.Parse(WithBranch("develop"))
	require.NoError(t, err)
}
