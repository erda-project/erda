package accesskey

import (
	context "context"
	"fmt"

	pb "github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type accessKeyService struct {
	p *provider
}

func (s *accessKeyService) QueryAccessKeys(ctx context.Context, req *pb.QueryAccessKeysRequest) (*pb.QueryAccessKeysResponse, error) {
	objs, err := s.p.dao.QueryAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.AccessKeysItem, len(objs))
	for i, obj := range objs {
		res[i] = &pb.AccessKeysItem{
			Id:          obj.ID,
			AccessKey:   obj.AccessKey,
			SecretKey:   obj.SecretKey,
			Status:      obj.Status,
			SubjectType: obj.SubjectType,
			Subject:     obj.Subject,
			Description: obj.Description,
			CreatedAt:   timestamppb.New(obj.CreatedAt),
		}
	}
	return &pb.QueryAccessKeysResponse{Data: res}, nil
}

func (s *accessKeyService) GetAccessKey(ctx context.Context, req *pb.GetAccessKeysRequest) (*pb.GetAccessKeysResponse, error) {
	obj := AccessKey{}
	db := s.p.DB.Where(&AccessKey{ID: req.Id}).Find(&obj)
	if db.RecordNotFound() {
		return nil, errors.NewNotFoundError("access-key")
	}
	if db.Error != nil {
		return nil, db.Error
	}

	return &pb.GetAccessKeysResponse{Data: &pb.AccessKeysItem{
		Id:          obj.ID,
		AccessKey:   obj.AccessKey,
		SecretKey:   obj.SecretKey,
		Status:      obj.Status,
		SubjectType: obj.SubjectType,
		Subject:     obj.Subject,
		Description: obj.Description,
		CreatedAt:   timestamppb.New(obj.CreatedAt),
	}}, nil
}

func (s *accessKeyService) CreateAccessKeys(ctx context.Context, req *pb.CreateAccessKeysRequest) (*pb.CreateAccessKeysResponse, error) {
	if err := validateSubjectType(req.SubjectType); err != nil {
		return nil, err
	}
	obj, err := s.p.dao.CreateAccessKey(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAccessKeysResponse{Data: &pb.AccessKeysItem{
		Id:          obj.ID,
		AccessKey:   obj.AccessKey,
		SecretKey:   obj.SecretKey,
		Status:      obj.Status,
		SubjectType: obj.SubjectType,
		Subject:     obj.Subject,
		Description: obj.Description,
		CreatedAt:   timestamppb.New(obj.CreatedAt),
	}}, nil
}

func (s *accessKeyService) UpdateAccessKeys(ctx context.Context, req *pb.UpdateAccessKeysRequest) (*pb.UpdateAccessKeysResponse, error) {
	if err := validateStatus(req.Status); err != nil {
		return nil, err
	}

	q := s.p.DB.Model(&AccessKey{}).Where(&AccessKey{ID: req.Id})
	updated := AccessKey{}
	if req.Status.String() != "" {
		updated.Status = req.Status
	}
	if req.Description != "" {
		updated.Description = req.Description
	}
	q = q.Update(updated)

	if q.Error != nil {
		return nil, q.Error
	}
	return nil, nil
}

func (s *accessKeyService) DeleteAccessKeys(ctx context.Context, req *pb.DeleteAccessKeysRequest) (*pb.DeleteAccessKeysResponse, error) {
	q := s.p.DB.Model(&AccessKey{}).Where(&AccessKey{ID: req.Id}).Update(&AccessKey{Status: pb.StatusEnum_DELETED})
	if q.Error != nil {
		return nil, q.Error
	}
	return nil, nil
}

func validateStatus(field pb.StatusEnum_Status) error {
	if _, ok := pb.StatusEnum_Status_value[field.String()]; !ok {
		return fmt.Errorf("invalid status")
	} else {
		return nil
	}
}

func validateSubjectType(field pb.SubjectTypeEnum_SubjectType) error {
	if _, ok := pb.SubjectTypeEnum_SubjectType_value[field.String()]; !ok {
		return fmt.Errorf("invalid subjectType")
	} else {
		return nil
	}
}
