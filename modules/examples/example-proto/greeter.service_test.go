package example

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/examples/pb"
)

func Test_greeterService_SayHello(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.HelloRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.HelloResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// {
		// 	"case 1",
		// 	"erda.example.GreeterService",
		// 	`erda.example:`,
		// 	args{
		// 		context.TODO(),
		// 		&pb.HelloRequest{
		// 			// TODO: setup fields
		// 		},
		// 	},
		// 	&pb.HelloResponse{
		// 		// TODO: setup fields.
		// 	},
		// 	false,
		// },
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
			srv := hub.Service(tt.service).(pb.GreeterServiceServer)
			got, err := srv.SayHello(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("greeterService.SayHello() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("greeterService.SayHello() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
