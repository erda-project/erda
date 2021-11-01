package entity

import (
	"context"

	"github.com/erda-project/erda-proto-go/oap/entity/pb"
	"github.com/erda-project/erda/modules/core/monitor/entity/storage"
	"github.com/erda-project/erda/pkg/common/errors"
)

type entityService struct {
	p       *provider
	storage storage.Storage
}

func (s *entityService) SetEntity(ctx context.Context, req *pb.SetEntityRequest) (*pb.SetEntityResponse, error) {
	if req.Data == nil {
		return nil, errors.NewMissingParameterError("body")
	}
	err := s.storage.SetEntity(ctx, req.Data)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.SetEntityResponse{
		Data: req.Data.Id,
	}, nil
}

func (s *entityService) RemoveEntity(ctx context.Context, req *pb.RemoveEntityRequest) (*pb.RemoveEntityResponse, error) {
	ok, err := s.storage.RemoveEntity(ctx, req.Type, req.Key)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.RemoveEntityResponse{
		Ok: ok,
	}, nil
}

func (s *entityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.GetEntityResponse, error) {
	entity, err := s.storage.GetEntity(ctx, req.Type, req.Key)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.GetEntityResponse{
		Data: entity,
	}, nil
}

func (s *entityService) ListEntities(ctx context.Context, req *pb.ListEntitiesRequest) (*pb.ListEntitiesResponse, error) {
	list, total, err := s.storage.ListEntities(ctx, &storage.ListOptions{
		Type:   req.Type,
		Labels: req.Labels,
		Limit:  int(req.Limit),
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.ListEntitiesResponse{
		Data: &pb.EntityList{
			List:  list,
			Total: total,
		},
	}, nil
}
