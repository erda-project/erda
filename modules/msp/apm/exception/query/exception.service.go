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
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/recallsong/go-utils/conv"
	"google.golang.org/protobuf/types/known/structpb"

	eventpb "github.com/erda-project/erda-proto-go/core/monitor/event/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	entitypb "github.com/erda-project/erda-proto-go/oap/entity/pb"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	"github.com/erda-project/erda/pkg/common/errors"
)

type exceptionService struct {
	p      *provider
	Metric metricpb.MetricServiceServer
	Event  eventpb.EventQueryServiceServer
	Entity entitypb.EntityServiceServer
}

func (s *exceptionService) GetExceptions(ctx context.Context, req *pb.GetExceptionsRequest) (*pb.GetExceptionsResponse, error) {
	var exceptions []*pb.Exception

	if strings.Contains(s.p.Cfg.QuerySource, "cassandra") {
		// do cassandra query
		exceptionsFromCassandra := fetchErdaErrorFromCassandra(ctx, s.Metric, s.p.cassandraSession.Session(), req)
		for _, exception := range exceptionsFromCassandra {
			exceptions = append(exceptions, exception)
		}
	}

	if strings.Contains(s.p.Cfg.QuerySource, "elasticsearch") {
		// do es query
		conditions := map[string]string{
			"terminusKey": req.ScopeID,
		}

		entityReq := &entitypb.ListEntitiesRequest{
			Type:   "error_exception",
			Labels: conditions,
			Limit:  int64(1000),
		}
		exceptionsFromElasticsearch, _ := fetchErdaErrorFromES(ctx, s.Event, s.Entity, entityReq, req.StartTime, req.EndTime)
		for _, exception := range exceptionsFromElasticsearch {
			exceptions = append(exceptions, exception)
		}
	}

	return &pb.GetExceptionsResponse{Data: exceptions}, nil
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
		tags := map[string]string{
			"terminusKey": req.ScopeID,
		}
		eventReq := &eventpb.GetEventsRequest{
			RelationId:   req.ExceptionID,
			RelationType: "exception",
			Tags:         tags,
			PageNo:       1,
			PageSize:     999,
			Start:        time.Now().Add(-time.Hour * 24 * 7).UnixNano(),
			End:          time.Now().UnixNano(),
		}

		items, err := fetchErdaEventFromES(ctx, s.Event, eventReq)
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

		tags := map[string]string{
			"terminusKey": req.ScopeID,
		}
		eventReq := &eventpb.GetEventsRequest{
			EventId:  req.ExceptionEventID,
			Tags:     tags,
			PageNo:   1,
			PageSize: 999,
			Start:    time.Now().Add(-time.Hour * 24 * 7).UnixNano(),
			End:      time.Now().UnixNano(),
			Debug:    true,
		}
		items, err := fetchErdaEventFromES(ctx, s.Event, eventReq)
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

func fetchErdaErrorFromCassandra(ctx context.Context, metric metricpb.MetricServiceServer, session *gocql.Session, req *pb.GetExceptionsRequest) []*pb.Exception {
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

		fetchErdaErrorEventCount(ctx, metric, session, req, &exception)
		if exception.EventCount > 0 {
			exceptions = append(exceptions, &exception)
		}
	}
	return exceptions
}

