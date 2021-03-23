package dbclient

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestClient_CheckExistPipelineCronByApplicationBranchYmlName(t *testing.T) {
	exist, pipelineCron, err := client.CheckExistPipelineCronByApplicationBranchYmlName(125, "feature/cron-pipeline", "pipeline.yml")
	require.True(t, exist)
	spew.Dump(pipelineCron)
	require.NoError(t, err)
}
