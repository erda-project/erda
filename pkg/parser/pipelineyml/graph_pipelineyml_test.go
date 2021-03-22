package pipelineyml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertGraphPipelineYml(t *testing.T) {
	fb, err := ioutil.ReadFile("./samples/graph_pipelineyml.yaml")
	assert.NoError(t, err)
	b, err := ConvertGraphPipelineYmlContent(fb)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

func TestConvertToGraphPipelineYml(t *testing.T) {
	fb, err := ioutil.ReadFile("./samples/pipeline_cicd.yml")
	assert.NoError(t, err)
	graph, err := ConvertToGraphPipelineYml(fb)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(graph, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(b))
}
