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
	"fmt"
	"regexp"
	"sort"
	"time"

	linq "github.com/ahmetb/go-linq/v3"
	"github.com/mohae/deepcopy"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/pkg/common/apis"
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
	items, _, err := s.queryLogItems(ctx, req, func(sel *storage.Selector) *storage.Selector {
		return s.tryFillQueryMeta(ctx, sel)
	}, true, false)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogResponse{Lines: items}, nil
}

func (s *logQueryService) GetLogByRuntime(ctx context.Context, req *pb.GetLogByRuntimeRequest) (*pb.GetLogByRuntimeResponse, error) {
	if len(req.ApplicationId) <= 0 {
		return nil, errors.NewMissingParameterError("applicationId")
	}
	items, _, err := s.queryLogItems(ctx, req, func(sel *storage.Selector) *storage.Selector {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "tags.dice_application_id",
			Op:    storage.EQ,
			Value: req.ApplicationId,
		})
		s.tryFillQueryMeta(ctx, sel)
		return sel
	}, true, false)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogByRuntimeResponse{Lines: items}, nil
}

func (s *logQueryService) GetLogByOrganization(ctx context.Context, req *pb.GetLogByOrganizationRequest) (*pb.GetLogByOrganizationResponse, error) {
	if len(req.ClusterName) <= 0 {
		return nil, errors.NewMissingParameterError("clusterName")
	}
	items, _, err := s.queryLogItems(ctx, req, func(sel *storage.Selector) *storage.Selector {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "tags.dice_cluster_name",
			Op:    storage.EQ,
			Value: req.ClusterName,
		})
		s.tryFillQueryMeta(ctx, sel)
		return sel
	}, true, false)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogByOrganizationResponse{Lines: items}, nil
}

func (s *logQueryService) GetLogByExpression(ctx context.Context, req *pb.GetLogByExpressionRequest) (*pb.GetLogByExpressionResponse, error) {
	items, total, err := s.queryLogItems(ctx, req, nil, false, true)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogByExpressionResponse{Lines: items, Total: total}, nil
}

func (s *logQueryService) LogAggregation(ctx context.Context, req *pb.LogAggregationRequest) (*pb.LogAggregationResponse, error) {
	if s.storageReader == nil {
		return nil, fmt.Errorf("elasticsearch storage provider is required")
	}
	aggregator, ok := s.storageReader.(storage.Aggregator)
	if !ok {
		return nil, fmt.Errorf("%T not implment %T", s.storageReader, aggregator)
	}

	aggReq, err := s.toAggregation(req)
	if err != nil {
		return nil, err
	}
	aggResp, err := aggregator.Aggregate(ctx, aggReq)
	if err != nil {
		return nil, err
	}

	aggregations := map[string]*pb.AggregationResult{}
	for name, agg := range aggResp.Aggregations {
		var buckets []*pb.AggregationBucket
		linq.From(agg.Buckets).Select(func(item interface{}) interface{} {
			bucket := item.(*storage.AggregationBucket)
			key, _ := structpb.NewValue(bucket.Key)
			return &pb.AggregationBucket{
				Key:   key,
				Count: bucket.Count,
			}
		}).ToSlice(&buckets)
		aggregations[name] = &pb.AggregationResult{Buckets: buckets}
	}
	result := &pb.LogAggregationResponse{
		Total:        aggResp.Total,
		Aggregations: aggregations,
	}
	return result, nil
}

func (s *logQueryService) ScanLogsByExpression(req *pb.GetLogByExpressionRequest, stream pb.LogQueryService_ScanLogsByExpressionServer) error {
	return s.walkLogItems(context.Background(), req, nil, func(item *pb.LogItem) error {
		return stream.Send(item)
	})
}

func (s *logQueryService) tryFillQueryMeta(ctx context.Context, sel *storage.Selector, orgNames ...string) *storage.Selector {
	if len(sel.Meta.OrgNames) > 0 {
		return sel
	}
	orgNameList := []string{""}
	if reqOrg := apis.GetHeader(ctx, "org"); len(reqOrg) > 0 {
		orgNameList = append(orgNameList, reqOrg)
	}
	contains := func(orgName string) bool {
		for _, item := range orgNameList {
			if item == orgName {
				return true
			}
		}
		return false
	}
	for _, orgName := range orgNames {
		if contains(orgName) {
			continue
		}
		orgNameList = append(orgNameList, orgName)
	}
	sel.Meta.OrgNames = orgNameList
	return sel
}

