// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
)

type (
	config struct {
		QueryTimeout time.Duration `file:"query_timeout"`
		WriteTimeout time.Duration `file:"write_timeout"`
		IndexType    string        `file:"index_type" default:"default"`
	}
	provider struct {
		Cfg          *config
		Log          logs.Logger
		ES           elasticsearch.Interface `autowired:"elasticsearch"`
		queryTimeout string
		writeTimeout string
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {
	if p.Cfg.QueryTimeout > 0 {
		p.queryTimeout = fmt.Sprintf("%dms", p.Cfg.QueryTimeout.Milliseconds())
	}
	if p.Cfg.WriteTimeout > 0 {
		p.writeTimeout = fmt.Sprintf("%dms", p.Cfg.WriteTimeout.Milliseconds())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return &esStorage{
		client:       p.ES.Client(),
		typ:          p.Cfg.IndexType,
		queryTimeout: p.queryTimeout,
		writeTimeout: p.writeTimeout,
	}
}

// func (p *provider) process(hits *elastic.SearchHits) error {
// 	for _, hit := range hits.Hits {
// 		_ = hit
// 		// fmt.Println(string(*hit.Source))
// 	}
// 	return nil
// }

// func (p *provider) testQuery2() error {
// 	s := &esStorage{
// 		client: p.ES.Client(),
// 		log:    p.Log,
// 	}
// 	end := 1632287981213 * int64(time.Millisecond)
// 	start := end - 30*int64(time.Second)
// 	sel := &storage.Selector{
// 		StartTime:     start,
// 		EndTime:       end,
// 		PartitionKeys: []string{"spot-docker_container_summary-full_cluster-*"},
// 	}
// 	it := s.Iterator(context.TODO(), sel)
// 	defer it.Close()

// 	it.Prev()
// 	fmt.Println(it.Value())
// 	it.First()
// 	fmt.Println(it.Value())

// 	// var total int
// 	// for it.Prev() {
// 	// 	fmt.Println(it.Value())
// 	// 	total++
// 	// }
// 	// fmt.Println("total", total)
// 	return it.Error()
// }

// func (p *provider) testQuery() error {
// 	client := p.ES.Client()

// 	// now := time.Now()
// 	// end := now.UnixNano()
// 	// start := now.Add(-5 * time.Minute).UnixNano()
// 	end := 1632287981213 * int64(time.Millisecond)
// 	start := end - 60*int64(time.Minute)

// 	searchSource := elastic.NewSearchSource()
// 	query := elastic.NewBoolQuery().Filter(elastic.NewRangeQuery("timestamp").Gte(start).Lte(end))
// 	searchSource.Query(query)

// 	indices := []string{"spot-docker_container_summary-full_cluster-*"}
// 	const pageSize = 8 * 1024

// 	startTime := time.Now()
// 	resp, err := client.Scroll(indices...).KeepAlive("1m").
// 		IgnoreUnavailable(true).AllowNoIndices(true).
// 		SearchSource(searchSource).Size(pageSize).Do(context.TODO())
// 	fmt.Println("query duration: ", time.Since(startTime))
// 	if err != nil {
// 		if err == io.EOF {
// 			return nil
// 		}
// 		return err
// 	}
// 	scrollIds := map[string]struct{}{
// 		resp.ScrollId: {},
// 	}
// 	defer func() {
// 		if len(scrollIds) > 0 {
// 			var list []string
// 			for id := range scrollIds {
// 				if len(id) > 0 {
// 					list = append(list, id)
// 					fmt.Println("clear ", id)
// 				}
// 			}
// 			_, err := client.ClearScroll(list...).Do(context.TODO())
// 			if err != nil {
// 				p.Log.Error(err)
// 			}
// 		}
// 	}()
// 	var total int
// 	for {
// 		total += len(resp.Hits.Hits)
// 		err := p.process(resp.Hits)
// 		if err != nil {
// 			return err
// 		}
// 		if len(resp.ScrollId) <= 0 || len(resp.Hits.Hits) <= 0 {
// 			fmt.Println("break", resp.ScrollId, len(resp.Hits.Hits))
// 			break
// 		}

// 		startTime := time.Now()
// 		resp, err = client.Scroll().KeepAlive("1m").ScrollId(resp.ScrollId).Size(pageSize).IgnoreUnavailable(true).AllowNoIndices(true).Do(context.TODO())
// 		fmt.Println("query duration: ", time.Since(startTime))
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return err
// 		}
// 		scrollIds[resp.ScrollId] = struct{}{}
// 	}
// 	fmt.Println(total, " ", resp.Hits.TotalHits)
// 	return nil
// }

func init() {
	servicehub.Register("elasticsearch-storage", &servicehub.Spec{
		Services:   []string{"elasticsearch-storage-reader", "elasticsearch-storage-writer"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
