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
	"context"
	"encoding/json"
	"time"

	"github.com/olivere/elastic"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/oap/entity/pb"
	"github.com/erda-project/erda/modules/core/monitor/entity/storage"
)

func (p *provider) SetEntity(ctx context.Context, data *pb.Entity) error {
	index, id, typ, body, err := p.encodeToDocument(ctx)(data)
	if err != nil {
		return err
	}
	data.Id = id
	now := time.Now()
	data.CreateTimeUnixNano = now.UnixNano()
	data.UpdateTimeUnixNano = now.UnixNano()

	// always override
	// _, err = p.es.Client().Index().
	// 	Index(index).
	// 	Type(typ).
	// 	Id(id).
	// 	BodyJson(body).
	// 	Timeout(p.writeTimeoutMS).
	// 	Do(ctx)

	// upsert
	values := data.Values
	if values == nil {
		values = map[string]*structpb.Value{}
	}
	labels := data.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	_, err = p.es.Client().Update().Index(index).
		Type(typ).
		Id(id).
		Upsert(body).
		Doc(map[string]interface{}{
			"id":                 data.Id,
			"type":               data.Type,
			"key":                data.Key,
			"values":             values,
			"labels":             labels,
			"updateTimeUnixNano": data.UpdateTimeUnixNano,
		}).
		Timeout(p.writeTimeoutMS).
		Do(ctx)
	return err
}

func (p *provider) RemoveEntity(ctx context.Context, typ, key string) (bool, error) {
	_, err := p.es.Client().Delete().
		Index(p.getIndex(typ, key)).
		Type(p.Cfg.IndexType).
		Id(p.getDocumentID(typ, key)).
		Timeout(p.writeTimeoutMS).
		Do(ctx)
	if err != nil {
		if e, ok := err.(*elastic.Error); ok && e.Status == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *provider) GetEntity(ctx context.Context, typ, key string) (*pb.Entity, error) {
	query := elastic.NewBoolQuery().
		Filter(elastic.NewTermQuery("type", typ)).
		Filter(elastic.NewTermQuery("key", key))
	searchSource := elastic.NewSearchSource().Query(query)
	resp, err := p.es.Client().
		Search(p.getIndex(typ, key)).
		IgnoreUnavailable(true).AllowNoIndices(true).Timeout(p.queyTimeoutMS).
		SearchSource(searchSource).Size(1).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Hits == nil || len(resp.Hits.Hits) <= 0 || resp.Hits.Hits[0].Source == nil {
		return nil, nil
	}
	return parseData(*resp.Hits.Hits[0].Source)
}

func (p *provider) ListEntities(ctx context.Context, opts *storage.ListOptions) ([]*pb.Entity, int64, error) {
	query := elastic.NewBoolQuery()
	var typ string
	if opts != nil {
		typ = opts.Type
		if len(opts.Type) > 0 {
			query = query.Filter(elastic.NewTermQuery("type", opts.Type))
		}
		for k, v := range opts.Labels {
			query = query.Filter(elastic.NewTermQuery("labels."+k, v))
		}
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	searchSource := elastic.NewSearchSource().Query(query)
	resp, err := p.es.Client().
		Search(p.getIndices(typ)...).
		IgnoreUnavailable(true).AllowNoIndices(true).Timeout(p.queyTimeoutMS).
		SearchSource(searchSource).Size(limit).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}
	if resp == nil || resp.Hits == nil {
		return nil, 0, nil
	}
	list := make([]*pb.Entity, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		data, err := parseData(*hit.Source)
		if err != nil {
			continue
		}
		list = append(list, data)
	}
	return list, resp.Hits.TotalHits, nil
}

func parseData(data []byte) (*pb.Entity, error) {
	entity := &pb.Entity{}
	err := json.Unmarshal(data, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
