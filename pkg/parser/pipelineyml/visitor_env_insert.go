package pipelineyml

type EnvInsertVisitor struct {
	envs map[string]string
}

func NewEnvInsertVisitor(envs map[string]string) *EnvInsertVisitor {
	return &EnvInsertVisitor{envs: envs}
}

func (v *EnvInsertVisitor) Visit(s *Spec) {
	if v.envs == nil {
		return
	}

	if s.Envs == nil {
		s.Envs = make(map[string]string)
	}

	// insert into struct
	for k, v := range v.envs {
		s.Envs[k] = v
	}

	// encode to new yaml
	_, err := GenerateYml(s)
	if err != nil {
		s.appendError(err)
		return
	}
}
