package example

import (
	"context"

	"github.com/erda-project/erda-proto-go/examples/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type greeterService struct {
	p *provider
}

func (s *greeterService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	// TODO .
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
