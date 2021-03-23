package diceyml

type ComposeVisitor struct {
	DefaultVisitor
	envObj *Object
	env    string
}

func NewComposeVisitor(envObj *Object, env string) DiceYmlVisitor {
	return &ComposeVisitor{envObj: envObj, env: env}
}

func (o *ComposeVisitor) VisitEnvObjects(v DiceYmlVisitor, obj *EnvObjects) {

	if o.envObj == nil {
		return
	}

	if (*obj) == nil {
		obj_ := EnvObjects(map[string]*EnvObject{})
		*obj = obj_
	}

	if e, ok := (*obj)[o.env]; !ok || e == nil {
		(*obj)[o.env] = new(EnvObject)
	}

	overrideIfNotZero(o.envObj.Envs, &(*obj)[o.env].Envs)
	overrideIfNotZero(o.envObj.Services, &(*obj)[o.env].Services)
	overrideIfNotZero(o.envObj.AddOns, &(*obj)[o.env].AddOns)

}

func Compose(obj, envObj *Object, env string) {
	visitor := NewComposeVisitor(envObj, env)

	obj.Accept(visitor)

}
