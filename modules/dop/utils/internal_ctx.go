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
	"context"

	"github.com/erda-project/erda/pkg/common/apis"
)

// WithInternalClientContext TODO bad hard-coded "dop", you should get it in a-native-way, quite like get module name by erda-infra's ability.
func WithInternalClientContext(ctx context.Context) context.Context {
	return apis.WithInternalClientContext(ctx, "dop")
}
