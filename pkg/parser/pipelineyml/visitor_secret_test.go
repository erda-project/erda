package pipelineyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretVisitor_Visit(t *testing.T) {
	secrets := map[string]string{
		"env_1": "((a))",
		"depth": "1",
		"a":     "b",
	}

	yamlByte := `version: 1.1
stages:
envs:
  ENV_1: ((env_1))
- stage:
  - git-checkout:
      params:
        depth: ((depth))
`

	visitor := NewSecretVisitor([]byte(yamlByte), secrets, 1)
	s := Spec{}
	visitor.Visit(&s)
	assert.Error(t, s.mergeErrors())
}

//func TestRenderSecrets(t *testing.T) {
//	input := []byte("((a))((b))((c))")
//	secret := map[string]string{
//		"a": "1",
//		"b": "2",
//	}
//	output, err := RenderSecrets(input, secret)
//	assert.Error(t, err)
//	_ = output
//
//	secret["c"] = "3"
//	output, err = RenderSecrets(input, secret)
//	assert.NoError(t, err)
//}
