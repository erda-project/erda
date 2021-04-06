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

package addon

import (
	"github.com/erda-project/erda/apistructs"
)

func Create(op AddonOperator, sg *apistructs.ServiceGroup) error {
	if err := op.Validate(sg); err != nil {
		return err
	}
	k8syml := op.Convert(sg)

	return op.Create(k8syml)
}

func Inspect(op AddonOperator, sg *apistructs.ServiceGroup) (*apistructs.ServiceGroup, error) {
	if err := op.Validate(sg); err != nil {
		return nil, err
	}
	return op.Inspect(sg)
}

func Remove(op AddonOperator, sg *apistructs.ServiceGroup) error {
	return op.Remove(sg)
}

func Update(op AddonOperator, sg *apistructs.ServiceGroup) error {
	if err := op.Validate(sg); err != nil {
		return err
	}
	k8syml := op.Convert(sg)

	return op.Update(k8syml)
}
