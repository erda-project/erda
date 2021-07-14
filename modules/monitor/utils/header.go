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

	"github.com/erda-project/erda-infra/pkg/transport"
)

func NewContextWithHeader(ctx context.Context) context.Context {
	header := transport.Header{}
	for k, vals := range transport.ContextHeader(ctx) {
		header.Append(k, vals...)
	}
	return transport.WithHeader(context.Background(), header)
}
