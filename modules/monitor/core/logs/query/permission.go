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

package query

import (
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) checkApplicationID(ctx httpserver.Context) (string, error) {
	appID, err := p.getApplicationID(ctx)
	if err != nil {
		return "", err
	}
	return appID, nil
}

func (p *provider) checkContainerLog(ctx httpserver.Context) (string, error) {
	orgID, err := p.checkOrgCluster(ctx)
	if err != nil {
		return "", err
	}
	return orgID, nil
}

func (p *provider) provider() {

}
