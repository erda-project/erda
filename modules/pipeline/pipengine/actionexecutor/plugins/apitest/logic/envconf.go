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

package logic

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

type EnvConfig struct {
	Name         string                        `env:"ACTION_NAME"`
	URL          string                        `env:"ACTION_URL" required:"true"`
	Method       string                        `env:"ACTION_METHOD" required:"true"`
	Params       []APIParam                    `env:"ACTION_PARAMS"`
	Headers      []APIHeader                   `env:"ACTION_HEADERS"`
	Body         apistructs.APIBody            `env:"ACTION_BODY"`
	OutParams    []apistructs.APIOutParam      `env:"ACTION_OUT_PARAMS"`
	Asserts      []APIAssert                   `env:"ACTION_ASSERTS"`
	GlobalConfig *apistructs.AutoTestAPIConfig `env:"AUTOTEST_API_GLOBAL_CONFIG"`

	MetaFile string `env:"METAFILE"`
}

func generateAPIInfoFromEnv(cfg EnvConfig) *apistructs.APIInfo {
	var params []apistructs.APIParam
	for _, p := range cfg.Params {
		params = append(params, p.convert())
	}
	var headers []apistructs.APIHeader
	for _, h := range cfg.Headers {
		headers = append(headers, h.convert())
	}
	var asserts []apistructs.APIAssert
	for _, a := range cfg.Asserts {
		asserts = append(asserts, a.convert())
	}
	return &apistructs.APIInfo{
		Name:      cfg.Name,
		URL:       cfg.URL,
		Method:    cfg.Method,
		Headers:   headers,
		Params:    params,
		Body:      cfg.Body,
		OutParams: cfg.OutParams,
		Asserts:   [][]apistructs.APIAssert{asserts}, // 目前有且只有一组断言
	}
}

type APIParam struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Desc  string      `json:"desc"`
}

func (p APIParam) convert() apistructs.APIParam {
	return apistructs.APIParam{
		Key:   p.Key,
		Value: strutil.String(p.Value),
		Desc:  p.Desc,
	}
}

type APIHeader struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Desc  string      `json:"desc"`
}

func (h APIHeader) convert() apistructs.APIHeader {
	return apistructs.APIHeader{
		Key:   h.Key,
		Value: strutil.String(h.Value),
		Desc:  h.Desc,
	}
}

type APIAssert struct {
	Arg      string      `json:"arg"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func (a APIAssert) convert() apistructs.APIAssert {
	return apistructs.APIAssert{
		Arg:      a.Arg,
		Operator: a.Operator,
		Value:    strutil.String(a.Value),
	}
}
