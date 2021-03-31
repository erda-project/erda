package filters

import (
	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type Filter interface {
	Filter(m *types.Message) *errors.DispatchError
	Name() string
}
