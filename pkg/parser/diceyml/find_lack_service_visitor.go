// 找出依赖却不存在的 service
package diceyml

import (
	"fmt"
)

type FindLackServiceVisitor struct {
	DefaultVisitor
	lack ValidateError
}

func NewFindLackServiceVisitor() DiceYmlVisitor {
	return &FindLackServiceVisitor{
		lack: ValidateError{},
	}
}

func (o *FindLackServiceVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	for srvName, srv := range *obj {
		for _, name := range srv.DependsOn {
			if _, ok := (*obj)[name]; !ok {
				o.lack[yamlHeaderRegexWithUpperHeader([]string{"services", srvName}, "depends_on")] = fmt.Errorf("found lacked services: %v", name)
			}
		}
	}
}

func FindLackService(obj *Object) ValidateError {
	visitor := NewFindLackServiceVisitor()
	obj.Accept(visitor)
	return visitor.(*FindLackServiceVisitor).lack
}
