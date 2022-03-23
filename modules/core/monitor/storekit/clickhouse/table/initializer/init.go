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

package initializer

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"

	cfgpkg "github.com/recallsong/go-utils/config"
)

func (p *provider) initDDLs() error {
	for _, file := range p.Cfg.DDLs {
		data, err := ioutil.ReadFile(file.Path)
		if err != nil {
			return fmt.Errorf("failed to read file: %s", file.Path)
		}
		data = cfgpkg.EscapeEnv(data)
		regex, _ := regexp.Compile("[^;]+[;$]")
		ddls := regex.FindAllString(string(data), -1)

		for _, ddl := range ddls {
			err := p.Clickhouse.Client().Exec(context.Background(), ddl)
			if err == nil {
				continue
			}
			p.Log.Warnf("failed to execute ddl of file[%s], ddl: %s, err: %s", file.Path, ddl, err)
			if file.IgnoreErr {
				continue
			}
			return err
		}
	}
	return nil
}
