package dispatcher

import (
	"fmt"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/dispatcher/filters"
	"github.com/erda-project/erda/modules/eventbox/types"

	"github.com/sirupsen/logrus"
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
