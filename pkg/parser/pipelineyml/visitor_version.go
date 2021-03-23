package pipelineyml

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type VersionVisitor struct{}

func NewVersionVisitor() *VersionVisitor {
	return &VersionVisitor{}
}

func (v *VersionVisitor) Visit(s *Spec) {
	switch s.Version {
	case "":
		s.appendError(errors.New("no version"))
	case Version1dot1:
		return
	default:
		s.appendError(errors.Errorf("invalid version: %s, only support 1.1", s.Version))
	}
}

func GetVersion(data []byte) (string, error) {
	type version struct {
		Version string `yaml:"version"`
	}
	var ver version
	if err := yaml.Unmarshal(data, &ver); err != nil {
		return "", err
	}
	return ver.Version, nil
}