func (s *logQueryService) toAggregation(req *pb.LogAggregationRequest) (*storage.Aggregation, error) {
	if req.Query == nil {
		return nil, fmt.Errorf("query should not be nil")
	}
	if len(req.Aggregations) == 0 {
		return nil, fmt.Errorf("aggregations should not be empty")
	}
	agg := &storage.Aggregation{}
	sel, err := toQuerySelector(req.Query)
	if err != nil {
		return nil, err
	}
	agg.Selector = sel

	for _, descriptor := range req.Aggregations {
		aggDesc := &storage.AggregationDescriptor{
			Name:  descriptor.Name,
			Field: descriptor.Field,
		}
		agg.Aggs = append(agg.Aggs, aggDesc)
		switch descriptor.Type {
		case pb.AggregationType_Histogram:
			var histogramOptions pb.HistogramAggOptions
			err = descriptor.Options.UnmarshalTo(&histogramOptions)
			if err != nil {
				return nil, err
			}
			aggDesc.Typ = storage.AggregationHistogram
			aggDesc.Options = storage.HistogramAggOptions{
				MinimumInterval: histogramOptions.MinimumInterval,
				PreferredPoints: histogramOptions.PreferredPoints,
				FixedInterval:   histogramOptions.FixedInterval,
			}
		case pb.AggregationType_Terms:
			var termsOptions pb.TermsAggOptions
			err = descriptor.Options.UnmarshalTo(&termsOptions)
			if err != nil {
				return nil, err
			}
			aggDesc.Typ = storage.AggregationTerms
			aggDesc.Options = storage.TermsAggOptions{
				Size:    termsOptions.Size,
				Missing: termsOptions.Missing.AsInterface(),
			}
		}
	}

	return agg, nil
}

func (s *logQueryService) queryLogItems(ctx context.Context, req Request, fn func(sel *storage.Selector) *storage.Selector, ascendingResult bool, withTotal bool) ([]*pb.LogItem, int64, error) {
	sel, err := toQuerySelector(req)
	if err != nil {
		return nil, 0, err
	}
	if fn != nil {
		sel = fn(sel)
	}
	var it storekit.Iterator
	if withTotal {
		it, err = s.getIterator(ctx, sel, req.GetLive())
	} else {
		it, err = s.getOrderedIterator(ctx, s.splitSelectors(sel, 24*time.Hour, 10), req.GetLive())
	}

	if err != nil {
		return nil, 0, errors.NewInternalServerError(err)
	}
	defer it.Close()

	items, err := toLogItems(ctx, it, req.GetCount() >= 0, getLimit(req.GetCount()), ascendingResult)
	if err != nil {
		return nil, 0, errors.NewInternalServerError(err)
	}
	if !withTotal {
		return items, 0, nil
	}
	counter, ok := it.(storekit.Counter)
	if !ok {
		return items, 0, fmt.Errorf("failed to get total: %T not implement %T", it, counter)
	}
	total, err := counter.Total()
	if err != nil {
		return items, 0, errors.NewInternalServerError(err)
	}
	return items, total, nil
}

