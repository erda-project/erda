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

package kratos

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/internal/core/user/common"
)

func Test_getUserByKey(t *testing.T) {
	type args struct {
		kratosPrivateAddr string
		key               string
	}
	tests := []struct {
		name    string
		args    args
		want    []common.User
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				key: "a",
			},
			want: []common.User{
				{
					ID:    "1",
					Name:  "a1",
					State: UserActive,
				},
				{
					ID:    "2",
					Nick:  "a2",
					State: UserActive,
				},
			},
			wantErr: false,
		},
	}

	monkey.Patch(getUserPage,
		func(kratosPrivateAddr string, page, perPage int) ([]common.User, error) {
			if page == 2 {
				return nil, nil
			}
			return []common.User{
				{
					ID:    "1",
					Name:  "a1",
					State: UserActive,
				},
				{
					ID:    "2",
					Nick:  "a2",
					State: UserActive,
				},
				{
					ID:    "3",
					Email: "abc@gmail.com",
					State: UserInActive,
				},
			}, nil
		})
	defer monkey.UnpatchAll()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserByKey(tt.args.kratosPrivateAddr, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserByKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserByKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserPage(t *testing.T) {
	type args struct {
		kratosPrivateAddr string
		page              int
		perPage           int
	}
	tests := []struct {
		name    string
		args    args
		want    []common.User
		wantErr bool
	}{
		{
			args: args{
				page:    1,
				perPage: 1,
			},
			want: []common.User{
				{
					ID:    "1",
					State: UserActive,
				},
			},
		},
	}

	monkey.Patch(getIdentityPage,
		func(kratosPrivateAddr string, page, perPage int) ([]*OryKratosIdentity, error) {
			return []*OryKratosIdentity{
				{
					ID:    "1",
					State: UserActive,
				},
			}, nil
		})
	defer monkey.UnpatchAll()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserPage(tt.args.kratosPrivateAddr, tt.args.page, tt.args.perPage)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserByIDs(t *testing.T) {
	type args struct {
		kratosPrivateAddr string
		userIDs           []string
	}
	tests := []struct {
		name    string
		args    args
		want    []common.User
		wantErr bool
	}{
		{
			args: args{
				userIDs: []string{"1"},
			},
			want: []common.User{
				{
					ID:    "1",
					State: UserActive,
				},
			},
		},
	}

	monkey.Patch(getUserByID,
		func(kratosPrivateAddr string, userID string) (*common.User, error) {
			return &common.User{
					ID:    "1",
					State: UserActive,
				},
				nil
		})
	defer monkey.UnpatchAll()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserByIDs(tt.args.kratosPrivateAddr, tt.args.userIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserByIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserByIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserByID(t *testing.T) {
	type args struct {
		kratosPrivateAddr string
		userID            string
	}
	tests := []struct {
		name    string
		args    args
		want    *common.User
		wantErr bool
	}{
		{
			args: args{
				userID: "1",
			},
			want: &common.User{
				ID:    "1",
				State: UserActive,
			},
		},
	}

	monkey.Patch(getIdentity,
		func(kratosPrivateAddr string, userID string) (*OryKratosIdentity, error) {
			return &OryKratosIdentity{
					ID:    "1",
					State: UserActive,
				},
				nil
		})
	defer monkey.UnpatchAll()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserByID(tt.args.kratosPrivateAddr, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
