package gittarutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepo_FetchFiles(t *testing.T) {
	_, err := r.FetchFiles("feature/init_sql", "pipeline.yml")
	require.NoError(t, err)

	_, err = r.FetchFiles("feature/init_sql", "dice.yml")
	require.NoError(t, err)

	_, err = r.FetchFiles("feature/init_sql", "no")
	require.Error(t, err)
}

func TestRepo_FetchPipelineYml(t *testing.T) {
	_, err := r.FetchPipelineYml("feature/empty-pipelineyml")
	require.Error(t, err)
}

func TestRepo_FetchFile(t *testing.T) {
	_, err := r.FetchFile("feature/init_sql", "README.md")
	require.NoError(t, err)
}
