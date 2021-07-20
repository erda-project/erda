package expression

import (
	context "context"
	pb "github.com/erda-project/erda-proto-go/msp/apm/expression/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type expressionService struct {
	p *provider
}

func (s *expressionService) GetExpression(ctx context.Context, req *pb.GetExpressionRequest) (*pb.GetExpressionResponse, error) {
	if req.Type == "" {
		return nil, errors.NewMissingParameterError(req.Type)
	}
	var expressions []*pb.Expression
	readExpression(GetFS(req.Type), req.Type, &expressions)
	return &pb.GetExpressionResponse{Data: expressions}, nil
}
