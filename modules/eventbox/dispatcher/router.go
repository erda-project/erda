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

package dispatcher

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/dispatcher/filters"
	"github.com/erda-project/erda/modules/eventbox/types"
)

// dispatcher 内部处理 message 的相关逻辑

//               +-------------------------+
//               |     Router              |
//               |                         |
//               |     +-------------+     |      +-------------+
//               |     |             |     |      |  backend    |
// src ----------+---->|      A      +-----+------>  message    |
//               |     |             |     |      |  consumer   |
//               |     +-------------+     |      | e.g.dingding|
//               +-------------------------+      +-------------+
// A: []filter
//
// []filter:
//     +---------------+  +---------------+  +----------------+	 +-----------------+
//     | unifylabels   +--> registerlabel +--> webhookfilter  +-->  lastfilter     |
//     |               |  |               |  |                |	 |                 |
//     +---------------+  +---------------+  +----------------+	 +-----------------+
//
//
type Router struct {
	dispatcher *DispatcherImpl
	filters    []filters.Filter
}

func NewRouter(dispatcher *DispatcherImpl) (*Router, error) {
	r := &Router{
		dispatcher: dispatcher,
		filters:    []filters.Filter{},
	}

	unifyLabelsFilter := filters.NewUnifyLabelsFilter()
	registerFilter := filters.NewRegisterFilter(dispatcher.GetRegister())
	webhookFilter, err := filters.NewWebhookFilter()
	if err != nil {
		return nil, fmt.Errorf("init webhookfilter: %v", err)
	}
	lastFilter := filters.NewLastFilter(dispatcher.GetSubscribersPool(), dispatcher.GetSubscribers())

	r.RegisterFilter(unifyLabelsFilter)
	r.RegisterFilter(registerFilter)
	r.RegisterFilter(webhookFilter)
	r.RegisterFilter(lastFilter)

	return r, nil
}

func (r *Router) RegisterFilter(f filters.Filter) {
	logrus.Infof("Router register filter [%s]", f.Name())
	r.filters = append(r.filters, f)
}

func (r *Router) Route(m *types.Message) *errors.DispatchError {
	for i, f := range r.filters {
		derr := f.Filter(m)
		if derr != nil && !derr.IsOK() {
			if len(r.filters)-1 != i {
				logrus.Warnf("Route: %v", derr)
			}
			return derr
		}
		if derr != nil && derr.FilterInfo != "" {
			logrus.Warnf("Route: FilterInfo: %s", derr.FilterInfo)
		}
	}
	return errors.New()
}

func (r *Router) GetFilters() []filters.Filter {
	return r.filters
}
