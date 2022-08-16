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

package dbclient

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

func InitMysqlMock() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if nil != err {
		return nil, nil, err
	}
	mdb, err := gorm.Open("mysql", db)
	if err != nil {
		return nil, nil, err
	}
	return mdb, mock, err
}
