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

package timeout

//import (
//	"context"
//	"testing"
//	"time"
//
//	"github.com/erda-project/erda/pkg/jsonstore/mem"
//
//	"github.com/stretchr/testify/assert"
//)
//
//var memStore, _ = mem.New()
//var store, _ = New(1, memStore)
//
//func TestTimeoutNormal(t *testing.T) {
//	ctx := context.Background()
//	assert.Nil(t, store.Put(ctx, "k1", "v1"))
//	kv, err := store.Get(ctx, "k1")
//	assert.Nil(t, err)
//	assert.Equal(t, "k1", string(kv.Key))
//	assert.Equal(t, "v1", string(kv.Value))
//
//	time.Sleep(2 * time.Second)
//	_, err = store.Get(ctx, "k1")
//	assert.NotNil(t, err)
//
//}
//
//func TestTimeoutPrefixGet(t *testing.T) {
//	ctx := context.Background()
//
//	assert.Nil(t, store.Put(ctx, "/k2/p1", "v2"))
//	_, err := store.PrefixGet(ctx, "/k2")
//	assert.Nil(t, err)
//	assert.Nil(t, store.Put(ctx, "/k2/p2", "v2"))
//	assert.Nil(t, store.Put(ctx, "/k2/p3", "v2"))
//	assert.Nil(t, store.Put(ctx, "/k2/p4", "v2"))
//	assert.Nil(t, store.Put(ctx, "/k2/p5", "v2"))
//	kvs, err := store.PrefixGet(ctx, "/k2")
//	assert.Nil(t, err)
//	assert.Equal(t, 5, len(kvs))
//
//	time.Sleep(2 * time.Second)
//
//	kvs, err = store.PrefixGet(ctx, "/k2")
//	assert.Equal(t, 0, len(kvs))
//	assert.Nil(t, err)
//}
//
//func TestTimeoutPrefixRemove(t *testing.T) {
//	ctx := context.Background()
//	assert.Nil(t, store.Put(ctx, "/k3/p0", "v3"))
//	assert.Nil(t, store.Put(ctx, "/k3/p1", "v3"))
//	assert.Nil(t, store.Put(ctx, "/k3/p2", "v3"))
//	assert.Nil(t, store.Put(ctx, "/k3/p3", "v3"))
//	assert.Nil(t, store.Put(ctx, "/k3/p4", "v3"))
//	assert.Nil(t, store.Put(ctx, "/k3/p5", "v3"))
//
//	kvs, err := store.PrefixRemove(ctx, "/k3")
//	assert.Equal(t, 6, len(kvs))
//	assert.Nil(t, err)
//	kvs, err = store.PrefixRemove(ctx, "/k3")
//	assert.Equal(t, 0, len(kvs))
//	assert.Nil(t, err)
//}
//
//func TestTimeoutPutSameKey(t *testing.T) {
//	ctx := context.Background()
//	assert.Nil(t, store.Put(ctx, "k4", "v4"))
//	time.Sleep(600 * time.Millisecond)
//	assert.Nil(t, store.Put(ctx, "k4", "v5"))
//	v, err := store.Get(ctx, "k4")
//	assert.Nil(t, err)
//	assert.Equal(t, "v5", string(v.Value))
//	time.Sleep(800 * time.Millisecond)
//	_, err = store.Get(ctx, "k4")
//	assert.NotNil(t, err)
//}
