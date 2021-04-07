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

package kmstypes

import (
	"context"

	"github.com/pkg/errors"
)

const (
	CtxKeyConfigMap = "configMap"
)

// PluginCreateFn be used to create a kms plugin instance
type PluginCreateFn func(ctx context.Context) Plugin

var PluginFactory = map[PluginKind]PluginCreateFn{}

func RegisterPlugin(kind PluginKind, create PluginCreateFn) error {
	if !kind.Validate() {
		return errors.Errorf("invalid plugin kind: %s", kind)
	}
	if _, ok := PluginFactory[kind]; ok {
		return errors.Errorf("duplicate to register kms plugin: %s", kind)
	}
	PluginFactory[kind] = create
	return nil
}

// StoreCreateFn be used to create a kms plugin instance
type StoreCreateFn func(ctx context.Context) Store

var StoreFactory = map[StoreKind]StoreCreateFn{}

func RegisterStore(kind StoreKind, create StoreCreateFn) error {
	if !kind.Validate() {
		return errors.Errorf("failed to register store, invalid store kind: %s", kind)
	}
	if _, ok := StoreFactory[kind]; ok {
		return errors.Errorf("failed to register store, duplicate store kind: %s", kind)
	}
	StoreFactory[kind] = create
	return nil
}
