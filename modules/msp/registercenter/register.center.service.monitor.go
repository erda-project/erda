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

package registercenter

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

// GetDubboInterfaceTime depracated
func (s *registerCenterService) GetDubboInterfaceTime(ctx context.Context, req *pb.GetDubboInterfaceTimeRequest) (*pb.GetDubboInterfaceTimeResponse, error) {
	return nil, errors.NewUnimplementedError("GetDubboInterfaceTime")
}

// GetDubboInterfaceQPS depracated
func (s *registerCenterService) GetDubboInterfaceQPS(ctx context.Context, req *pb.GetDubboInterfaceQPSRequest) (*pb.GetDubboInterfaceQPSResponse, error) {
	return nil, errors.NewUnimplementedError("GetDubboInterfaceQPS")
}

// GetDubboInterfaceFailed depracated
func (s *registerCenterService) GetDubboInterfaceFailed(ctx context.Context, req *pb.GetDubboInterfaceFailedRequest) (*pb.GetDubboInterfaceFailedResponse, error) {
	return nil, errors.NewUnimplementedError("GetDubboInterfaceFailed")
}

// GetDubboInterfaceAvgTime depracated
func (s *registerCenterService) GetDubboInterfaceAvgTime(ctx context.Context, req *pb.GetDubboInterfaceAvgTimeRequest) (*pb.GetDubboInterfaceAvgTimeResponse, error) {
	return nil, errors.NewUnimplementedError("GetDubboInterfaceAvgTime")
}
