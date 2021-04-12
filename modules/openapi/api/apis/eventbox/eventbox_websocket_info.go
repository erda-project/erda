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

package eventbox

import "github.com/erda-project/erda/modules/openapi/api/apis"

var EVENTBOX_WEBSOCKET_INFO = apis.ApiSpec{
	Path:        "/api/websocket/info",
	BackendPath: "/api/dice/eventbox/ws/info",
	Host:        "eventbox.marathon.l4lb.thisdcos.directory:9528",
	Scheme:      "ws",
	Method:      "GET",
	CheckLogin:  true,
	Doc: `
summary: dice's websocket proxy info
`,
}
