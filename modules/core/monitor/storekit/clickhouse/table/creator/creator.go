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

package creator

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	cfgpkg "github.com/recallsong/go-utils/config"

	"github.com/erda-project/erda/modules/core/monitor/storekit/clickhouse/table"
)

const (
	tableNameKey      = "<table_name>"
	aliasTableNameKey = "<alias_table_name>"
)

func (p *provider) Ensure(ctx context.Context, tenant, key string) (_ <-chan error, aliasName string) {
	tableName := table.NormalizeKey(fmt.Sprintf("%s_%s", tenant, key))
	aliasName = table.NormalizeKey(fmt.Sprintf("%s", tenant))

	if _, ok := p.created.Load(tableName); ok {
		return
	}
	if p.Loader.ExistsTable(tenant, key) {
		return
	}

	ch := make(chan error, 1)
	p.createCh <- request{
		TableName: tableName,
		AliasName: aliasName,
		Wait:      ch,
		Ctx:       ctx,
	}
	return ch, aliasName
}

func (p *provider) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case req := <-p.createCh:
			var err error
			if req.Ctx == nil {
				req.Ctx = context.Background()
			}
			func() {
				if _, ok := p.created.Load(req.TableName); ok {
					return
				}
				p.createLock.Lock()
				defer p.createLock.Unlock()
				if _, ok := p.created.Load(req.TableName); ok {
					return
				}
				for {
					err = p.createTable(ctx, req.TableName, req.AliasName)
					if err == nil {
						p.created.Store(req.TableName, true)
						return
					}
					p.Log.Error(err)
					select {
					case <-ctx.Done():
						return
					case <-req.Ctx.Done():
						return
					default:
						time.Sleep(2 * time.Second)
					}
				}
			}()
			if req.Wait != nil {
				if err != nil {
					req.Wait <- err
				} else {
					req.Wait <- nil
				}
				close(req.Wait)
			}
		}
	}
	return nil
}

func (p *provider) createTable(ctx context.Context, tableName, aliasTableName string) error {
	replacer := strings.NewReplacer(tableNameKey, tableName,
		aliasTableNameKey, aliasTableName)

	data, err := ioutil.ReadFile(p.Cfg.DDLTemplate)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", p.Cfg.DDLTemplate)
	}
	data = cfgpkg.EscapeEnv(data)
	regex, _ := regexp.Compile("[^;]+[;$]")
	ddls := regex.FindAllString(string(data), -1)
	for _, ddl := range ddls {
		ddl = replacer.Replace(ddl)
		err = p.Clickhouse.Client().Exec(ctx, ddl)
		if err != nil {
			return fmt.Errorf("failed to exec ddl[%s], err: %s", ddl, err)
		}
	}
	return nil
}
