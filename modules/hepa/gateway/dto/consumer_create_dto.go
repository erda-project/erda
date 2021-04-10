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
	CT_PRO = "project"
	CT_ORG = "org"
)

type OpenConsumerDto struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"-"`
}

type ConsumerCreateDto struct {
	OrgId        int    `json:"orgId"`
	ProjectId    string `json:"projectId"`
	Env          string `json:"env"`
	ConsumerId   string `json:"consumerId"`
	ConsumerName string `json:"consumerName"`
	Description  string `json:"description"`
}

func NewConsumerCreateDto() *ConsumerCreateDto {
	return &ConsumerCreateDto{ConsumerName: "default"}
}

func (dto ConsumerCreateDto) IsEmpty() bool {
	return len(dto.ProjectId) == 0 || len(dto.Env) == 0 || len(dto.ConsumerName) == 0
}
