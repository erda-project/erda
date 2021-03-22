package lru

import (
	"context"
	"testing"

	"github.com/erda-project/erda/pkg/jsonstore/mem"

	"github.com/stretchr/testify/assert"
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
