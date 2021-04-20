// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
