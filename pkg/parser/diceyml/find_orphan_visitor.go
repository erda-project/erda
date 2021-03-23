// 找出没有被其他service依赖的后端service， 前端service不考虑（也就是有expose的service）
package diceyml

type FindOrphanVisitor struct {
	DefaultVisitor
	orphan []string
}

func NewFindOrphanVisitor() DiceYmlVisitor {
	return &FindOrphanVisitor{}
}

func (o *FindOrphanVisitor) VisitServices(v DiceYmlVisitor, obj *Services) {
	orphan := []string{}
	depends := map[string]struct{}{}
	for _, srv := range *obj {
		for _, name := range srv.DependsOn {
			depends[name] = struct{}{}
		}
	}
	for name, srv := range *obj {
		_, ok := depends[name]
		if !ok && srv.Expose == nil {
			orphan = append(orphan, name)
		}
	}
	o.orphan = orphan
}

func FindOrphan(obj *Object) []string {
	visitor := NewFindOrphanVisitor()
	obj.Accept(visitor)
	return visitor.(*FindOrphanVisitor).orphan
}
