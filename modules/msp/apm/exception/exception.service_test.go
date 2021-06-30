package exception

import (
	context "context"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/apm/exception/pb"
	reflect "reflect"
	testing "testing"
)

func Test_exceptionService_GetExceptions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExceptionsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExceptionsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.exception.ExceptionService",
		//			`
		//erda.msp.apm.exception:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExceptionsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExceptionsResponse{
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
			srv := hub.Service(tt.service).(pb.ExceptionServiceServer)
			got, err := srv.GetExceptions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("exceptionService.GetExceptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("exceptionService.GetExceptions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
