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

package dto

import "github.com/erda-project/erda/modules/hepa/gateway/exdto"

type OpenLimitRuleDto struct {
	ConsumerId string                 `json:"consumerId"`
	PackageId  string                 `json:"packageId"`
	Method     string                 `json:"method"`
	ApiPath    string                 `json:"apiPath"`
	Limit      exdto.LimitType        `json:"limit"`
	KongConfig map[string]interface{} `json:"-"`
	ApiId      string                 `json:"-"`
}
