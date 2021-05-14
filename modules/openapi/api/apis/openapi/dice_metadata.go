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

package openapi

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/conf"
)

var DICE_METADATA = apis.ApiSpec{
	Path:   "/metadata.json",
	Scheme: "http",
	Method: "GET",
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		meta := make(map[string]interface{})
		meta["version"] = map[string]interface{}{
			"dice_version": version.Version,
			"git_commit":   version.CommitID,
			"go_version":   version.GoVersion,
			"built":        version.BuildTime,
		}
		meta["openapi_public_url"] = conf.SelfPublicURL()

		metaBytes, _ := json.MarshalIndent(meta, "", "  ")
		rw.Write(metaBytes)
	},
	Doc: "Dice 平台对外的元信息",
}
