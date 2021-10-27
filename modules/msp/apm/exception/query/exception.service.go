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
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/recallsong/go-utils/conv"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	error_storage "github.com/erda-project/erda/modules/msp/apm/exception/erda-error/storage"
	event_storage "github.com/erda-project/erda/modules/msp/apm/exception/erda-event/storage"
	"github.com/erda-project/erda/pkg/common/errors"
)

type exceptionService struct {
	p                  *provider
	ErrorStorageReader error_storage.Storage
	EventStorageReader event_storage.Storage
}

func (s *exceptionService) GetExceptions(ctx context.Context, req *pb.GetExceptionsRequest) (*pb.GetExceptionsResponse, error) {
	var exceptions []*pb.Exception

	if strings.Contains(s.p.Cfg.QuerySource, "cassandra") {
		// do cassandra query
		exceptionsFromCassandra := fetchErdaErrorFromCassandra(s.p.cassandraSession.Session(), req)
		for _, exception := range exceptionsFromCassandra {
			exceptions = append(exceptions, exception)
		}
	}

	if strings.Contains(s.p.Cfg.QuerySource, "cassandra") {
		// do es query
		exceptionsFromElasticsearch, _ := fetchErdaErrorFromES(ctx, s.ErrorStorageReader, s.EventStorageReader, req, true, 1000)
		for _, exception := range exceptionsFromElasticsearch {
			exceptions = append(exceptions, exception)
		}
	}

	return &pb.GetExceptionsResponse{Data: exceptions}, nil
}

func fetchErdaErrorFromCassandra(session *gocql.Session, req *pb.GetExceptionsRequest) []*pb.Exception {
	iter := session.Query("SELECT * FROM error_description_v2 where terminus_key=? ALLOW FILTERING", req.ScopeID).Iter()

	var exceptions []*pb.Exception
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		exception := pb.Exception{}
		tags := row["tags"].(map[string]string)
		exception.Id = row["error_id"].(string)
		exception.ScopeID = conv.ToString(row["terminus_key"])
		exception.ClassName = conv.ToString(tags["class"])
		exception.Method = conv.ToString(tags["method"])
		exception.Type = conv.ToString(tags["type"])
		exception.ExceptionMessage = conv.ToString(tags["exception_message"])
		exception.File = conv.ToString(tags["file"])
		exception.ServiceName = conv.ToString(tags["service_name"])
		exception.ApplicationID = conv.ToString(tags["application_id"])
		exception.RuntimeID = conv.ToString(tags["runtime_id"])
		layout := "2006-01-02 15:04:05"

		stat := "SELECT timestamp,count FROM error_count WHERE error_id= ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC"
		iterCount := session.Query(stat, exception.Id, req.StartTime*1e6, req.EndTime*1e6).Iter()
		count := int64(0)
		index := 0
		for {
			rowCount := make(map[string]interface{})
			if !iterCount.MapScan(rowCount) {
				break
			}
			if index == 0 {
				exception.CreateTime = time.Unix(conv.ToInt64(rowCount["timestamp"], 0)/1e9, 10).Format(layout)
			}
			count += conv.ToInt64(rowCount["count"], 0)
			index++
			if index == iterCount.NumRows() {
				exception.UpdateTime = time.Unix(conv.ToInt64(rowCount["timestamp"], 0)/1e9, 10).Format(layout)
			}
		}
		exception.EventCount = count
		if exception.EventCount > 0 {
			exceptions = append(exceptions, &exception)
		}
	}
	return exceptions
}

func (s *exceptionService) GetExceptionEventIds(ctx context.Context, req *pb.GetExceptionEventIdsRequest) (*pb.GetExceptionEventIdsResponse, error) {
	var data []string

	if strings.Contains(s.p.Cfg.QuerySource, "cassandra") {
		// do cassandra query
		iter := s.p.cassandraSession.Session().Query("SELECT event_id FROM error_event_mapping WHERE error_id= ? limit ?", req.ExceptionID, 999).Iter()

		for {
			row := make(map[string]interface{})
			if !iter.MapScan(row) {
				break
			}
			data = append(data, conv.ToString(row["event_id"]))
		}
	}
	if strings.Contains(s.p.Cfg.QuerySource, "elasticsearch") {
		//do es query
		sel := &event_storage.Selector{
			StartTime: time.Now().Add(-time.Hour * 24 * 7).UnixNano(),
			EndTime:   time.Now().UnixNano(),
			ErrorId:   req.ExceptionID,
		}
		items, err := fetchErdaEventFromES(ctx, s.EventStorageReader, sel, true, 999)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}

		for _, value := range items {
			data = append(data, value.EventId)
		}
	}

	return &pb.GetExceptionEventIdsResponse{Data: data}, nil
}

func (s *exceptionService) GetExceptionEvent(ctx context.Context, req *pb.GetExceptionEventRequest) (*pb.GetExceptionEventResponse, error) {
	event := pb.ExceptionEvent{}

	if strings.Contains(s.p.Cfg.QuerySource, "elasticsearch") {
		// do es query
		sel := &event_storage.Selector{
			StartTime:   time.Now().Add(-time.Hour * 24 * 7).UnixNano(),
			EndTime:     time.Now().UnixNano(),
			EventId:     req.ExceptionEventID,
			TerminusKey: req.ScopeID,
		}
		items, err := fetchErdaEventFromES(ctx, s.EventStorageReader, sel, true, 1)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}

		if len(items) > 0 {
			item := items[0]
			event.Tags = item.Tags
			event.Id = item.EventId
			event.ExceptionID = item.ErrorId
			event.RequestID = item.RequestId
			event.RequestSampled = conv.ToBool(event.Tags["request_sampled"], false)
			event.Metadata = item.MetaData
			event.RequestContext = item.RequestContext
			event.RequestHeaders = item.RequestHeaders
			event.Timestamp = item.Timestamp / int64(time.Millisecond)
			var stacks []*pb.Stacks
			for _, info := range item.Stacks {
				var stack pb.Stacks
				var stackMap map[string]*structpb.Value
				if err := json.Unmarshal([]byte(info), &stackMap); err != nil {
					continue
				}

				stack.Stack = stackMap
				stacks = append(stacks, &stack)
			}
			event.Stacks = stacks
		}
	}

	if strings.Contains(s.p.Cfg.QuerySource, "cassandra") && len(event.Id) <= 0 {
		// do cassandra query
		event = fetchErdaEventFromCassandra(s.p.cassandraSession.Session(), req)
	}

	return &pb.GetExceptionEventResponse{Data: &event}, nil
}

