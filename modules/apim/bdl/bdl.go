package bdl

import (
	"github.com/erda-project/erda/bundle"
)

var Bdl *bundle.Bundle

func Init(opts ...bundle.Option) {
	if Bdl == nil {
		Bdl = bundle.New(opts...)
	}
}
