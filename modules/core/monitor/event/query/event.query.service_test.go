package query

import (
	context "context"
	reflect "reflect"
	testing "testing"

	"github.com/erda-project/erda/modules/core/monitor/event"
	"github.com/golang/mock/gomock"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/monitor/event/pb"
)

// -go:generate mockgen -destination=./mock_storage.go -package query -source=../storage/storage.go Storage
func Test_eventQueryService_GetEvents(t *testing.T) {

	type args struct {
		ctx context.Context
		req *pb.GetEventsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetEventsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//{
		//	"case 1",
		//	"erda.core.monitor.event.EventQueryService",
		//	`
		//	erda.core.monitor.event.query:
		//	`,
		//	args{
		//		context.TODO(),
		//		&pb.GetEventsRequest{
		//			// TODO: setup fields
		//		},
		//	},
		//	&pb.GetEventsResponse{
		//		// TODO: setup fields.
		//	},
		//	false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.EventQueryServiceServer)
			got, err := srv.GetEvents(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventQueryService.GetEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("eventQueryService.GetEvents() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := NewMockStorage(ctrl)
	storage.EXPECT().
		QueryPaged(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*event.Event{
			{EventID: "event-id-1"},
		}, nil)

	querySvc := &eventQueryService{
		storageReader: storage,
	}
	result, err := querySvc.GetEvents(context.Background(), &pb.GetEventsRequest{
		Start:        1,
		End:          2,
		TraceId:      "trace-id",
		RelationId:   "res-id",
		RelationType: "res-type",
	})
	if err != nil {
		t.Errorf("should not throw error")
	}
	if result == nil || len(result.Data.Items) != 1 {
		t.Errorf("assert result failed")
	}

}
