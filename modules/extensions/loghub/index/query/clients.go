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

	"github.com/erda-project/erda/modules/extensions/loghub/index/query/db"
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
	Cluster    string
	URLs       string
	LogVersion string
	Indices    []string
	Entrys     []*IndexEntry
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

func (p *provider) getAllESClients() []*ESClient {
	list, err := p.db.LogDeployment.List()
	if err != nil {
		return nil
	}
	return p.getESClientsFromLogAnalyticsByLogDeployment("", list...)
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
	return p.getESClientsFromLogAnalyticsByLogDeployment(addon, list...)
}

func (p *provider) getESClientsFromLogAnalyticsByLogDeployment(addon string, logDeployments ...*db.LogDeployment) []*ESClient {
	type ESConfig struct {
		Security bool   `json:"securityEnable"`
		Username string `json:"securityUsername"`
		Password string `json:"securityPassword"`
	}
	var indices map[string]map[string][]*IndexEntry
	if p.C.IndexPreload {
		v := p.indices.Load()
		if v != nil {
			indices = v.(map[string]map[string][]*IndexEntry)
		}
	}
	var clients []*ESClient
	for _, d := range logDeployments {
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
			c := &ESClient{
				Client:     client,
				Cluster:    d.ClusterName,
				LogVersion: LogVersion2,
				URLs:       d.ESURL,
				Indices:    getLogIndices("rlogs-", addon),
			}
			clients = append(clients, c)

			if p.C.IndexPreload && indices != nil && len(addon) > 0 {
				if addons, ok := indices[d.ClusterName]; ok {
					c.Entrys = addons[addon]
				}
			}

		} else {
			clients = append(clients, &ESClient{
				Client:     client,
				Cluster:    d.ClusterName,
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
