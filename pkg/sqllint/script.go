package sqllint

type Script interface {
	Data() []byte
	Name() string
}

func NewScript(name string, data []byte) Script {
	return &script{
		data: data,
		name: name,
	}
}

type script struct {
	data []byte
	name string
}

func (s *script) Name() string {
	return s.name
}

func (s *script) Data() []byte {
	return s.data
}
