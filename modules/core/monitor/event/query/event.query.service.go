package query

import (
	context "context"

	pb "github.com/erda-project/erda-proto-go/core/monitor/event/pb"
	commonPb "github.com/erda-project/erda-proto-go/oap/common/pb"
	oapPb "github.com/erda-project/erda-proto-go/oap/event/pb"
	"github.com/erda-project/erda/modules/core/monitor/event/storage"
)

type eventQueryService struct {
	p             *provider
	storageReader storage.Storage
}

func (s *eventQueryService) GetEvents(ctx context.Context, req *pb.GetEventsRequest) (*pb.GetEventsResponse, error) {
	sel := &storage.Selector{
		Start: req.Start,
		End:   req.End,
		Debug: req.Debug,
	}
	if len(req.EventId) > 0 {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "event_id",
			Op:    storage.EQ,
			Value: req.EventId,
		})
	}
	if len(req.RelationType) > 0 {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "relation_type",
			Op:    storage.EQ,
			Value: req.RelationType,
		})
	}
	if len(req.RelationId) > 0 {
		sel.Filters = append(sel.Filters, &storage.Filter{
			Key:   "relation_id",
			Op:    storage.EQ,
			Value: req.RelationId,
		})
	}
	list, err := s.storageReader.QueryPaged(ctx, sel, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	resp := &pb.GetEventsResponse{Data: &pb.GetEventsResult{}}
	for _, item := range list {
		data := &oapPb.Event{
			EventID:      item.EventID,
			Name:         item.Name,
			Kind:         oapPb.Event_EventKind(oapPb.Event_EventKind_value[item.Kind]),
			TimeUnixNano: uint64(item.Timestamp),
			Attributes:   item.Tags,
			Message:      item.Content,
		}
		if item.Relations != nil {
			data.Relations = &commonPb.Relation{
				ResID:   item.Relations.ResID,
				ResType: item.Relations.ResType,
				TraceID: item.Relations.TraceID,
			}
		}
		resp.Data.Items = append(resp.Data.Items, data)
	}

	return resp, nil
}