func fetchErdaErrorEventCount(ctx context.Context, metric metricpb.MetricServiceServer, session *gocql.Session, req *pb.GetExceptionsRequest, exception *pb.Exception) {
	layout := "2006-01-02 15:04:05"
	count := int64(0)

	stat := "SELECT timestamp,count FROM error_count WHERE error_id= ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC"
	iterCount := session.Query(stat, exception.Id, req.StartTime*1e6, req.EndTime*1e6).Iter()
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

	metricreq := &metricpb.QueryWithInfluxFormatRequest{
		Start:     conv.ToString(req.StartTime * 1e6), // or timestamp
		End:       conv.ToString(req.EndTime * 1e6),   // or timestamp
		Statement: `SELECT timestamp, count::field FROM error_count WHERE error_id::tag=$error_id ORDER BY timestamp`,
		Params: map[string]*structpb.Value{
			"error_id": structpb.NewStringValue(exception.Id),
		},
	}

	resp, err := metric.QueryWithInfluxFormat(ctx, metricreq)
	if err != nil {
		exception.UpdateTime = exception.CreateTime
	} else {
		rows := resp.Results[0].Series[0].Rows
		for index, row := range rows {
			if index == 0 && exception.CreateTime == "" {
				exception.CreateTime = time.Unix(int64(row.Values[0].GetNumberValue())/1e9, 10).Format(layout)
			}
			if index == len(rows)-1 {
				exception.UpdateTime = time.Unix(int64(row.Values[0].GetNumberValue())/1e9, 10).Format(layout)
			}
			count = count + int64(row.Values[1].GetNumberValue())
		}
	}

	exception.EventCount = count
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

func fetchErdaErrorFromES(ctx context.Context, Event eventpb.EventQueryServiceServer, Entity entitypb.EntityServiceServer, req *entitypb.ListEntitiesRequest, startTime int64, endTime int64) (exceptions []*pb.Exception, err error) {
	listEntity, err := Entity.ListEntities(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	for _, value := range listEntity.Data.List {
		exception := pb.Exception{}
		tags := value.Values
		exception.Id = value.Key
		exception.ScopeID = value.Labels["terminusKey"]
		exception.ClassName = tags["class"].GetStringValue()
		exception.Method = tags["method"].GetStringValue()
		exception.Type = tags["type"].GetStringValue()
		exception.ExceptionMessage = tags["exception_message"].GetStringValue()
		exception.File = tags["file"].GetStringValue()
		exception.ServiceName = tags["service_name"].GetStringValue()
		exception.ApplicationID = tags["application_id"].GetStringValue()
		exception.RuntimeID = tags["runtime_id"].GetStringValue()
		layout := "2006-01-02 15:04:05"

		eventReq := &eventpb.GetEventsRequest{
			RelationId:   value.Key,
			RelationType: "exception",
			Start:        startTime * 1e6,
			End:          endTime * 1e6,
			PageNo:       1,
			PageSize:     10000,
			Debug:        false,
		}
		items, err := fetchErdaEventFromES(ctx, Event, eventReq)
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

	return exceptions, nil
}

type ErdaErrors []*exception.Erda_error

func (s ErdaErrors) Len() int      { return len(s) }
func (s ErdaErrors) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ErdaErrors) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

type ErdaEvents []*exception.Erda_event

func (s ErdaEvents) Len() int      { return len(s) }
func (s ErdaEvents) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ErdaEvents) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

func fetchErdaEventFromES(ctx context.Context, Event eventpb.EventQueryServiceServer, req *eventpb.GetEventsRequest) (list []*exception.Erda_event, err error) {
	eventsResp, err := Event.GetEvents(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	for _, item := range eventsResp.Data.Items {
		erdaEvent := &exception.Erda_event{}
		erdaEvent.EventId = item.EventID
		erdaEvent.Timestamp = int64(item.TimeUnixNano)
		erdaEvent.ErrorId = item.Relations.ResID
		erdaEvent.RequestId = item.Attributes["requestId"]

		var stacks []string
		json.Unmarshal([]byte(item.Message), &stacks)
		erdaEvent.Stacks = stacks

		tagsMap := make(map[string]string)
		json.Unmarshal([]byte(item.Attributes["tags"]), &tagsMap)
		erdaEvent.Tags = tagsMap

		requestContextMap := make(map[string]string)
		json.Unmarshal([]byte(item.Attributes["requestContext"]), &requestContextMap)
		erdaEvent.RequestContext = requestContextMap

		requestHeadersMap := make(map[string]string)
		json.Unmarshal([]byte(item.Attributes["requestHeaders"]), &requestHeadersMap)
		erdaEvent.RequestHeaders = requestHeadersMap

		metaDataMap := make(map[string]string)
		json.Unmarshal([]byte(item.Attributes["metaData"]), &metaDataMap)
		erdaEvent.MetaData = metaDataMap

		list = append(list, erdaEvent)
	}

	return list, nil
}
