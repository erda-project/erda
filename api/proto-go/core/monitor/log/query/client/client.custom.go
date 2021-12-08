// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"

var _ pb.LogQueryServiceServer = (*logQueryServiceWrapper)(nil)

// ScanLogsByExpression do not call me
// Notice: if you encounter func conflicting, delete that auto-generated one
func (s *logQueryServiceWrapper) ScanLogsByExpression(req *pb.GetLogByExpressionRequest, stream pb.LogQueryService_ScanLogsByExpressionServer) error {
	panic("do not call me")
}
