package diceyml

import (
	"fmt"
)

type CheckEnvSizeVisitor struct {
	DefaultVisitor
	collectErrors ValidateError
}

func NewCheckEnvSizeVisitor() DiceYmlVisitor {
	return &CheckEnvSizeVisitor{
		collectErrors: ValidateError{},
	}
}

func (o *CheckEnvSizeVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	for k, v := range obj.Envs {
		if toolong(v) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{"envs"}, k)] = fmt.Errorf("global env too long(more than 20000 chars, ~20kb): [%s]", k)
		}
	}
}

func (o *CheckEnvSizeVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	for k, v := range obj.Envs {
		if toolong(v) {
			o.collectErrors[yamlHeaderRegexWithUpperHeader([]string{o.currentService, "envs"}, k)] = fmt.Errorf("env too long(more than 20000 chars, ~20kb): %s", k)
		}
	}
}

func toolong(s string) bool {
	return len(s) > 20000 // ~20kb
}

func CheckEnvSize(obj *Object) ValidateError {
	visitor := NewCheckEnvSizeVisitor()
	obj.Accept(visitor)
	return visitor.(*CheckEnvSizeVisitor).collectErrors
}
