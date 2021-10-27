package elasticsearch

//
//import (
//	"context"
//	"fmt"
//	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
//	"github.com/erda-project/erda/modules/msp/apm/exception/erda-error/storage"
//)
//
//func (p *provider) Count(ctx context.Context, sel *storage.Selector) int64 {
//	indices := p.Loader.Indices(ctx, sel.StartTime, sel.EndTime, loader.KeyPath{
//		Recursive: true,
//	})
//	fmt.Println(indices)
//
//	if len(indices) <= 0 {
//		return 0
//	}
//
//	// do query
//	ctx, cancel := context.WithTimeout(ctx, p.Cfg.QueryTimeout)
//	defer cancel()
//
//	count, err := p.client.Count(indices...).
//		IgnoreUnavailable(true).AllowNoIndices(true).Q("timestamp:[" + string(sel.StartTime) + " TO " + string(sel.EndTime) + "] AND error_id:" + sel.ErrorId).Do(ctx)
//	if err != nil {
//		return 0
//	}
//
//	return count
//}
