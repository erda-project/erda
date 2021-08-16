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

package semver

import (
	"fmt"
	"regexp"
)

// Semantic Version
// see: https://semver.org
// see: https://github.com/semver/semver/issues/232#issuecomment-405596809
var SemverRegexp = regexp.MustCompile(`^(v)?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func Valid(ver string) bool {
	return SemverRegexp.MatchString(ver)
}

// New .
func New(major int, vers ...int) string {
	minor := 0
	patch := 0
	if len(vers) > 0 {
		if len(vers) > 0 {
			minor = vers[0]
		}
		if len(vers) > 1 {
			patch = vers[1]
		}
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
