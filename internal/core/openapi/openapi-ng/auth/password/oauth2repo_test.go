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

package password

import (
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"gopkg.in/oauth2.v3"
	oauth2model "gopkg.in/oauth2.v3/models"
)

type mockClientStore struct {
	getByID          func(id string) (oauth2.ClientInfo, error)
	create           func(info oauth2.ClientInfo) error
	getByIDCalls     int
	createCalls      int
	lastGetByID      string
	lastCreateClient oauth2.ClientInfo
}

func (m *mockClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	m.getByIDCalls++
	m.lastGetByID = id
	if m.getByID == nil {
		return nil, nil
	}
	return m.getByID(id)
}

func (m *mockClientStore) Create(info oauth2.ClientInfo) error {
	m.createCalls++
	m.lastCreateClient = info
	if m.create == nil {
		return nil
	}
	return m.create(info)
}

func TestOAuth2Repo_GetOrCreateOpenAPIClient(t *testing.T) {
	tests := []struct {
		name              string
		getByID           func(id string) (oauth2.ClientInfo, error)
		create            func(info oauth2.ClientInfo) error
		wantErr           bool
		wantID            string
		wantSecret        string
		wantGetByIDCalls  int
		wantCreateCalls   int
		expectCreateCheck bool
	}{
		{
			name: "existing client",
			getByID: func(id string) (oauth2.ClientInfo, error) {
				return &oauth2model.Client{ID: id, Secret: "secret"}, nil
			},
			wantID:           openapiClientID,
			wantSecret:       "secret",
			wantGetByIDCalls: 1,
			wantCreateCalls:  0,
		},
		{
			name: "create when not found",
			getByID: func(id string) (oauth2.ClientInfo, error) {
				return nil, gorm.ErrRecordNotFound
			},
			create: func(info oauth2.ClientInfo) error {
				return nil
			},
			wantID:            openapiClientID,
			wantGetByIDCalls:  1,
			wantCreateCalls:   1,
			expectCreateCheck: true,
		},
		{
			name: "get by id error",
			getByID: func(id string) (oauth2.ClientInfo, error) {
				return nil, errors.New("boom")
			},
			wantErr:          true,
			wantGetByIDCalls: 1,
			wantCreateCalls:  0,
		},
		{
			name: "create error",
			getByID: func(id string) (oauth2.ClientInfo, error) {
				return nil, gorm.ErrRecordNotFound
			},
			create: func(info oauth2.ClientInfo) error {
				return errors.New("create failed")
			},
			wantErr:          true,
			wantGetByIDCalls: 1,
			wantCreateCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockClientStore{
				getByID: tt.getByID,
				create:  tt.create,
			}
			repo := &OAuth2Repo{clientStore: store}

			item, err := repo.GetOrCreateOpenAPIClient()
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, item)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantID, item.ID)
				if tt.wantSecret != "" {
					require.Equal(t, tt.wantSecret, item.Secret)
				} else {
					require.NotEmpty(t, item.Secret)
				}
			}

			require.Equal(t, tt.wantGetByIDCalls, store.getByIDCalls)
			require.Equal(t, openapiClientID, store.lastGetByID)
			require.Equal(t, tt.wantCreateCalls, store.createCalls)

			if tt.expectCreateCheck {
				require.NotNil(t, store.lastCreateClient)
				require.Equal(t, openapiClientID, store.lastCreateClient.GetID())
				require.Equal(t, item.Secret, store.lastCreateClient.GetSecret())
			}
		})
	}
}
