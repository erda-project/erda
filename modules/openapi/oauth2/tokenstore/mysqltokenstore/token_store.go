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

package mysqltokenstore

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

type TokenStore struct {
	db         *gorm.DB
	gcDisabled bool
	gcInterval time.Duration
	ticker     *time.Ticker
}

// TokenStoreItem data item
type TokenStoreItem struct {
	ID        int64 `gorm:"primary_key"`
	CreatedAt time.Time
	ExpiredAt *time.Time
	Code      string
	Access    string
	Refresh   string
	Data      TokenStoreItemData
}

type TokenStoreItemData struct {
	TokenInfo oauth2.TokenInfo `json:"tokenInfo"`
}

func (data TokenStoreItemData) Value() (driver.Value, error) {
	if b, err := json.Marshal(data); err != nil {
		return nil, errors.Errorf("failed to json encode token data, err: %v", err)
	} else {
		return string(b), nil
	}
}

func (data *TokenStoreItemData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for token data")
	}
	if len(v) == 0 {
		return nil
	}
	var tm struct {
		TokenInfo *models.Token `json:"tokenInfo"`
	}
	if err := json.Unmarshal(v, &tm); err != nil {
		return errors.Errorf("failed to json decode token data, err: %v", err)
	}
	data.TokenInfo = tm.TokenInfo
	return nil
}

func (TokenStoreItem) TableName() string {
	return "openapi_oauth2_tokens"
}

// NewTokenStore creates PostgreSQL store instance
func NewTokenStore(options ...TokenStoreOption) (*TokenStore, error) {
	db, err := dbengine.Open()
	if err != nil {
		return nil, err
	}

	store := &TokenStore{
		db:         db.DB,
		gcInterval: 10 * time.Minute,
	}

	for _, o := range options {
		o(store)
	}

	if !store.gcDisabled {
		store.ticker = time.NewTicker(store.gcInterval)
		go store.gc()
	}

	return store, err
}

// Close close the store
func (s *TokenStore) Close() error {
	if !s.gcDisabled {
		s.ticker.Stop()
	}
	return nil
}

func (s *TokenStore) gc() {
	for range s.ticker.C {
		s.clean()
	}
}

func (s *TokenStore) clean() {
	now := time.Unix(time.Now().Unix(), 0)

	err := s.db.Where("expired_at is not null and expired_at <= ?", now).Or("code='' AND access='' AND refresh=''").Delete(&TokenStoreItem{}).Error
	if err != nil {
		logrus.Errorf("[alert] failed to gc clean expired openapi oauth2 token, err: %v", err)
		return
	}
}

// Create create and store the new token information
func (s *TokenStore) Create(info oauth2.TokenInfo) error {
	item := &TokenStoreItem{
		Data:      TokenStoreItemData{TokenInfo: info},
		CreatedAt: time.Unix(time.Now().Unix(), 0),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiredAt = handleExpiredAt(info.GetCodeCreateAt(), info.GetCodeExpiresIn())
	} else {
		item.Access = info.GetAccess()
		item.ExpiredAt = handleExpiredAt(info.GetAccessCreateAt(), info.GetAccessExpiresIn())

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiredAt = handleExpiredAt(info.GetRefreshCreateAt(), info.GetRefreshExpiresIn())
		}
	}

	return s.db.Create(&TokenStoreItem{
		CreatedAt: item.CreatedAt,
		ExpiredAt: item.ExpiredAt,
		Code:      item.Code,
		Access:    item.Access,
		Refresh:   item.Refresh,
		Data:      item.Data,
	}).Error
}

func handleExpiredAt(createdAt time.Time, expiredIn time.Duration) *time.Time {
	if expiredIn == 0 {
		return nil
	}
	return &[]time.Time{createdAt.Add(expiredIn)}[0]
}

// RemoveByCode delete the authorization code
func (s *TokenStore) RemoveByCode(code string) error {
	err := s.db.Where("code = ?", code).Delete(&TokenStoreItem{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// RemoveByAccess use the access token to delete the token information
func (s *TokenStore) RemoveByAccess(access string) error {
	err := s.db.Where("access = ?", access).Delete(&TokenStoreItem{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *TokenStore) RemoveByRefresh(refresh string) error {
	err := s.db.Where("refresh = ?", refresh).Delete(&TokenStoreItem{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// GetByCode use the authorization code for token information data
func (s *TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var tokenItem TokenStoreItem
	err := s.db.Where("code = ?", code).First(&tokenItem).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return tokenItem.Data.TokenInfo, nil
}

// GetByAccess use the access token for token information data
func (s *TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	var tokenItem TokenStoreItem
	err := s.db.Where("access = ?", access).First(&tokenItem).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return tokenItem.Data.TokenInfo, nil
}

// GetByRefresh use the refresh token for token information data
func (s *TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var tokenItem TokenStoreItem
	err := s.db.Where("refresh = ?", refresh).First(&tokenItem).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return tokenItem.Data.TokenInfo, nil
}
