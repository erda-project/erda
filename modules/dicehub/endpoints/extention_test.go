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

package endpoints

// import (
// 	"encoding/json"
// 	"fmt"
// 	"testing"

// 	"github.com/erda-project/erda/modules/dicehub/conf"
// 	"github.com/erda-project/erda/modules/dicehub/dbclient"
// 	"github.com/erda-project/erda/modules/dicehub/service/extension"
// 	"github.com/erda-project/erda/pkg/i18n"
// )

// func Test_menuExtWithFile(t *testing.T) {
// 	conf.Load()
// 	db, _ := dbclient.Open()
// 	ext := extension.New(
// 		extension.WithDBClient(db),
// 	)
// 	data, _ := ext.QueryExtensions("", "", "")

// 	i18nLoader := i18n.NewLoader()

// 	i18nLoader.LoadFile("../../../i18n/dicehub.json")
// 	i18nLoader.DefaultLocale("zh-CN")
// 	local := i18nLoader.Locale("zh-CN")

// 	result := menuExtWithLocale(data, local)
// 	byteData, _ := json.Marshal(result)
// 	fmt.Println(string(byteData))
// }
