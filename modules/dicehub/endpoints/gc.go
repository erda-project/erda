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

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/httpserver"
)

// ReleaseGC 通过GET /gc API触发releaese gc时处理逻辑
func (e *Endpoints) ReleaseGC(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	logrus.Infof("trigger release gc by api[ %s %s]!", r.Method, r.RequestURI)
	go func() {
		if err := e.release.RemoveDeprecatedsReleases(time.Now()); err != nil {
			logrus.Warnf("remove deprecated release error: %v", err)
		}
	}()

	return httpserver.OkResp("trigger release gc success")
}
