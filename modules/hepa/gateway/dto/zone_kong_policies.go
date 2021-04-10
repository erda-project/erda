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

const (
	BASE_PRIORITY                 = 0
	WILDCARD_DOMAIN_BASE_PRIORITY = 1000
)

type ZoneKongPolicies struct {
	Id          string `json:"id"`
	Regex       string `json:"regex"`
	Enables     string `json:"enable"`
	Disables    string `json:"disable"`
	Priority    int    `json:"priority"`
	PackageName string `json:"packageName"`
	ProjectId   string `json:"-"`
	Env         string `json:"-"`
}

type SortByRegexList []ZoneKongPolicies

func (list SortByRegexList) Len() int { return len(list) }

func (list SortByRegexList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list SortByRegexList) Less(i, j int) bool {
	return list[i].Priority > list[j].Priority
}
