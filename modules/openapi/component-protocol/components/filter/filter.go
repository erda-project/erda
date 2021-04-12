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

package filter

type CommonFilter struct {
	Version    string                     `json:"version,omitempty"`
	Name       string                     `json:"name,omitempty"`
	Type       string                     `json:"type,omitempty"`
	Props      Props                      `json:"props,omitempty"`
	Operations map[OperationKey]Operation `json:"operations,omitempty"`
}

type Props struct {
	Delay uint64 `json:"delay,omitempty"`
}

type PropConditionKey string

func (k PropConditionKey) String() string { return string(k) }

type PropCondition struct {
	Key         PropConditionKey       `json:"key,omitempty"`
	Label       string                 `json:"label,omitempty"`
	EmptyText   string                 `json:"emptyText,omitempty"`
	Fixed       bool                   `json:"fixed,omitempty"`
	ShowIndex   int                    `json:"showIndex,omitempty"`
	HaveFilter  bool                   `json:"haveFilter,omitempty"`
	Type        PropConditionType      `json:"type,omitempty"`
	QuickSelect QuickSelect            `json:"quickSelect,omitempty"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Options     []PropConditionOption  `json:"options,omitempty"`
	CustomProps map[string]interface{} `json:"customProps,omitempty"`
}

type QuickSelect struct {
	Label        string       `json:"label,omitempty"`
	OperationKey OperationKey `json:"operationKey,omitempty"`
}

type PropConditionOption struct {
	Label string      `json:"label,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Icon  string      `json:"icon,omitempty"`
}

type PropConditionType string

var (
	PropConditionTypeSelect    PropConditionType = "select"
	PropConditionTypeInput     PropConditionType = "input"
	PropConditionTypeDateRange PropConditionType = "dateRange"
)

type StateKey string

type OperationKey string
type Operation struct {
	Key    OperationKey `json:"key,omitempty"`
	Reload bool         `json:"reload,omitempty"`
}

func (k OperationKey) String() string {
	return string(k)
}
