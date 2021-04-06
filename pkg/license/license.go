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

package license

import (
	"encoding/json"
	"errors"
	"time"
)

var aesKey = "0123456789abcdef"

// ParseLicense 解析license
func ParseLicense(licenseKey string) (*License, error) {
	if licenseKey == "" {
		return nil, errors.New("licenseKey is empty")
	}
	bytes, err := AesDecrypt(licenseKey, aesKey)
	if err != nil {
		return nil, err
	}
	var license License
	err = json.Unmarshal([]byte(bytes), &license)
	return &license, err
}

type License struct {
	ExpireDate time.Time `json:"expireDate"`
	IssueDate  time.Time `json:"issueDate"`
	User       string    `json:"user"`
	Data       Data      `json:"data"`
}

type Data struct {
	MaxHostCount uint64 `json:"maxHostCount"`
}

func (license *License) IsExpired() bool {
	return license.ExpireDate.Before(time.Now())
}
