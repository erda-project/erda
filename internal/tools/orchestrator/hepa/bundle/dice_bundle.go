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

package bundle

import (
	"time"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var Bundle *bundle.Bundle

func init() {
	bundleOpts := []bundle.Option{
		bundle.WithOrchestrator(),
		bundle.WithCMP(),
		bundle.WithMSP(),
		bundle.WithScheduler(),
		bundle.WithErdaServer(),
		bundle.WithClusterManager(),
		bundle.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second*10, time.Second*60),
		)),
	}
	Bundle = bundle.New(bundleOpts...)
}
