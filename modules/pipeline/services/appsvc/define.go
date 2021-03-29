package appsvc

import (
	"github.com/erda-project/erda/bundle"
)

type AppSvc struct {
	bdl *bundle.Bundle
}

func New(bdl *bundle.Bundle) *AppSvc {
	s := AppSvc{}
	s.bdl = bdl
	return &s
}