func fetchErdaEventFromCassandra(session *gocql.Session, req *pb.GetExceptionEventRequest) pb.ExceptionEvent {
	iter := session.Query("SELECT * FROM error_events WHERE event_id = ?", req.ExceptionEventID).Iter()
	event := pb.ExceptionEvent{}
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		event.Tags = row["tags"].(map[string]string)
		if conv.ToString(event.Tags["terminus_key"]) != req.ScopeID {
			continue
		}
		event.Id = conv.ToString(row["event_id"])
		event.ExceptionID = conv.ToString(row["error_id"])
		event.RequestID = conv.ToString(row["request_id"])
		event.RequestSampled = conv.ToBool(event.Tags["request_sampled"], false)
		event.Metadata = row["meta_data"].(map[string]string)
		event.RequestContext = row["request_context"].(map[string]string)
		event.RequestHeaders = row["request_headers"].(map[string]string)
		event.Timestamp = row["timestamp"].(int64) / int64(time.Millisecond)
		var stacks []*pb.Stacks
		for _, info := range row["stacks"].([]string) {
			var stack pb.Stacks
			var stackMap map[string]*structpb.Value
			if err := json.Unmarshal([]byte(info), &stackMap); err != nil {
				continue
			}
			stack.Stack = stackMap
			stacks = append(stacks, &stack)
		}
		event.Stacks = stacks
	}
	return event
}

func fetchErdaErrorFromES(ctx context.Context, errorStorage error_storage.Storage, eventStorage event_storage.Storage, req *pb.GetExceptionsRequest, forward bool, limit int) (exceptions []*pb.Exception, err error) {
	sel := error_storage.Selector{
		StartTime:   req.StartTime * 1e6,
		EndTime:     req.EndTime * 1e6,
		TerminusKey: req.ScopeID,
	}

	it, err := errorStorage.Iterator(ctx, &sel)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer it.Close()

	var list []*exception.Erda_error
	if forward {
		for it.Next() {
			if len(list) >= limit {
				break
			}
			e, ok := it.Value().(*exception.Erda_error)
			if !ok {
				continue
			}
			list = append(list, e)
		}
	} else {
		for it.Prev() {
			e, ok := it.Value().(*exception.Erda_error)
			if !ok {
				continue
			}
			if len(list) >= limit {
				break
			}
			list = append(list, e)
		}
		sort.Sort(ErdaErrors(list))
	}

	for _, value := range list {
		exception := pb.Exception{}
		tags := value.Tags
		exception.Id = value.ErrorId
		exception.ScopeID = value.TerminusKey
		exception.ClassName = conv.ToString(tags["class"])
		exception.Method = conv.ToString(tags["method"])
		exception.Type = conv.ToString(tags["type"])
		exception.ExceptionMessage = conv.ToString(tags["exception_message"])
		exception.File = conv.ToString(tags["file"])
		exception.ServiceName = conv.ToString(tags["service_name"])
		exception.ApplicationID = conv.ToString(tags["application_id"])
		exception.RuntimeID = conv.ToString(tags["runtime_id"])
		layout := "2006-01-02 15:04:05"

		items, err := fetchErdaEventFromES(ctx, eventStorage, &event_storage.Selector{
			StartTime: req.StartTime * 1e6,
			EndTime:   req.EndTime * 1e6,
			ErrorId:   value.ErrorId,
		}, true, 999)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		count := int64(0)
		for index, item := range items {
			if index == 0 {
				exception.CreateTime = time.Unix(item.Timestamp/1e9, 10).Format(layout)
			}
			if index == len(items)-1 {
				exception.UpdateTime = time.Unix(item.Timestamp/1e9, 10).Format(layout)
			}
			count++
		}
		exception.EventCount = count
		if exception.EventCount > 0 {
			exceptions = append(exceptions, &exception)
		}
	}

	return exceptions, it.Error()
}

type ErdaErrors []*exception.Erda_error

func (s ErdaErrors) Len() int      { return len(s) }
func (s ErdaErrors) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ErdaErrors) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

func fetchErdaEventFromES(ctx context.Context, storage event_storage.Storage, sel *event_storage.Selector, forward bool, limit int) (list []*exception.Erda_event, err error) {

	it, err := storage.Iterator(ctx, sel)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	defer it.Close()

	if forward {
		for it.Next() {
			if len(list) >= limit {
				return list, nil
			}
			e, ok := it.Value().(*exception.Erda_event)
			if !ok {
				continue
			}
			list = append(list, e)
		}
	} else {
		for it.Prev() {
			e, ok := it.Value().(*exception.Erda_event)
			if !ok {
				continue
			}
			if len(list) >= limit {
				return list, nil
			}
			list = append(list, e)
		}
		sort.Sort(ErdaEvents(list))
	}

	return list, it.Error()
}

type ErdaEvents []*exception.Erda_event

func (s ErdaEvents) Len() int      { return len(s) }
func (s ErdaEvents) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ErdaEvents) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}
