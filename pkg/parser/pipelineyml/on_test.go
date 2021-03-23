package pipelineyml

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOn(t *testing.T) {
	b, err := ioutil.ReadFile("./samples/on.yml")
	assert.NoError(t, err)
	newYml, err := ConvertToGraphPipelineYml(b)
	assert.NoError(t, err)
	assert.Contains(t, newYml.YmlContent, "push")
	//n := &apistructs.PipelineYml{}
	//assert.EqualValues(t, newYml.On, n.On)
	fmt.Println(newYml.YmlContent)
}
