package pipelineyml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateYml(t *testing.T) {
	s := &Spec{Version: "1.1"}
	b, err := GenerateYml(s)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

func TestGenerateYml_NilAction(t *testing.T) {
	s := []byte(`
version: 1.1
stages:
- stage:
    - git-checkout:
`)

	y, err := New(s, WithSecrets(map[string]string{}))
	assert.NoError(t, err)

	b, err := GenerateYml(y.s)
	assert.NoError(t, err)
	fmt.Println(string(b))
}
