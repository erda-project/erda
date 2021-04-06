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

package apitestsv2

import (
	"fmt"
	"reflect"

	"github.com/erda-project/erda/apistructs"
)

func checkBodyType(body apistructs.APIBody, expectType reflect.Kind) error {
	if body.Content == nil {
		return nil
	}
	_type := reflect.TypeOf(body.Content).Kind()
	if _type != expectType {
		return fmt.Errorf("invalid body type (%s) while Content-Type is %s", _type.String(), body.Type.String())
	}
	return nil
}
