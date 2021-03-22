package pipelineyml

type Visitor interface {
	Visit(s *Spec)
}

type Visitable interface {
	Accept(v Visitor)
}

func (s *Spec) Accept(v Visitor) {
	v.Visit(s)
}
