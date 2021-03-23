package extmarketsvc

import (
	"github.com/erda-project/erda/bundle"
)

type ExtMarketSvc struct {
	bdl *bundle.Bundle
}

func New(bdl *bundle.Bundle) *ExtMarketSvc {
	s := ExtMarketSvc{}
	s.bdl = bdl
	return &s
}
