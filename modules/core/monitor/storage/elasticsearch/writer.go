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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/core/monitor/storage"
	"github.com/olivere/elastic"
)

const maxUnixMillisecond int64 = 9999999999999

func getUnixMillisecond(ts int64) int64 {
	if ts > maxUnixMillisecond {
		return ts / int64(time.Millisecond)
	}
	return ts
}

func (s *esStorage) NewWriter(opts *storage.WriteOptions) (storage.Writer, error) {
	return newWriter(s, opts), nil
}

func (s *esStorage) NewBatchWriter(opts *storage.WriteOptions) (storage.BatchWriter, error) {
	return newWriter(s, opts), nil
}

type esWriter struct {
	opts    *storage.WriteOptions
	client  *elastic.Client
	typ     string
	timeout string
}

func newWriter(s *esStorage, opts *storage.WriteOptions) *esWriter {
	timeout := s.writeTimeout
	if opts.Timeout > 0 {
		timeout = fmt.Sprintf("%dms", opts.Timeout.Milliseconds())
	}
	return &esWriter{
		opts:    opts,
		client:  s.client,
		typ:     s.typ,
		timeout: timeout,
	}
}

func (w *esWriter) Close() error { return nil }

func (w *esWriter) Write(val *storage.Data) error {
	if val == nil {
		return storage.ErrEmptyData
	}
	index, id := w.opts.PartitionKeyFunc(val), w.opts.KeyFunc(val)
	_, err := w.client.Index().
		Index(index).Id(id).Type(w.typ).
		BodyJson(&Data{
			Timestamp: val.Timestamp,
			Date:      getUnixMillisecond(val.Timestamp),
			Tags:      val.Labels,
			Fields:    val.Fields,
		}).Timeout(w.timeout).Do(context.Background())
	if err != nil {
		return &storage.WriteError{
			Data: val,
			Err:  err,
		}
	}
	return err
}

func (w *esWriter) WriteN(list ...*storage.Data) (int, error) {
	if len(list) <= 0 {
		return 0, nil
	}
	requests := make([]elastic.BulkableRequest, len(list), len(list))
	for i, data := range list {
		index, id := w.opts.PartitionKeyFunc(data), w.opts.KeyFunc(data)
		req := elastic.NewBulkIndexRequest().Index(index).Id(id).Type(w.typ).Doc(&Data{
			Timestamp: data.Timestamp,
			Date:      getUnixMillisecond(data.Timestamp),
			Tags:      data.Labels,
			Fields:    data.Fields,
		})
		requests[i] = req
	}
	res, err := w.client.Bulk().Add(requests...).Timeout(w.timeout).Do(context.Background())
	if err != nil {
		var berr storage.BatchWriteError
		for _, item := range list {
			berr.Errors = append(berr.Errors, &storage.WriteError{
				Data: item,
				Err:  err,
			})
		}
		return 0, &berr
	}
	if res.Errors {
		if len(res.Items) != len(list) {
			return 0, fmt.Errorf("request items(%d), but response items(%d)", len(list), len(res.Items))
		}
		var berr storage.BatchWriteError
		for i, item := range res.Items {
			for _, result := range item { // len(item) is 1, contains index request only
				if !(result.Status >= 200 && result.Status <= 299) {
					var sb strings.Builder
					json.NewEncoder(&sb).Encode(result)
					berr.Errors = append(berr.Errors, &storage.WriteError{
						Data: list[i],
						Err:  errors.New(sb.String()),
					})
					break
				}
			}
		}
		return len(list) - len(berr.Errors), &berr
	}
	return len(list), nil
}
