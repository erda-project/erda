package mbox

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/utils"
)

type MBox struct {
	db  *dao.DBClient
	uc  *utils.UCClient
	bdl *bundle.Bundle
}

type Option func(*MBox)

func New(options ...Option) *MBox {
	o := &MBox{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *MBox) {
		o.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(o *MBox) {
		o.bdl = bdl
	}
}

func (o *MBox) CreateMBox(createReq *apistructs.CreateMBoxRequest) error {
	return o.db.CreateMBox(createReq)
}

func (o *MBox) QueryMBox(queryReq *apistructs.QueryMBoxRequest) (*apistructs.QueryMBoxData, error) {
	return o.db.QueryMBox(queryReq)
}

func (o *MBox) GetMBox(id int64, orgID int64, userID string) (*apistructs.MBox, error) {
	mbox, err := o.db.GetMBox(id, orgID, userID)
	if err != nil {
		return nil, err
	}
	if mbox.Status == apistructs.MBoxUnReadStatus {
		err := o.db.SetMBoxReadStatus(&apistructs.SetMBoxReadStatusRequest{
			OrgID:  orgID,
			IDs:    []int64{id},
			UserID: userID,
		})
		if err != nil {
			return nil, err
		}
	}
	return mbox, nil
}

func (o *MBox) SetMBoxReadStatus(request *apistructs.SetMBoxReadStatusRequest) error {
	return o.db.SetMBoxReadStatus(request)
}

func (o *MBox) GetMBoxStats(orgID int64, userID string) (*apistructs.QueryMBoxStatsData, error) {
	return o.db.GetMBoxStats(orgID, userID)
}
