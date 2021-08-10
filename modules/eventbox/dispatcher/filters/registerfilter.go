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
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/constant"
	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/register"
	"github.com/erda-project/erda/modules/eventbox/types"
)

type RegisterFilter struct {
	reg register.Register
}

func NewRegisterFilter(r register.Register) Filter {
	return &RegisterFilter{
		reg: r,
	}
}

func (*RegisterFilter) Name() string {
	return "RegisterFilter"
}

func (r *RegisterFilter) Filter(m *types.Message) *errors.DispatchError {
	derr := errors.New()
	keys, ok := m.Labels[types.LabelKey(constant.RegisterLabelKey)]
	if !ok {
		return derr
	}
	keys_, err := getLabelKeys(keys)
	if err != nil {
		errStr := fmt.Errorf("RegisterFilter: register value illegal type: %v", err)
		logrus.Error(errStr)
		derr.FilterErr = errStr
		return derr
	}
	for _, key := range keys_ {
		keyLabels := r.reg.PrefixGet(key)
		if keyLabels == nil {
			infoStr := fmt.Sprintf("RegisterFilter: prefixget labelkey: %s, notfound", key)
			logrus.Warn(infoStr)
			derr.FilterInfo = infoStr
		}
		for _, labels := range keyLabels {
			m.Labels = mergeLabels(labels, m.Labels)
		}
	}
	return derr

}

// l2 has higher priority
func mergeLabels(l1, l2 map[types.LabelKey]interface{}) map[types.LabelKey]interface{} {
	l := make(map[types.LabelKey]interface{})
	for k, v := range l1 {
		l[k] = v
	}
	for k, v := range l2 {
		l[k] = v
	}
	return l
}

func getLabelKeys(keys interface{}) ([]string, error) {
	switch i := keys.(type) {
	case []interface{}:
		r := []string{}
		for _, k := range i {
			k_, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("expect string or []string, got []%s", reflect.TypeOf(k))
			}
			r = append(r, k_)
		}
		return r, nil
	case string:
		return []string{i}, nil
	case []string:
		return i, nil
	default:
		return nil, fmt.Errorf("expect string or []string, got %s", reflect.TypeOf(keys))
	}

}
