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
	"runtime"
	"strings"

	"github.com/olivere/elastic"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/pkg/netportal"
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
	body := jsonx.MarshalAndIntend(source)
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
		client, err := elastic.NewClient(options...)
		if err != nil {
			continue
		}
		if d.ClusterType == 1 {
			netp := &NetPortal{
				ClusterName: d.ClusterName,
			}
			client.RequestBuilder = netp.NewRequest
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

// NetPortal .
type NetPortal struct {
	Addr        string
	Domain      string
	ClusterName string
}

func (np *NetPortal) NewRequest(method, url string) (*elastic.Request, error) {
	req, err := netportal.NewNetportalRequest(np.ClusterName, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "elastic/"+elastic.Version+" ("+runtime.GOOS+"-"+runtime.GOARCH+")")
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return (*elastic.Request)(req), nil
}
