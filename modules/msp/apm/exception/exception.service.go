package exception

import (
	context "context"
	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	"github.com/gocql/gocql"
	"github.com/recallsong/go-utils/conv"
	"time"
)

type exceptionService struct {
	p *provider
}

func (s *exceptionService) GetExceptions(ctx context.Context, req *pb.GetExceptionsRequest) (*pb.GetExceptionsResponse, error) {

	iter := s.p.cassandraSession.Query("SELECT * FROM error_description_v2 where terminus_key=? ALLOW FILTERING", req.GetScopeId()).
		Consistency(gocql.All).
		RetryPolicy(nil).
		Iter()

	var exceptions []*pb.Exception
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		exception := pb.Exception{}
		tags := row["tags"].(map[string]string)
		exception.Id = row["error_id"].(string)
		exception.ScopeId = conv.ToString(row["terminus_key"])
		exception.ClassName = conv.ToString(tags["class"])
		exception.Method = conv.ToString(tags["method"])
		exception.Type = conv.ToString(tags["type"])
		exception.ExceptionMessage = conv.ToString(tags["exception_message"])
		exception.File = conv.ToString(tags["file"])
		exception.ServiceName = conv.ToString(tags["service_name"])
		exception.ApplicationId = conv.ToString(tags["application_id"])
		exception.RuntimeId = conv.ToString(tags["runtime_id"])
		layout := "2006-01-02 15:04:05"

		stat := "SELECT timestamp,count FROM error_count WHERE error_id= ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC LIMIT 1"
		iterCount := s.p.cassandraSession.Query(stat, exception.Id, req.StartTime*1e6, req.EndTime*1e6).
			Consistency(gocql.All).
			RetryPolicy(nil).
			Iter()
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
