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
