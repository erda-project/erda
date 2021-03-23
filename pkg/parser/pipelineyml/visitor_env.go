package pipelineyml

import (
	"regexp"

	"github.com/pkg/errors"
)

var envKeyRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

type EnvVisitor struct {
	envs map[string]string
}

func NewEnvVisitor(envs map[string]string) *EnvVisitor {
	return &EnvVisitor{envs: envs}
}

func (v *EnvVisitor) Visit(s *Spec) {
	if s.Envs == nil {
		s.Envs = make(map[string]string, 0)
	}

	for k, v := range v.envs {
		s.Envs[k] = v
	}

	checkErr := CheckEnvs(s.Envs)
	for _, err := range checkErr {
		s.appendError(err)
	}
}

func CheckEnvs(envs map[string]string) []error {
	var errs []error
	for k := range envs {
		if !envKeyRegexp.MatchString(k) {
			errs = append(errs, errors.Errorf("invalid env key: %s (must match: %s)", k, envKeyRegexp.String()))
		}
	}
	return errs
}
