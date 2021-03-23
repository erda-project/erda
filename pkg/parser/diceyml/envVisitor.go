// 原本的parser会把 global env 塞到各个 service 的 env 去
package diceyml

type EnvVisitor struct {
	DefaultVisitor
	globalEnv map[string]string
}

func NewEnvVisitor() DiceYmlVisitor {
	return &EnvVisitor{}
}

func (o *EnvVisitor) VisitObject(v DiceYmlVisitor, obj *Object) {
	override(obj.Envs, &o.globalEnv)
	obj.Services.Accept(v)
}

func (o *EnvVisitor) VisitService(v DiceYmlVisitor, obj *Service) {
	if obj.Envs == nil {
		obj.Envs = map[string]string{}
	}
	for k, v := range o.globalEnv {
		if _, ok := obj.Envs[k]; !ok {

			obj.Envs[k] = v
		}
	}
}

func ExpandGlobalEnv(obj *Object) {
	visitor := NewEnvVisitor()
	obj.Accept(visitor)
}