func (s *logQueryService) walkLogItems(ctx context.Context, req Request, fn func(sel *storage.Selector) (*storage.Selector, error), walk func(item *pb.LogItem) error) error {
	//if req.GetCount() < 0 {
	//	return errors.NewInvalidParameterError("count", "not allowed negative")
	//}
	sel, err := toQuerySelector(req)
	if err != nil {
		return err
	}
	if fn != nil {
		sel, err = fn(sel)
		if err != nil {
			return err
		}
	}
	it, err := s.getIterator(ctx, sel, req.GetLive())
	if err != nil {
		return errors.NewInternalServerError(err)
	}
	defer it.Close()

	next := it.Next
	if req.GetCount() < 0 {
		next = it.Prev
	}

	for next() {
		log, ok := it.Value().(*pb.LogItem)
		if !ok {
			continue
		}
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

func (s *logQueryService) splitSelectors(sel *storage.Selector, interval time.Duration, maxSlices int) []*storage.Selector {
	var sels []*storage.Selector
	end := sel.End
	count := 0

	for end > sel.Start {
		start := end - int64(interval)
		if start < sel.Start || (count+1) >= maxSlices {
			start = sel.Start
		}

		subSel := deepcopy.Copy(sel).(*storage.Selector)
		subSel.Start = start
		subSel.End = end
		sels = append(sels, subSel)
		count++

		end = start
	}

	length := len(sels)
	if length <= 1 {
		return sels
	}
	reversed := make([]*storage.Selector, length)
	for i := 0; i < length; i++ {
		reversed[i] = sels[length-1-i]
	}

	return reversed
}

func (s *logQueryService) getOrderedIterator(ctx context.Context, sels []*storage.Selector, live bool) (storekit.Iterator, error) {
	var its []storekit.Iterator
	for _, item := range sels {
		it, err := s.getIterator(ctx, item, live)
		if err != nil {
			return nil, err
		}
		its = append(its, it)
	}

	return storekit.OrderedIterator(its...), nil
}

func (s *logQueryService) getIterator(ctx context.Context, sel *storage.Selector, live bool) (storekit.Iterator, error) {
	if sel.Scheme == "advanced" {
		if s.storageReader == nil {
			return storekit.EmptyIterator{}, nil
		}
		return s.storageReader.Iterator(ctx, sel)
	}
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
		if _err != nil {
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
	GetCount() int64
	GetLive() bool
	GetDebug() bool
}

type ByContainerIdRequest interface {
	Request
	GetOffset() int64
	GetPattern() string
	GetRequestId() string
	GetId() string
	GetSource() string
	GetStream() string
}

type ByExpressionRequest interface {
	Request
	GetQueryExpression() string
	GetQueryMeta() *pb.QueryMeta
	GetExtraFilter() *pb.ExtraFilter
}

const (
	defaultQueryCount     = 50
	maxQueryCount         = 700
	maxTimeRange          = 30 * 24 * int64(time.Hour)
	defaultQueryTimeRange = 7 * 24 * int64(time.Hour)
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
	if sel.Start <= 0 {
		sel.Start = sel.End - defaultQueryTimeRange
		if sel.Start < 0 {
			sel.Start = 0
		}
	} else if sel.Start > 0 && req.GetCount() >= 0 {
		// avoid duplicating previous log
		// TODO: check by offset
		sel.Start++
	}

	if sel.End < sel.Start {
		return nil, errors.NewInvalidParameterError("(start,end]", "start must be less than end")
	} else if sel.End-sel.Start > maxTimeRange {
		if queryMeta, ok := req.(ByExpressionRequest); !ok || queryMeta.GetQueryMeta() != nil && !queryMeta.GetQueryMeta().GetIgnoreMaxTimeRangeLimit() {
			return nil, errors.NewInvalidParameterError("(start,end]", "time range is too large")
		}
	}

	if byContainerIdRequest, ok := req.(ByContainerIdRequest); ok {
		if len(byContainerIdRequest.GetRequestId()) > 0 {
			sel.Scheme = "trace"
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "tags.request_id",
				Op:    storage.EQ,
				Value: byContainerIdRequest.GetRequestId(),
			})
		} else if len(byContainerIdRequest.GetId()) > 0 {
			sel.Scheme = byContainerIdRequest.GetSource()
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "id",
				Op:    storage.EQ,
				Value: byContainerIdRequest.GetId(),
			})
			if len(byContainerIdRequest.GetSource()) > 0 {
				sel.Filters = append(sel.Filters, &storage.Filter{
					Key:   "source",
					Op:    storage.EQ,
					Value: byContainerIdRequest.GetSource(),
				})
			}
			if len(byContainerIdRequest.GetStream()) > 0 {
				sel.Filters = append(sel.Filters, &storage.Filter{
					Key:   "stream",
					Op:    storage.EQ,
					Value: byContainerIdRequest.GetStream(),
				})
			}
		} else {
			return nil, errors.NewMissingParameterError("id")
		}
		if len(byContainerIdRequest.GetPattern()) > 0 {
			_, err := regexp.Compile(byContainerIdRequest.GetPattern())
			if err != nil {
				return nil, errors.NewInvalidParameterError("pattern", err.Error())
			}
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "content",
				Op:    storage.REGEXP,
				Value: byContainerIdRequest.GetPattern(),
			})
		}
	}

	if byExpressionRequest, ok := req.(ByExpressionRequest); ok {
		sel.Scheme = "advanced"
		if expr := byExpressionRequest.GetQueryExpression(); len(expr) > 0 {
			sel.Filters = append(sel.Filters, &storage.Filter{
				Key:   "_",
				Op:    storage.EXPRESSION,
				Value: expr,
			})
		}
		if extraFilter := byExpressionRequest.GetExtraFilter(); extraFilter != nil {
			if after := byExpressionRequest.GetExtraFilter().GetAfter(); after != nil {
				sel.Skip.AfterId = &storage.UniqueId{Id: after.Id, Timestamp: after.UnixNano, Offset: after.Offset}
			}
			if offset := byExpressionRequest.GetExtraFilter().GetPositionOffset(); offset > 0 {
				sel.Skip.FromOffset = int(offset)
			}
		}
		if meta := byExpressionRequest.GetQueryMeta(); meta != nil {
			sel.Meta = storage.QueryMeta{
				OrgNames:              []string{meta.GetOrgName()},
				MspEnvIds:             meta.GetMspEnvIds(),
				Highlight:             meta.GetHighlight(),
				PreferredBufferSize:   int(meta.GetPreferredBufferSize()),
				PreferredIterateStyle: storage.IterateStyle(meta.PreferredIterateStyle),
			}
		}
	}

	return sel, nil
}

func toLogItems(ctx context.Context, it storekit.Iterator, forward bool, limit int, ascendingResult bool) (list []*pb.LogItem, err error) {
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
		if ascendingResult {
			sort.Sort(storage.Logs(list))
		}
	}
	return list, it.Error()
}
