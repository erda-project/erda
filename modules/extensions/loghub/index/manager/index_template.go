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

package manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/olivere/elastic"
	cfgpkg "github.com/recallsong/go-utils/config"
)

func (p *provider) setupIndexTemplate(client *elastic.Client) error {
	if len(p.C.IndexTemplateName) <= 0 || len(p.C.IndexTemplateFile) <= 0 {
		return nil
	}
	template, err := ioutil.ReadFile(p.C.IndexTemplateFile)
	if err != nil {
		return fmt.Errorf("fail to load index template: %s", err)
	}
	template = cfgpkg.EscapeEnv(template)
	body := string(template)
	p.L.Info("load index template: \n", body)
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		resp, err := client.IndexPutTemplate(p.C.IndexTemplateName).
			BodyString(body).Do(ctx)
		if err != nil {
			cancel()
			return fmt.Errorf("fail to set index template: %s", err)
		}
		cancel()
		if resp.Acknowledged {
			break
		}
	}
	p.L.Infof("Put index template (%s) success", p.C.IndexTemplateName)
	return nil
}
