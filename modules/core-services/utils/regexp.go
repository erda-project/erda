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

package utils

import "regexp"

// org regexp modify from pkg/strutil/regexp.go/reg, but org name can not be pure digital
var orgReg = regexp.MustCompile(`^(?:[a-z]+|[0-9]+[a-z]+|[0-9]+[-]+[a-z0-9])+(?:(?:(?:[-]*)[a-z0-9]+)+)?$`)

// IsValidOrgName check org name can contain a-z0-9- but can not pure 0-9
func IsValidOrgName(repo string) bool {
	return orgReg.MatchString(repo)
}
