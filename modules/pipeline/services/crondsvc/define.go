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

package crondsvc

import (
	"sync"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/jsonstore"
)

type CrondSvc struct {
	crond    *cron.Cron
	cronChan chan string
	mu       *sync.Mutex
	dbClient *dbclient.Client
	bdl      *bundle.Bundle
	js       jsonstore.JsonStore
}

func New(dbClient *dbclient.Client, bdl *bundle.Bundle, js jsonstore.JsonStore) *CrondSvc {
	d := CrondSvc{}
	d.cronChan = make(chan string, 10)
	d.crond = cron.New()
	d.mu = &sync.Mutex{}
	d.dbClient = dbClient
	d.bdl = bdl
	d.js = js
	return &d
}
