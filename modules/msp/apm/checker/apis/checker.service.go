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

package apis

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type checkerService struct {
	p *provider
}

func (s *checkerService) CreateChecker(ctx context.Context, req *pb.CreateCheckerRequest) (*pb.CreateCheckerResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("CreateChecker")
}
func (s *checkerService) UpdateChecker(ctx context.Context, req *pb.UpdateCheckerRequest) (*pb.UpdateCheckerResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("UpdateChecker")
}
func (s *checkerService) DeleteChecker(ctx context.Context, req *pb.UpdateCheckerRequest) (*pb.UpdateCheckerResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("DeleteChecker")
}
func (s *checkerService) ListCheckers(ctx context.Context, req *pb.ListCheckersRequest) (*pb.ListCheckersResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("ListCheckers")
}
func (s *checkerService) DescribeCheckers(ctx context.Context, req *pb.DescribeCheckersRequest) (*pb.DescribeCheckersResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("DescribeCheckers")
}
func (s *checkerService) DescribeChecker(ctx context.Context, req *pb.DescribeCheckerRequest) (*pb.DescribeCheckerResponse, error) {
	// TODO .
	return nil, errors.NewUnimplementedError("DescribeChecker")
}
