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

type KongCredentialListDto struct {
	Total int64               `json:"total"`
	Data  []KongCredentialDto `json:"data"`
}

type KongCredentialDto struct {
	ConsumerId string `json:"consumer_id,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	Id         string `json:"id,omitempty"`
	// key-auth, sign-auth
	Key string `json:"key,omitempty"`
	// oauth2
	RedirectUrl interface{} `json:"redirect_uri,omitempty"`
	// v2
	RedirectUrls []string `json:"redirect_uris,omitempty"`
	Name         string   `json:"name,omitempty"`
	ClientId     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	// sign-auth, hmac-auth
	Secret string `json:"secret,omitempty"`
	// hmac-auth
	Username string `json:"username,omitempty"`
}

func (dto *KongCredentialDto) ToHmacReq() {
	if dto == nil {
		return
	}
	if dto.Username == "" {
		dto.Username = dto.Key
		dto.Key = ""
	}
}

func (dto *KongCredentialDto) ToHmacResp() {
	if dto == nil {
		return
	}
	if dto.Key == "" {
		dto.Key = dto.Username
		dto.Username = ""
	}
}

func (dto *KongCredentialDto) ToV2() {
	if dto == nil {
		return
	}
	if url, ok := dto.RedirectUrl.([]string); ok && len(url) > 0 {
		dto.RedirectUrls = url
		dto.RedirectUrl = nil
	}
}

func (dto *KongCredentialDto) Compatiable() {
	if dto == nil {
		return
	}
	dto.RedirectUrl = dto.RedirectUrls
}
