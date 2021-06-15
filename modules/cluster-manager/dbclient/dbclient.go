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

package dbclient

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

type DBClient struct {
	*dbengine.DBEngine
}

func Open() *DBClient {
	return &DBClient{DBEngine: dbengine.MustOpen()}
}
