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

package cmd

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/version"
	. "github.com/erda-project/erda/tools/cli/command"
)

var VERSION = Command{
	Name:      "version",
	ShortHelp: "Show dice version info",
	Example:   `$ dice version`,
	Run:       RunVersion,
}

func RunVersion(ctx *Context) error {
	fmt.Println(version.String())

	return nil
}
