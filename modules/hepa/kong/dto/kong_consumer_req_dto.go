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

type KongConsumerReqDto struct {
	// 用户名，必填
	Username string `json:"username"`
	// 第三方唯一id，选填
	CustomId string `json:"custom_id,omitempty"`
}

func (dto KongConsumerReqDto) IsEmpty() bool {
	return len(dto.Username) == 0
}

func NewKongConsumerReqDto(name string, customId string) *KongConsumerReqDto {
	return &KongConsumerReqDto{
		Username: name,
		CustomId: customId,
	}
}
