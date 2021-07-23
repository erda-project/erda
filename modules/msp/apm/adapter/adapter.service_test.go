package adapter

import (
	context "context"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
	reflect "reflect"
	testing "testing"
)

func Test_adapterService_GetAdapters(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetAdapterRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetAdapterResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.adapter.AdapterService",
		//			`
		//erda.msp.apm.adapter:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetAdapterRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetAdapterResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
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
			srv := hub.Service(tt.service).(pb.AdapterServiceServer)
			got, err := srv.GetAdapters(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapterService.GetAdapters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("adapterService.GetAdapters() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
