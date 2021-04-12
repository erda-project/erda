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

package mysqlclientstore

import (
	"time"

	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erda-project/erda/pkg/dbengine"
)

type ClientStore struct {
	db *gorm.DB
}

// ClientStoreItem data item
type ClientStoreItem struct {
	ID        string `gorm:"primary_key"`
	Secret    string
	Domain    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ClientStoreItem) TableName() string {
	return "openapi_oauth2_token_clients"
}

// NewClientStore creates PostgreSQL store instance
func NewClientStore() (*ClientStore, error) {
	db, err := dbengine.Open()
	if err != nil {
		return nil, err
	}
	store := &ClientStore{db: db.DB}
	return store, err
}

func (s *ClientStore) toClientInfo(item ClientStoreItem) oauth2.ClientInfo {
	return &models.Client{ID: item.ID, Secret: item.Secret}
}

// GetByID retrieves and returns client information by id
func (s *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	if err := s.db.Where("id = ?", id).Find(&item).Error; err != nil {
		return nil, err
	}

	return s.toClientInfo(item), nil
}

// Create creates and stores the new client information
func (s *ClientStore) Create(info oauth2.ClientInfo) error {
	return s.db.Create(&ClientStoreItem{
		ID:        info.GetID(),
		Secret:    info.GetSecret(),
		Domain:    info.GetDomain(),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}).Error
}
