package elasticsearch

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
	"time"
)

func (p *provider) Count(ctx context.Context, traceId string) int64 {
	indices := p.Loader.Indices(ctx, time.Now().Add(-time.Hour*24*7).UnixNano(), time.Now().UnixNano(), loader.KeyPath{
		Recursive: true,
	})

	fmt.Println(indices)

	if len(indices) <= 0 {
		return 0
	}

	// do query
	ctx, cancel := context.WithTimeout(ctx, p.Cfg.QueryTimeout)
	defer cancel()

	count, err := p.client.Count(indices...).
		IgnoreUnavailable(true).AllowNoIndices(true).Q("trace_id:" + traceId).Do(ctx)
	if err != nil {
		return 0
	}

	return count
}
