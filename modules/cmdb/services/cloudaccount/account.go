// Package account 账号逻辑封装
package cloudaccount

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
)

// CloudAccount 账号逻辑封装
type CloudAccount struct {
	db *dao.DBClient
}

// Option 定义 CloudAccount 对象的配置选项
type Option func(*CloudAccount)

// New 新建 CloudAccount 实例，操作账号资源
func New(options ...Option) *CloudAccount {
	account := &CloudAccount{}
	for _, op := range options {
		op(account)
	}
	return account
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(a *CloudAccount) {
		a.db = db
	}
}

// Create 创建账号
func (a *CloudAccount) Create(orgID int64, createReq *apistructs.CloudAccountCreateRequest) (*apistructs.CloudAccountInfo, error) {
	// 检查 name 重复
	account, err := a.db.GetAccountByName(orgID, createReq.Name)
	if err != nil {
		logrus.Errorf("failed to get cloud account, (%v)", err)
		return nil, errors.Errorf("failed to get cloud account")
	}
	if account.ID > 0 {
		return nil, errors.Errorf("failed to create cloud account (name already exists)")
	}

	account = model.CloudAccount{
		CloudProvider:   createReq.CloudProvider,
		Name:            createReq.Name,
		AccessKeyID:     createReq.AccessKeyID,
		AccessKeySecret: createReq.AccessKeySecret,
		OrgID:           orgID,
	}
	if err = a.db.CreateAccount(&account); err != nil {
		logrus.Warnf("failed to insert account to db, (%v)", err)
		return nil, apierrors.ErrCreateCloudAccount.InternalError(err)
	}

	accountInfo := apistructs.CloudAccountInfo{
		ID:            account.ID,
		CloudProvider: account.CloudProvider,
		Name:          account.Name,
		OrgID:         account.OrgID,
	}

	return &accountInfo, nil
}

// Update 跟新账号
func (a *CloudAccount) Update(orgID, accountID int64, updateReq *apistructs.CloudAccountUpdateRequest) (*apistructs.CloudAccountInfo, error) {
	account, err := a.db.GetAccountByID(orgID, accountID)
	if err != nil {
		return nil, apierrors.ErrUpdateCloudAccount.InvalidState(fmt.Sprintf("failed to find update account, err %v", err))
	}

	// 检查 name 重复
	another, err := a.db.GetAccountByName(orgID, updateReq.Name)
	if err != nil {
		logrus.Warnf("failed to get account, (%v)", err)
		return nil, apierrors.ErrUpdateCloudAccount.InternalError(err)
	}
	if another.ID > 0 {
		return nil, apierrors.ErrUpdateCloudAccount.InvalidState("failed to update account (name already exists)")
	}

	// 更新 account 信息
	account.CloudProvider = updateReq.CloudProvider
	account.Name = updateReq.Name
	account.AccessKeyID = updateReq.AccessKeyID
	account.AccessKeySecret = updateReq.AccessKeySecret

	if err = a.db.UpdateAccount(&account); err != nil {
		logrus.Warnf("failed to update account, (%v)", err)
		return nil, apierrors.ErrUpdateCloudAccount.InternalError(err)
	}

	accountInfo := apistructs.CloudAccountInfo{
		ID:            account.ID,
		CloudProvider: account.CloudProvider,
		Name:          account.Name,
		OrgID:         account.OrgID,
	}
	return &accountInfo, nil
}

// Delete 删除云账号
func (a *CloudAccount) Delete(orgID, accountID int64) error {
	if err := a.db.DeleteAccount(orgID, accountID); err != nil {
		logrus.Warnf("failed to delete account, (%v)", err)
		return apierrors.ErrDeleteCloudAccount.InternalError(err)
	}

	logrus.Infof("delete account: %d", accountID)
	return nil
}

// List 列出云账号
func (a *CloudAccount) ListByOrgID(orgID int64) ([]apistructs.CloudAccountInfo, error) {
	accounts, err := a.db.GetAccountsByOrgID(orgID)
	if err != nil {
		logrus.Infof("failed to get accounts, (%v)", err)
		return nil, apierrors.ErrListCloudAccount.InternalError(err)
	}

	accountInfos := make([]apistructs.CloudAccountInfo, 0, len(accounts))
	for _, account := range accounts {
		accountInfo := apistructs.CloudAccountInfo{
			ID:            account.ID,
			CloudProvider: account.CloudProvider,
			Name:          account.Name,
			OrgID:         account.OrgID,
		}
		accountInfos = append(accountInfos, accountInfo)
	}

	return accountInfos, nil
}

// Get 获取云账号
func (a *CloudAccount) GetByID(orgID, accountID int64) (*apistructs.CloudAccountAllInfo, error) {
	account, err := a.db.GetAccountByID(orgID, accountID)
	if err != nil {
		logrus.Infof("failed to get account, (%v)", err)
		return nil, errors.Errorf("failed to get account")
	}

	accountAllInfo := apistructs.CloudAccountAllInfo{
		CloudAccountInfo: apistructs.CloudAccountInfo{
			ID:            account.ID,
			CloudProvider: account.CloudProvider,
			Name:          account.Name,
			OrgID:         account.OrgID,
		},
		AccessKeyID:     account.AccessKeyID,
		AccessKeySecret: account.AccessKeySecret,
	}

	return &accountAllInfo, nil
}
