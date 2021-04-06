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

import (
	"errors"

	"github.com/erda-project/erda/apistructs"
)

func CheckAppMode(mode string) error {
	switch mode {
	case string(apistructs.ApplicationModeService), string(apistructs.ApplicationModeLibrary),
		string(apistructs.ApplicationModeBigdata), string(apistructs.ApplicationModeAbility),
		string(apistructs.ApplicationModeMobile), string(apistructs.ApplicationModeApi):
	default:
		return errors.New("invalid request, mode is invalid")
	}
	return nil
}
