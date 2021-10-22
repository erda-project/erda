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

package cleaner

import (
	"net/http"

	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRoutes(routes httpserver.Router, prefix string) error {
	routes.GET(prefix+"/cleanner/rollover-body", p.getRolloverBody)
	return nil
}

func (p *provider) getRolloverBody(r *http.Request) interface{} {
	return p.rolloverBodyForDiskClean
}
