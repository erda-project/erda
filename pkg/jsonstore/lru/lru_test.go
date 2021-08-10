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

package lru

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/jsonstore/mem"
)

var objects = [][2]string{
	{"k0", "v0"},
	{"k1", "v1"},
	{"k2", "v2"},
	{"k3", "v3"},
	{"k4", "v4"},
}
var memStore, _ = mem.New()
var store, _ = New(4, memStore)

func TestLruMiss(t *testing.T) {
	ctx := context.Background()

	for _, o := range objects {
		err := store.Put(ctx, o[0], o[1])
		assert.Nil(t, err)
	}
	_, err := store.Get(ctx, "k0")
	assert.NotNil(t, err)
	assert.Equal(t, "not found", err.Error())
}

func TestLruGetUpdate(t *testing.T) {
	ctx := context.Background()

	for _, o := range objects {
		err := store.Put(ctx, o[0], o[1])
		assert.Nil(t, err)
	}
	_, err := store.Get(ctx, "k3")
	assert.Nil(t, err)
	assert.Equal(t, "k3", store.keyList.Front().Value.(string))
}
