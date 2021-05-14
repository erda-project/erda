// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package example

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/erda-project/erda-proto-go/examples/pb"
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
