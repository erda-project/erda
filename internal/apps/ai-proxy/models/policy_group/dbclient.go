package policy_group

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/sqlutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (db *DBClient) Create(ctx context.Context, req *pb.PolicyGroupCreateRequest) (*pb.PolicyGroup, error) {
	pg := &PolicyGroup{
		ClientID:  req.ClientId,
		Name:      req.Name,
		Desc:      req.Desc,
		Mode:      common_types.PolicyGroupMode(req.Mode),
		StickyKey: req.StickyKey,
		Branches:  req.Branches,
		Source:    req.Source,
	}
	if err := db.DB.WithContext(ctx).Model(pg).Create(pg).Error; err != nil {
		return nil, err
	}
	return pg.ToProtobuf(), nil
}

func (db *DBClient) Get(ctx context.Context, req *pb.PolicyGroupGetRequest) (*pb.PolicyGroup, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	pg := &PolicyGroup{BaseModel: common.BaseModelWithID(req.Id)}
	where := &PolicyGroup{ClientID: req.ClientId}
	if err := db.DB.WithContext(ctx).Model(pg).Where(where).First(pg).Error; err != nil {
		return nil, err
	}
	return pg.ToProtobuf(), nil
}

func (db *DBClient) Update(ctx context.Context, req *pb.PolicyGroupUpdateRequest) (*pb.PolicyGroup, error) {
	u := &PolicyGroup{
		BaseModel: common.BaseModelWithID(req.Id),
		ClientID:  req.ClientId,
	}
	m := make(map[string]any)
	needUpdate := false
	if req.Name != nil {
		m["name"] = *req.Name
		needUpdate = true
	}
	if req.Desc != nil {
		m["desc"] = *req.Desc
		needUpdate = true
	}
	if req.Mode != nil {
		m["mode"] = common_types.PolicyGroupMode(*req.Mode)
		needUpdate = true
	}
	if req.StickyKey != nil {
		m["sticky_key"] = *req.StickyKey
		needUpdate = true
	}
	if len(req.Branches) > 0 {
		b, err := json.Marshal(req.Branches)
		if err != nil {
			return nil, err
		}
		m["branches"] = string(b)
		needUpdate = true
	}
	if req.Source != nil {
		m["source"] = *req.Source
		needUpdate = true
	}
	if !needUpdate {
		return nil, fmt.Errorf("nothing need update")
	}
	if err := db.DB.WithContext(ctx).Model(u).Where(u).Updates(m).Error; err != nil {
		return nil, err
	}
	return db.Get(ctx, &pb.PolicyGroupGetRequest{Id: req.Id, ClientId: req.ClientId})
}

func (db *DBClient) Delete(ctx context.Context, req *pb.PolicyGroupDeleteRequest) (*commonpb.VoidResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if req.Id == "" {
		return nil, gorm.ErrPrimaryKeyRequired
	}
	pg := &PolicyGroup{BaseModel: common.BaseModelWithID(req.Id)}
	where := &PolicyGroup{ClientID: req.ClientId}
	sql := db.DB.WithContext(ctx).Model(pg).Where(where).Delete(pg)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (db *DBClient) Paging(ctx context.Context, req *pb.PolicyGroupPagingRequest) (*pb.PolicyGroupPagingResponse, error) {
	c := &PolicyGroup{
		ClientID: req.ClientId,
		Mode:     common_types.PolicyGroupMode(req.Mode),
		Source:   req.Source,
	}
	sql := db.DB.WithContext(ctx).Model(c)

	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	// name / nameFull
	switch {
	case req.Name != "" && req.NameFull != "":
		sql = sql.Where("(name LIKE ? OR name = ?)",
			"%"+req.Name+"%", req.NameFull)
	case req.NameFull != "":
		sql = sql.Where("name = ?", req.NameFull)
	case req.Name != "":
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
	}

	sql, err := sqlutil.HandleOrderBy(sql, req.OrderBys)
	if err != nil {
		return nil, err
	}

	offset := (req.PageNum - 1) * req.PageSize
	var (
		total int64
		list  []PolicyGroup
	)
	if err := sql.Where(c).Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error; err != nil {
		return nil, err
	}
	resp := &pb.PolicyGroupPagingResponse{
		Total: total,
	}
	for i := range list {
		resp.List = append(resp.List, list[i].ToProtobuf())
	}
	return resp, nil
}

func checkBranches(branches []*pb.PolicyBranch) error {
	for _, branch := range branches {
		if err := checkBranch(branch); err != nil {
			return err
		}
	}
	return nil
}

func checkBranch(branch *pb.PolicyBranch) error {
	return nil
}
