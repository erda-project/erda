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

package cms

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
)

type ConfigManager interface {
	// IdempotentCreateNs create ns idempontently
	IdempotentCreateNs(ctx context.Context, ns string) error
	// IdempotentDeleteNs delete ns and its configs idempotently
	IdempotentDeleteNs(ctx context.Context, ns string) error
	// PrefixListNs list ns by prefix
	PrefixListNs(ctx context.Context, nsPrefix string) ([]*pb.PipelineCmsNs, error)
	// UpdateConfigs create if not exists; update if already exist
	UpdateConfigs(ctx context.Context, ns string, kvs map[string]*pb.PipelineCmsConfigValue) error
	// DeleteConfigs delete configs by keys; if ns not exists, return nil
	DeleteConfigs(ctx context.Context, ns string, keys ...string) error
	// GetConfigs get configs: if ns is empty, return nil; if keys is not empty, return specified configs
	GetConfigs(ctx context.Context, ns string, globalDecrypt bool, keys ...*pb.PipelineCmsConfigKey) (map[string]*pb.PipelineCmsConfigValue, error)
}

func transformKeysToStrSlice(keys ...*pb.PipelineCmsConfigKey) []string {
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key.Key)
	}
	return result
}
