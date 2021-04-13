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

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/modules/gittar/conf"
	"github.com/erda-project/erda/modules/gittar/webcontext"
)

func GetGoImportMeta(ctx *webcontext.Context) {
	// go import 元数据 相关参考
	// https://golang.org/cmd/go/#hdr-Remote_import_paths
	// https://www.jianshu.com/p/90ef66e41f3c
	// https://blog.zhaoweiguo.com/2019/09/24/golang-env-private-git/
	repoUrl := conf.UIPublicURL() + "/" + ctx.HttpRequest().URL.Path
	module := strings.Replace(repoUrl, "http://", "", 1)
	module = strings.Replace(repoUrl, "https://", "", 1)
	if ctx.Query("go-get") == "1" {
		ctx.Data(http.StatusOK,
			"text/html; charset=utf-8",
			[]byte(fmt.Sprintf(`<html>
		   <head>
		       <meta name="go-import" content="%s git %s" >
		   </head>
</html>`, module, repoUrl)))
	} else {
		ctx.Success("")
	}
}
