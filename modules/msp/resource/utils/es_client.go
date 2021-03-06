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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/modules/msp/instance/db"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

// log versions
const (
	LogVersion1 = "1.0.0"
	LogVersion2 = "2.0.0"
)

// ESClient .
type ESClient struct {
	*elastic.Client
	URLs       string
	LogVersion string
	Indices    []string
}

func (c *ESClient) CreateIndexWithAlias(index string, alias string) error {
	ctx := context.Background()
	createIndex, err := c.CreateIndex(index).BodyJson(map[string]interface{}{
		"aliases": map[string]interface{}{
			alias: map[string]interface{}{},
		},
	}).Do(ctx)
	if err != nil {
		return err
	}

	if !createIndex.Acknowledged {
		return fmt.Errorf("response code error")
	}

	return nil
}

func GetESClientsFromLogAnalytics(logDeployment *db.LogDeployment, addon string) *ESClient {
	type ESConfig struct {
		Security bool   `json:"securityEnable"`
		Username string `json:"securityUsername"`
		Password string `json:"securityPassword"`
	}

	if len(logDeployment.EsUrl) <= 0 {
		return nil
	}
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(strings.Split(logDeployment.EsUrl, ",")...),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	}
	if len(logDeployment.EsConfig) > 0 {
		var cfg ESConfig
		err := json.Unmarshal(reflectx.StringToBytes(logDeployment.EsConfig), &cfg)
		if err == nil {
			if cfg.Security && (cfg.Username != "" || cfg.Password != "") {
				options = append(options, elastic.SetBasicAuth(cfg.Username, cfg.Password))
			}
		}
	}
	if logDeployment.ClusterType == 1 {
		options = append(options, elastic.SetHttpClient(newHTTPClient(logDeployment.ClusterName)))
	}
	client, err := elastic.NewClient(options...)
	if err != nil {
		return nil
	}
	logDeployment.CollectorUrl = strings.TrimSpace(logDeployment.CollectorUrl)
	if len(logDeployment.CollectorUrl) > 0 {
		return &ESClient{
			Client:     client,
			LogVersion: LogVersion2,
			URLs:       logDeployment.EsUrl,
			Indices:    getLogIndices("rlogs-", addon),
		}
	} else {
		return &ESClient{
			Client:     client,
			LogVersion: LogVersion1,
			URLs:       logDeployment.EsUrl,
			Indices:    getLogIndices("spotlogs-", addon),
		}
	}
}

func getLogIndices(prefix, addon string) []string {
	if len(addon) > 0 {
		return []string{prefix + addon, prefix + addon + "-*"}
	}
	return []string{prefix + "*"}
}

func newHTTPClient(clusterName string) *http.Client {
	hclient := httpclient.New(httpclient.WithClusterDialer(clusterName))
	t := hclient.BackendClient().Transport.(*http.Transport)
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           t.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
