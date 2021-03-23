package pipelineyml

import (
	"testing"
)

func TestEnvVisitor_Visit(t *testing.T) {
	v := NewEnvVisitor(nil)

	s := Spec{
		Envs: map[string]string{
			",":        "",
			"-":        "",
			"A-":       "",
			"blog-web": "",
		},
	}

	s.Accept(v)

	if len(s.errs) != len(s.Envs) {
		t.Fatal(s.mergeErrors())
	}
}
