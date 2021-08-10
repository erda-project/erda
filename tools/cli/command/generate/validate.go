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

package main

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/tools/cli/command"
)

var (
	ErrName                = errors.New("not provide Name")
	ErrOptionalArgPosition = errors.New("optional Arg should be the last arg")
	ErrOptionalArgNum      = errors.New("too many optional arg, support only 1 optional arg yet")
)

func validate(cmd command.Command, fname string) error {
	if cmd.Name == "" {
		return errors.Wrap(ErrName, fname)
	}

	optionalArgNum := 0
	for _, arg := range cmd.Args {
		if arg.IsOption() {
			optionalArgNum++
		} else {
			if optionalArgNum > 0 {
				return errors.Wrap(ErrOptionalArgPosition, fname)
			}
		}
	}
	if optionalArgNum > 1 {
		return errors.Wrap(ErrOptionalArgNum, fname)
	}
	return nil
}
