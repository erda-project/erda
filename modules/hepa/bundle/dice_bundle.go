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

package bundle

import (
	"time"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
)

var Bundle *bundle.Bundle

func init() {
	bundleOpts := []bundle.Option{
		bundle.WithOrchestrator(),
		bundle.WithOps(),
		bundle.WithScheduler(),
		bundle.WithCMDB(),
		bundle.WithDiceHub(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*30),
		)),
	}
	Bundle = bundle.New(bundleOpts...)
}
