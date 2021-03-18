package diceyml

type CompatibleExposeVisitor struct {
	DefaultVisitor
}

func NewCompatibleExposeVisitor() DiceYmlVisitor {
	return &CompatibleExposeVisitor{}
}

func (o *CompatibleExposeVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	for _, v := range *obj {
		if len(v.Expose) > 0 && len(v.Ports) > 0 {
			v.Ports[0].Expose = true
		}
	}
}

func CompatibleExpose(obj *Object) {
	visitor := NewCompatibleExposeVisitor()
	obj.Accept(visitor)
}
