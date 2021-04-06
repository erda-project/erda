package uc

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type UserInfo struct {
	Name      string      `json:"name,omitempty"`
	Username  string      `json:"username,omitempty"`
	Nickname  string      `json:"nickname,omitempty"`
	AvatarUrl string      `json:"avatar_url,omitempty"`
	UserId    interface{} `json:"user_id,omitempty"`
}

// uc 2.0
type UserInfoDto struct {
	AvatarUrl string      `json:"avatarUrl,omitempty"`
	Email     string      `json:"email,omitempty"`
	UserId    interface{} `json:"id,omitempty"`
	NickName  string      `json:"nickName,omitempty"`
	Phone     string      `json:"phone,omitempty"`
	RealName  string      `json:"realName,omitempty"`
	Username  string      `json:"username,omitempty"`
}

func (u *UserInfoDto) Convert() (string, error) {
	switch u.UserId.(type) {
	case string:
		return u.UserId.(string), nil
	case int:
		return strconv.Itoa(u.UserId.(int)), nil
	case int64:
		return strconv.FormatInt(u.UserId.(int64), 10), nil
	case float64:
		return fmt.Sprintf("%g", u.UserId.(float64)), nil
	default:
		return "", errors.Errorf("invalid type of %v", reflect.TypeOf(u.UserId))
	}
}

func (u *UserInfoDto) GetUsername() string {
	if u.NickName != "" {
		return u.NickName
	}
	if u.RealName != "" {
		return u.RealName
	}
	if u.Phone != "" {
		return u.Phone
	}
	if u.Email != "" {
		return u.Email
	}

	// 由 uc 2.0 生成
	return u.Username
}

// Deprecated
func (u *UserInfo) Convert() (string, error) {
	switch u.UserId.(type) {
	case string:
		return u.UserId.(string), nil
	case int:
		return strconv.Itoa(u.UserId.(int)), nil
	case int64:
		return strconv.FormatInt(u.UserId.(int64), 10), nil
	case float64:
		return fmt.Sprintf("%g", u.UserId.(float64)), nil
	default:
		return "", errors.Errorf("invalid type of %v", reflect.TypeOf(u.UserId))
	}
}

// Deprecated
func (u *UserInfo) UserName() string {
	if u == nil {
		return ""
	}
	if len(u.Name) != 0 {
		return u.Name
	}
	if len(u.Nickname) != 0 {
		return u.Nickname
	}
	if u.Username != "" {
		return u.Username
	}
	logrus.Errorf("can not cat any user name from user=%v", u)
	return ""
}
