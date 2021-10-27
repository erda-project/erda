package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/core/monitor/settings/retention-strategy"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/creator"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	"github.com/erda-project/erda/modules/msp/apm/exception/erda-error/storage"
	"github.com/olivere/elastic"
)

type (
	config struct {
		QueryTimeout time.Duration `file:"query_timeout" default:"1m"`
		WriteTimeout time.Duration `file:"write_timeout" default:"1m"`
		ReadPageSize int           `file:"read_page_size" default:"1024"`
		IndexType    string        `file:"index_type" default:"errors"`
	}
	provider struct {
		Cfg          *config
		Log          logs.Logger
		ES1          elasticsearch.Interface `autowired:"elasticsearch@error" optional:"true"`
		ES2          elasticsearch.Interface `autowired:"elasticsearch" optional:"true"`
		Loader       loader.Interface        `autowired:"elasticsearch.index.loader@error"`
		Creator      creator.Interface       `autowired:"elasticsearch.index.creator@error" optional:"true"`
		Retention    retention.Interface     `autowired:"storage-retention-strategy@error" optional:"true"`
		es           elasticsearch.Interface
		client       *elastic.Client
		queryTimeout string
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	if p.ES1 != nil {
		p.es = p.ES1
	} else if p.ES2 != nil {
		p.es = p.ES2
	} else {
		return fmt.Errorf("elasticsearch is required")
	}
	p.client = p.es.Client()
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
		return nil, fmt.Errorf("elasticsearch.index.creator@error and storage-retention-strategy@error is required for Writer")
	}
	w := p.es.NewWriter(&elasticsearch.WriteOptions{
		Timeout: p.Cfg.WriteTimeout,
		Enc: func(val interface{}) (index, id, typ string, body interface{}, err error) {
			data := val.(*exception.Erda_error)
			var wait <-chan error
			wait, index = p.Creator.Ensure(data.Tags["org_name"])
			if wait != nil {
				select {
				case <-wait:
				case <-ctx.Done():
					return "", "", "", nil, storekit.ErrExitConsume
				}
			}
			return index, data.ErrorId, p.Cfg.IndexType, data, nil
		},
	})
	return w, nil
}

func init() {
	servicehub.Register("error-storage-elasticsearch", &servicehub.Spec{
		Services:   []string{"error-storage-elasticsearch-reader", "error-storage-writer"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
