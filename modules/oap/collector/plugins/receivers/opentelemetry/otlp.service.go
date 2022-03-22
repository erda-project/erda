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

package opentelemetry

import (
	"context"

	"github.com/erda-project/erda-infra/base/logs"
	common "github.com/erda-project/erda-proto-go/common/pb"
	pb "github.com/erda-project/erda-proto-go/oap/collector/receiver/opentelemetry/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
)

type otlpService struct {
	Log logs.Logger
	// writer writer.Writer
	p *provider
}

func (s *otlpService) Export(ctx context.Context, req *pb.PostSpansRequest) (*common.VoidResponse, error) {
	if req.Spans != nil && s.p.consumer != nil {
		for _, span := range req.Spans {
			s.p.consumer(odata.NewSpan(span))
			// 	if err := s.writer.Write(span); err != nil {
			// 		s.Log.Error("write opentelemetry traces to kafka failed.")
			// 	}
		}
	}
	return &common.VoidResponse{}, nil
}
