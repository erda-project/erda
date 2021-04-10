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

package ver

import (
	"fmt"
	"runtime"
)

var (
	GitCommit string
	Built     string
	GoVersion string = runtime.Version()
)

func String() string {
	if len(GitCommit) > 12 {
		GitCommit = GitCommit[0:12]
	}
	return fmt.Sprintf("GitCommit: %s, Built: %s, GoVersion: %s", GitCommit, Built, GoVersion)
}
