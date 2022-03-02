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

package extmarketsvc

import (
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/goroutinepool"
)

const (
	PoolSize = 20
)

type ExtMarketSvc struct {
	sync.Mutex
	bdl     *bundle.Bundle
	actions map[string]apistructs.ExtensionVersion
	pools   *goroutinepool.GoroutinePool
}

func New(bdl *bundle.Bundle) *ExtMarketSvc {
	s := ExtMarketSvc{}
	s.bdl = bdl
	s.actions = make(map[string]apistructs.ExtensionVersion)
	s.pools = goroutinepool.New(PoolSize)
	if err := s.constructAllActions(); err != nil {
		panic(err)
	}
	go s.continuousRefreshActionAsync()
	return &s
}
