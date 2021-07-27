package base

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/base/pipelinesvc"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type baseService struct {
	p *provider

	svc *pipelinesvc.PipelineSvc
}

func (s *baseService) PipelineCreate(ctx context.Context, req *pb.PipelineCreateRequest) (*pb.PipelineCreateResponse, error) {
	// compatible handler
	if req.AutoRun {
		req.AutoRunAtOnce = true
	}
	logrus.Debugf("pipeline create request: %+v", req)

	// authentication check
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrCreatePipeline.AccessDenied()
	}
	req.IdentityInfo = identityInfo

	p, err := s.svc.CreateV2(req)
	if err != nil {
		return nil, err
	}

	return &pb.PipelineCreateResponse{
		Data: s.svc.ConvertPipeline(p),
	}, nil
}
