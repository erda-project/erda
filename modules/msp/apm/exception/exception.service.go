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

package exception

import (
	context "context"
	"encoding/json"
	"time"

	"github.com/recallsong/go-utils/conv"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
)

type exceptionService struct {
	p *provider
}

func (s *exceptionService) GetExceptions(ctx context.Context, req *pb.GetExceptionsRequest) (*pb.GetExceptionsResponse, error) {

	iter := s.p.cassandraSession.Query("SELECT * FROM error_description_v2 where terminus_key=? ALLOW FILTERING", req.ScopeID).Iter()

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
		iterCount := s.p.cassandraSession.Query(stat, exception.Id, req.StartTime*1e6, req.EndTime*1e6).Iter()
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

	return &pb.GetExceptionsResponse{Data: exceptions}, nil
}

func (s *exceptionService) GetExceptionEventIds(ctx context.Context, req *pb.GetExceptionEventIdsRequest) (*pb.GetExceptionEventIdsResponse, error) {
	iter := s.p.cassandraSession.Query("SELECT event_id FROM error_event_mapping WHERE error_id= ? limit ?", req.ExceptionID, 999).Iter()

	var data []string
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		data = append(data, conv.ToString(row["event_id"]))
	}
	return &pb.GetExceptionEventIdsResponse{Data: data}, nil
}

func (s *exceptionService) GetExceptionEvent(ctx context.Context, req *pb.GetExceptionEventRequest) (*pb.GetExceptionEventResponse, error) {
	iter := s.p.cassandraSession.Query("SELECT * FROM error_events WHERE event_id = ?", req.ExceptionEventID).Iter()
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
	return &pb.GetExceptionEventResponse{Data: &event}, nil
}
