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
