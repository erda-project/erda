package script

// Script represents a script file
type Script struct {
	// data is script file content
	data []byte
	// name is script file name
	name string
}

// New returns a *Script
func New(name string, data []byte) Script {
	return Script{
		data: data,
		name: name,
	}
}

// Name returns script file name
func (s *Script) Name() string {
	return s.name
}

// Data returns script file content
func (s *Script) Data() []byte {
	return s.data
}
