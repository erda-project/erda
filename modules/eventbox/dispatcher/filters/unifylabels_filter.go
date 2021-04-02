package filters

import (
	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type UnifyLabelsFilter struct{}

func NewUnifyLabelsFilter() Filter {
	return &UnifyLabelsFilter{}
}

func (*UnifyLabelsFilter) Name() string {
	return "UnifyLabelsFilter"
}

func (*UnifyLabelsFilter) Filter(m *types.Message) *errors.DispatchError {
	labels := m.Labels
	normalizedLabels := map[types.LabelKey]interface{}{}
	for k, v := range labels {
		normalizedKey := k.NormalizeLabelKey()
		normalizedLabels[normalizedKey] = v
	}
	m.Labels = normalizedLabels
	return nil
}
