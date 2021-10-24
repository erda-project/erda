package elasticsearch

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/core/monitor/settings/retention-strategy"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/creator"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	"github.com/erda-project/erda/modules/msp/apm/exception/erda-event/storage"
	"github.com/olivere/elastic"
	"time"
)

type (
	config struct {
		QueryTimeout time.Duration `file:"query_timeout" default:"1m"`
		WriteTimeout time.Duration `file:"write_timeout" default:"1m"`
		ReadPageSize int           `file:"read_page_size" default:"1024"`
		IndexType    string        `file:"index_type" default:"events"`
	}
	provider struct {
		Cfg       *config
		Log       logs.Logger
		ES        elasticsearch.Interface `autowired:"elasticsearch"`
		Loader    loader.Interface        `autowired:"elasticsearch.index.loader@event"`
		Creator   creator.Interface       `autowired:"elasticsearch.index.creator@event" optional:"true"`
		Retention retention.Interface     `autowired:"storage-retention-strategy@event" optional:"true"`
		client    *elastic.Client
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.client = p.ES.Client()
	if p.Retention != nil {
		ctx.AddTask(func(c context.Context) error {
			p.Retention.Loading(ctx)
			return nil
		})
	}
	return nil
}

var _ storage.Storage = (*provider)(nil)

func (p *provider) NewWriter(ctx context.Context) (storekit.BatchWriter, error) {
	if p.Creator == nil || p.Retention == nil {
		return nil, fmt.Errorf("elasticsearch.index.creator@event and storage-retention-strategy@event is required for Writer")
	}
	w := p.ES.NewWriter(&elasticsearch.WriteOptions{
		Timeout: p.Cfg.WriteTimeout,
		Enc: func(val interface{}) (index, id, typ string, body interface{}, err error) {
			data := val.(*exception.Erda_event)
			var wait <-chan error
			wait, index = p.Creator.Ensure(data.Tags["org_name"])
			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return "", "", "", nil, storekit.ErrExitConsume
				}
			}
			return index, data.EventId, p.Cfg.IndexType, data, nil
		},
	})
	return w, nil
}

func init() {
	servicehub.Register("event-storage-elasticsearch", &servicehub.Spec{
		Services:   []string{"event-storage-elasticsearch-reader", "event-storage-writer"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
