package pipelineyml

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpgradeYmlFromV1(t *testing.T) {
	b, err := ioutil.ReadFile("./pipelineymlv1/samples/pipeline.yml")
	assert.NoError(t, err)
	newYmlByte, err := UpgradeYmlFromV1(b)
	assert.NoError(t, err)
	fmt.Println(string(newYmlByte))

}

func TestUpgradeYmlFromV1_PampasBlog(t *testing.T) {
	b, err := ioutil.ReadFile("./pipelineymlv1/samples/pipeline-pampas-blog.yml")
	assert.NoError(t, err)
	newYmlByte, err := UpgradeYmlFromV1(b)
	assert.NoError(t, err)
	fmt.Println(string(newYmlByte))
}
