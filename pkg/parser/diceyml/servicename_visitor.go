package diceyml

import (
	"fmt"
	"regexp"
)

var (
	serviceNameRegex = regexp.MustCompile(`^((([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])\.)*([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9]))$`)
	// serviceNameMaxLen = 14 // service name 最长长度
	envNameRegex = regexp.MustCompile(`^([-._a-zA-Z][-._a-zA-Z0-9]*)$`)
)

type ServiceNameVisitor struct {
	DefaultVisitor
	collectErrors ValidateError
}

func NewServiceNameVisitor() DiceYmlVisitor {
	return &ServiceNameVisitor{
		collectErrors: ValidateError{},
	}
}

func (o *ServiceNameVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	if !serviceNameRegex.MatchString(o.currentService) {
		header := yamlHeaderRegexWithUpperHeader([]string{"services"}, o.currentService)
		err := fmt.Errorf(
			"service name[%s] does not match regex: %s, e.g: aa.bb-cc.dd-ee", o.currentService, serviceNameRegex)
		o.collectErrors[header] = err
	}

	o.checkEnvNames([]string{"services", o.currentService, "envs"}, obj.Envs)
}

func (o *ServiceNameVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	o.checkEnvNames([]string{"envs"}, obj.Envs)
}

func ServiceNameCheck(obj *Object) ValidateError {
	visitor := NewServiceNameVisitor()
	obj.Accept(visitor)
	return visitor.(*ServiceNameVisitor).collectErrors
}

func (o *ServiceNameVisitor) checkEnvNames(prefix []string, envs map[string]string) {
	for k := range envs {
		if !envNameRegex.MatchString(k) {
			header := yamlHeaderRegexWithUpperHeader(prefix, k)
			err := fmt.Errorf(
				"env name [%s] not match regex: %s, e.g. 'my.env-name', or 'MY_ENV.NAME'", k, envNameRegex)
			o.collectErrors[header] = err
		}
	}
}
