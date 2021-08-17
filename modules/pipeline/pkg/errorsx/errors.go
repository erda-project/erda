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

package errorsx

import (
	"fmt"
	"strings"
)

type errType string

var (
	platformAbnormalError errType = "platform-error"
	userAbnormalError     errType = "user-error"
)

type fundamental struct {
	msg string
	errType
}

func (f *fundamental) Error() string {
	return fmt.Sprintf("errType: %s, %s", f.errType, f.msg)
}

func New(message string) error {
	return &fundamental{
		msg: message,
	}
}

func Errorf(format string, args ...interface{}) error {
	return &fundamental{
		msg: fmt.Sprintf(format, args...),
	}
}

func PlatformErrorf(format string, args ...interface{}) error {
	return &fundamental{
		msg:     fmt.Sprintf(format, args...),
		errType: platformAbnormalError,
	}
}

func UserErrorf(format string, args ...interface{}) error {
	return &fundamental{
		msg:     fmt.Sprintf(format, args...),
		errType: userAbnormalError,
	}
}

func IsPlatformError(err error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*fundamental)
	if !ok {
		return false
	}
	return e.errType == platformAbnormalError
}

func IsUserError(err error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*fundamental)
	if !ok {
		return false
	}
	return e.errType == userAbnormalError
}

func IsContainUserError(err error) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("errType: %s", userAbnormalError))
}
