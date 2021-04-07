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

package servicegroup

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

func (s ServiceGroupImpl) KillPod(ctx context.Context, namespace string, name string, containerID string) error {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return err
	}

	_, err := s.handleKillPod(ctx, &sg, containerID)
	if err != nil {
		return err
	}
	return nil
}
