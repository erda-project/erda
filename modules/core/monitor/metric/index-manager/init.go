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

package indexmanager

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/olivere/elastic"
	cfgpkg "github.com/recallsong/go-utils/config"

	"github.com/erda-project/erda/modules/core/monitor/metric"
)

func (p *provider) initTemplate(client *elastic.Client) error {
	if len(p.Cfg.IndexTemplateName) <= 0 || len(p.Cfg.IndexTemplatePath) <= 0 {
		return fmt.Errorf("index template name or file is empty")
	}
	template, err := ioutil.ReadFile(p.Cfg.IndexTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to load index template: %s", err)
	}
	template = cfgpkg.EscapeEnv(template)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		resp, err := client.IndexPutTemplate(p.Cfg.IndexTemplateName).
			BodyString(string(template)).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to set index template: %s", err)
		}
		if resp.Acknowledged {
			break
		}
	}
	p.Log.Infof("Setup index template (%s) success", p.Cfg.IndexTemplateName)

	emptyIndex := p.Loader.EmptyIndex()
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
		p.Log.Infof("Create empty index (%s) success", emptyIndex)

		// put dummy data
		for i := 0; i < 2; i++ {
			_, err = client.Index().Index(emptyIndex).Type(p.Cfg.IndexType).BodyJson(&metric.Metric{
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
			return fmt.Errorf("failed to init empty index: %s", err)
		}
		p.Log.Infof("Put dummy data into empty index (%s) success", emptyIndex)
	}

	if p.Cfg.EnableRollover {
		for i := 0; i < 2; i++ {
			// remove this indices, avoid collisions with aliases
			_, err := client.DeleteIndex(p.indexPrefix + "-*-rollover").Do(ctx)
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("failed to delete index %s: %s", p.indexPrefix+"-*-rollover", err)
		}
	}
	return nil
}
