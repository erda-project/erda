package example

import (
	"context"

	"github.com/erda-project/erda-proto-go/examples/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	p *provider
}

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// TODO .
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (s *userService) UpdateUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UpdateUserResponse, error) {
	// TODO .
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
