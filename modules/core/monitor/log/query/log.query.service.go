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

package query

import (
	context "context"
	"regexp"
	"sort"
	"time"

	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/pkg/common/errors"
)

type logQueryService struct {
	p                   *provider
	startTime           int64
	storageReader       storage.Storage
	k8sReader           storage.Storage
	frozenStorageReader storage.Storage
}

func (s *logQueryService) GetLog(ctx context.Context, req *pb.GetLogRequest) (*pb.GetLogResponse, error) {
	items, err := s.queryLogItems(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogResponse{Lines: items}, nil
}

func (s *logQueryService) GetLogByRuntime(ctx context.Context, req *pb.GetLogByRuntimeRequest) (*pb.GetLogByRuntimeResponse, error) {
	if len(req.ApplicationId) <= 0 {
		return nil, errors.NewMissingParameterError("applicationId")
	}
	items, err := s.queryLogItems(ctx, req, func(sel *storage.Selector) *storage.Selector {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "tags.dice_application_id",
			Op:    storage.EQ,
			Value: req.ApplicationId,
		})
		return sel
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetLogByRuntimeResponse{Lines: items}, nil
}

func (s *logQueryService) GetLogByOrganization(ctx context.Context, req *pb.GetLogByOrganizationRequest) (*pb.GetLogByOrganizationResponse, error) {
	if len(req.ClusterName) <= 0 {
		return nil, errors.NewMissingParameterError("applicationId")
	}
	items, err := s.queryLogItems(ctx, req, func(sel *storage.Selector) *storage.Selector {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "tags.dice_cluster_name",
			Op:    storage.EQ,
			Value: req.ClusterName,
		})
		return sel
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetLogByOrganizationResponse{Lines: items}, nil
}

func (s *logQueryService) queryLogItems(ctx context.Context, req Request, fn func(sel *storage.Selector) *storage.Selector) ([]*pb.LogItem, error) {
	sel, err := toQuerySelector(req)
	if err != nil {
		return nil, err
	}
	if fn != nil {
		sel = fn(sel)
	}
	it, err := s.getIterator(ctx, sel, req.GetLive())
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer it.Close()

	items, err := toLogItems(ctx, it, req.GetCount() >= 0, getLimit(req.GetCount()))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return items, nil
}

func (s *logQueryService) walkLogItems(ctx context.Context, req Request, fn func(sel *storage.Selector) *storage.Selector, walk func(item *pb.LogItem) error) error {
	if req.GetCount() < 0 {
		return errors.NewInvalidParameterError("count", "not allowed negative")
	}
	sel, err := toQuerySelector(req)
	if err != nil {
		return err
	}
	if fn != nil {
		fn(sel)
	}
	it, err := s.getIterator(ctx, sel, req.GetLive())
	if err != nil {
		return errors.NewInternalServerError(err)
	}
	defer it.Close()
	num, limit := 0, getLimit(req.GetCount())
	for it.Next() {
		if num >= limit {
			break
		}
		log, ok := it.Value().(*pb.LogItem)
		if !ok {
			continue
		}
		num++
		err := walk(log)
		if err != nil {
			return err
		}
	}
	if it.Error() != nil {
		return errors.NewInternalServerError(it.Error())
	}
	return nil
}

func (s *logQueryService) getIterator(ctx context.Context, sel *storage.Selector, live bool) (storekit.Iterator, error) {
	if sel.Scheme != "container" || !live {
		if s.storageReader != nil && (sel.Start > s.startTime || s.frozenStorageReader == nil) {
			return s.storageReader.Iterator(ctx, sel)
		}
		return s.tryGetIterator(ctx, sel, s.storageReader, s.frozenStorageReader)
	}
	if sel.End >= time.Now().Add(-24*time.Hour).UnixNano() {
		return s.tryGetIterator(ctx, sel, s.k8sReader, s.storageReader, s.frozenStorageReader)
	}
	return s.tryGetIterator(ctx, sel, s.storageReader, s.frozenStorageReader)
}

func (s *logQueryService) tryGetIterator(ctx context.Context, sel *storage.Selector, storages ...storage.Storage) (it storekit.Iterator, err error) {
	var its []storekit.Iterator
	for _, stor := range storages {
		if stor == nil {
			continue
		}
		it, _err := stor.Iterator(ctx, sel)
		if err != nil {
			s.p.Log.Errorf("failed to create %T.Iterator: %s", stor, _err)
			err = _err
			continue
		}
		its = append(its, it)
	}
	if len(its) == 0 {
		if err != nil {
			return nil, err
		}
		return storekit.EmptyIterator{}, nil
	} else if len(its) == 1 {
		return its[0], nil
	}
	return storekit.MergedHeadOverlappedIterator(storage.DefaultComparer, its...), nil
}

// Request .
type Request interface {
	GetStart() int64
	GetEnd() int64
	GetOffset() int64
	GetCount() int64
	GetPattern() string

	GetRequestId() string

	GetId() string
	GetSource() string
	GetStream() string
	GetLive() bool
	GetDebug() bool
}

const (
	defaultQueryCount = 50
	maxQueryCount     = 700
)

func getLimit(count int64) int {
	count = absInt(count)
	if count == 0 {
		return defaultQueryCount
	} else if count > maxQueryCount {
		return maxQueryCount
	}
	return int(count)
}

func absInt(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func toQuerySelector(req Request) (*storage.Selector, error) {
	sel := &storage.Selector{
		Start: req.GetStart(),
		End:   req.GetEnd(),
		Debug: req.GetDebug(),
	}

	if sel.End <= 0 {
		sel.End = time.Now().UnixNano()
	}
	if sel.Start < 0 {
		sel.Start = 0
	} else if sel.Start > 0 && req.GetCount() >= 0 {
		// avoid duplicating previous log
		// TODO: check by offset
		sel.Start++
	}

	if sel.End < sel.Start {
		return nil, errors.NewInvalidParameterError("(start,end]", "start must be less than end")
	}

	if len(req.GetRequestId()) > 0 {
		sel.Scheme = "trace"
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "tags.request_id",
			Op:    storage.EQ,
			Value: req.GetRequestId(),
		})
	} else if len(req.GetId()) > 0 {
		sel.Scheme = req.GetSource()
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "id",
			Op:    storage.EQ,
			Value: req.GetId(),
		})
		if len(req.GetSource()) > 0 {
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "source",
				Op:    storage.EQ,
				Value: req.GetSource(),
			})
		}
		if len(req.GetStream()) > 0 {
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "stream",
				Op:    storage.EQ,
				Value: req.GetStream(),
			})
		}
	} else {
		return nil, errors.NewMissingParameterError("id")
	}
	if len(req.GetPattern()) > 0 {
		_, err := regexp.Compile(req.GetPattern())
		if err != nil {
			return nil, errors.NewInvalidParameterError("pattern", err.Error())
		}
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "content",
			Op:    storage.REGEXP,
			Value: req.GetPattern(),
		})
	}
	return sel, nil
}

func toLogItems(ctx context.Context, it storekit.Iterator, forward bool, limit int) (list []*pb.LogItem, err error) {
	if limit <= 0 {
		return nil, nil
	}
	if forward {
		for it.Next() {
			log, ok := it.Value().(*pb.LogItem)
			if !ok {
				continue
			}
			list = append(list, log)
			if len(list) >= limit {
				break
			}
		}
	} else {
		for it.Prev() {
			log, ok := it.Value().(*pb.LogItem)
			if !ok {
				continue
			}
			list = append(list, log)
			if len(list) >= limit {
				break
			}
		}
		sort.Sort(storage.Logs(list))
	}
	return list, it.Error()
}
