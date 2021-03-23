// 该 visitor 处理 diceyml 中一些字段的默认值
package diceyml

type DefaultValueVisitor struct {
	DefaultVisitor
}

func NewDefaultValueVisitor() DiceYmlVisitor {
	return &DefaultValueVisitor{}
}

func (o *DefaultValueVisitor) VisitResources(v DiceYmlVisitor, obj *Resources) {
	if obj.Network == nil {
		obj.Network = map[string]string{}
	}
	if _, ok := obj.Network["mode"]; !ok {
		obj.Network["mode"] = "container"
	}
}

func SetDefaultValue(obj *Object) {
	visitor := NewDefaultValueVisitor()
	obj.Accept(visitor)
}
