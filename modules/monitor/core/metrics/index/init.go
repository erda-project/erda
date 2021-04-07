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

package indexmanager

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/olivere/elastic"
	cfgpkg "github.com/recallsong/go-utils/config"
)

func (p *provider) initTemplate(client *elastic.Client) error {
	if len(p.C.IndexTemplateName) <= 0 || len(p.C.IndexTemplatePath) <= 0 {
		return fmt.Errorf("index template name or file is empty")
	}
	template, err := ioutil.ReadFile(p.C.IndexTemplatePath)
	if err != nil {
		return fmt.Errorf("fail to load index template: %s", err)
	}
	template = cfgpkg.EscapeEnv(template)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		resp, err := client.IndexPutTemplate(p.C.IndexTemplateName).
			BodyString(string(template)).Do(ctx)
		if err != nil {
			return fmt.Errorf("fail to set index template: %s", err)
		}
		if resp.Acknowledged {
			break
		}
	}
	p.L.Infof("Put index template (%s) success", p.C.IndexTemplateName)

	emptyIndex := p.C.IndexPrefix + "-empty"
	exists, err := client.IndexExists(emptyIndex).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		for i := 0; i < 2; i++ {
			resp, e := client.CreateIndex(emptyIndex).Do(ctx)
			if err == nil && resp.Acknowledged {
				break
			}
			err = e
		}
		if err != nil {
			return err
		}
	}
	p.L.Infof("Put empty index (%s) success", emptyIndex)

	for i := 0; i < 2; i++ {
		_, err = client.Index().Index(emptyIndex).Type(p.C.IndexType).BodyJson(&metrics.Metric{
			Name:      "monitor",
			Timestamp: time.Now().UnixNano(),
			Tags:      map[string]string{"from": "monitor"},
			Fields:    map[string]interface{}{"value": 1},
		}).Do(ctx)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("fail to init empty index: %s", err)
	}

	if p.C.EnableRollover {
		for i := 0; i < 2; i++ {
			// remove this indices, avoid collisions with aliases
			_, err := client.DeleteIndex(p.C.IndexPrefix + "-*-rollover").Do(ctx)
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("fail to delete index %s: %s", p.C.IndexPrefix+"-*-rollover", err)
		}
	}
	return nil
}
