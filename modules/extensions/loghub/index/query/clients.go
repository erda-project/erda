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

package query

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/go-utils/reflectx"

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

func (c *ESClient) printSearchSource(searchSource *elastic.SearchSource) (string, error) {
	source, err := searchSource.Source()
	if err != nil {
		return "", fmt.Errorf("invalid search source: %s", err)
	}
	body := jsonx.MarshalAndIndent(source)
	body = c.URLs + "\n" + strings.Join(c.Indices, ",") + "\n" + body
	fmt.Println(body)
	return body, nil
}

func (p *provider) getESClients(orgID int64, req *LogRequest) []*ESClient {
	if len(req.ClusterName) > 0 || len(req.Addon) > 0 {
		if len(req.ClusterName) <= 0 || len(req.Addon) <= 0 {
			return nil
		}
		return p.getESClientsFromLogAnalyticsByCluster(strings.ReplaceAll(req.Addon, "*", ""), req.ClusterName)
	}
	filters := make(map[string]string)
	for _, item := range req.Filters {
		filters[item.Key] = item.Value
	}
	if filters["origin"] == "sls" {
		return p.getCenterESClients("sls-*")
	} else if filters["origin"] == "dice" {
		clients := p.getESClientsFromLogAnalytics(orgID)
		if len(clients) <= 0 {
			return p.getCenterESClients("rlogs-*")
		}
		return clients
	} else if filters["origin"] != "" {
		return p.getCenterESClients("__not-exist__*")
	}
	clients := append(p.getCenterESClients("sls-*"), p.getESClientsFromLogAnalytics(orgID)...)
	return clients
}

func (p *provider) getCenterESClients(indices ...string) []*ESClient {
	if p.C.QueryBackES {
		return []*ESClient{
			{Client: p.client, URLs: "-", Indices: indices},
			{Client: p.backClient, URLs: "-b", Indices: indices},
		}
	}
	return []*ESClient{
		{Client: p.client, URLs: "-", Indices: indices},
	}
}

func (p *provider) getESClientsFromLogAnalytics(orgID int64) []*ESClient {
	clusters, err := p.bdl.ListClusters("", uint64(orgID))
	if err != nil {
		return nil
	}
	var clusterNames []string
	for _, c := range clusters {
		clusterNames = append(clusterNames, c.Name)
	}
	return p.getESClientsFromLogAnalyticsByCluster("", clusterNames...)
}

func (p *provider) getESClientsFromLogAnalyticsByCluster(addon string, clusterNames ...string) []*ESClient {
	list, err := p.db.LogDeployment.QueryByClusters(clusterNames...)
	if err != nil {
		return nil
	}
	type ESConfig struct {
		Security bool   `json:"securityEnable"`
		Username string `json:"securityUsername"`
		Password string `json:"securityPassword"`
	}
	var clients []*ESClient
	for _, d := range list {
		if len(d.ESURL) <= 0 {
			continue
		}
		options := []elastic.ClientOptionFunc{
			elastic.SetURL(strings.Split(d.ESURL, ",")...),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(false),
		}
		if len(d.ESConfig) > 0 {
			var cfg ESConfig
			err := json.Unmarshal(reflectx.StringToBytes(d.ESConfig), &cfg)
			if err == nil {
				if cfg.Security && (cfg.Username != "" || cfg.Password != "") {
					options = append(options, elastic.SetBasicAuth(cfg.Username, cfg.Password))
				}
			}
		}
		if d.ClusterType == 1 {
			options = append(options, elastic.SetHttpClient(newHTTPClient(d.ClusterName)))
		}
		client, err := elastic.NewClient(options...)
		if err != nil {
			continue
		}
		d.CollectorURL = strings.TrimSpace(d.CollectorURL)
		if len(d.CollectorURL) > 0 {
			clients = append(clients, &ESClient{
				Client:     client,
				LogVersion: LogVersion2,
				URLs:       d.ESURL,
				Indices:    getLogIndices("rlogs-", addon),
			})
		} else {
			clients = append(clients, &ESClient{
				Client:     client,
				LogVersion: LogVersion1,
				URLs:       d.ESURL,
				Indices:    getLogIndices("spotlogs-", addon),
			})
		}
	}
	return clients
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
