package pipelineymlv1

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
)

func TestPipelineYml_Unmarshal(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-anchor.yml")
	require.NoError(t, err)
	y := New(b)
	err = y.Parse()
	require.NoError(t, err)
}

func TestPipelineYml_Parse(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-get-put.yml")
	require.NoError(t, err)
	y := New(b)
	err = y.Parse()
	require.NoError(t, err)

	yamlWithUnknownFields := []byte(
		`version: '1.0'

stages:
- name: stage-test
  source:
    context: repo/ui
`)
	y = New(yamlWithUnknownFields)
	err = y.Parse()
	require.Error(t, err)
}

func TestPipelineYml_Triggers(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-trigger.yml")
	require.NoError(t, err)

	y := New(b)
	err = y.Parse(WithNFSRealPath("/"))
	require.NoError(t, err)
	fmt.Printf("%#v\n", y.obj.Triggers)
}

func TestPipelinYmlDuplicate(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-duplicate.yml")
	require.NoError(t, err)

	y := New(b)
	err = y.Parse()
	require.NoError(t, err)

	fmt.Println(y.YAML())
}

func TestPipelinYmlErrTasktype(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-err-tasktype.yml")
	require.NoError(t, err)

	y := New(b)
	err = y.Parse()
	require.NoError(t, err)

	fmt.Println(y.YAML())
}

func TestPipelineYml_ValidateSingleTaskConfig(t *testing.T) {
	b, err := ioutil.ReadFile("../pipeline-decode.yml")
	require.NoError(t, err)

	y := New(b)
	err = y.Parse()
	require.NoError(t, err)

	fmt.Println(y.YAML())
}

func TestPipelineYml_GenerateTemplateEnvs(t *testing.T) {
	y := New([]byte("version: '1.0'"))
	err := y.Parse(WithPlaceholders([]apistructs.MetadataField{{Name: "A", Value: "B"}, {Name: "C", Value: "D"}}))
	require.NoError(t, err)
	require.True(t, len(y.metadata.PlaceHolderEnvs) == 2)
}
