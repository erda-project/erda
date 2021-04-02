package k8sflink

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/clientgo"
)

type Flink struct {
	Client       *clientgo.ClientSet
	ExecutorName executortypes.Name
	ExecutorKind executortypes.Kind
}

type Option func(f *Flink)

func New(ops ...Option) *Flink {
	f := &Flink{}
	for _, op := range ops {
		op(f)
	}
	return f
}

func WithClient(addr string) Option {
	return func(f *Flink) {
		cs, err := clientgo.New(addr)
		if err != nil {
			logrus.Fatal(err)
		}
		f.Client = cs
	}
}

func WithName(name executortypes.Name) Option {
	return func(f *Flink) {
		f.ExecutorName = name
	}
}

func WithKind(kind executortypes.Kind) Option {
	return func(f *Flink) {
		f.ExecutorKind = kind
	}
}
