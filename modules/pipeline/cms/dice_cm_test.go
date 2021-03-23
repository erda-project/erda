package cms_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/cms"
)

var (
	cm cms.ConfigManager
)

const (
	ns = "local-test"
)

var once sync.Once

func initDiceCM() {
	once.Do(func() {
		fmt.Println("init dice cm")
		os.Setenv("CMDB_ADDR", "cmdb.default.svc.cluster.local:9093")
		cm = cms.NewDiceCM("fake-project-id-")
	})
}

func TestDiceCM_IdempotentCreateNS(t *testing.T) {
	initDiceCM()
	err := cm.IdempotentCreateNS(context.Background(), ns)
	assert.NoError(t, err)

	err = cm.IdempotentCreateNS(context.Background(), ns)
	assert.NoError(t, err)
}

func TestDiceCM_IdempotentDeleteNS(t *testing.T) {
	initDiceCM()
	err := cm.IdempotentDeleteNS(context.Background(), ns)
	assert.NoError(t, err)

	err = cm.IdempotentDeleteNS(context.Background(), ns)
	assert.NoError(t, err)
}

func TestDiceCM_UpdateConfigs(t *testing.T) {
	initDiceCM()
	kvs := map[string]apistructs.PipelineCmsConfigValue{
		"c": {Value: "cc", EncryptInDB: true},
		"d": {Value: "dd", EncryptInDB: true},
	}
	err := cm.UpdateConfigs(context.Background(), ns, kvs)
	assert.NoError(t, err)

	err = cm.UpdateConfigs(context.Background(), ns, kvs)
	assert.NoError(t, err)
}

func TestDiceCM_DeleteConfigs(t *testing.T) {
	initDiceCM()
	// delete namespace
	assert.NoError(t, cm.IdempotentDeleteNS(context.Background(), ns))
	// add before delete
	kvs := map[string]apistructs.PipelineCmsConfigValue{
		"c": {Value: "cc", EncryptInDB: true},
		"d": {Value: "dd", EncryptInDB: true},
	}
	err := cm.UpdateConfigs(context.Background(), ns, kvs)
	assert.NoError(t, err)

	// delete
	keys := []string{"a", "b", "c"}
	err = cm.DeleteConfigs(context.Background(), ns, keys...)
	assert.NoError(t, err)

	err = cm.DeleteConfigs(context.Background(), ns, keys...)
	assert.NoError(t, err)

	kvs, err = cm.GetConfigs(context.Background(), ns, true)
	assert.NoError(t, err)
	assert.True(t, len(kvs) == 1)

	// delete
	keys = []string{"d"}
	err = cm.DeleteConfigs(context.Background(), ns, keys...)
	assert.NoError(t, err)

	err = cm.DeleteConfigs(context.Background(), ns, keys...)
	assert.NoError(t, err)

	kvs, err = cm.GetConfigs(context.Background(), ns, true)
	assert.NoError(t, err)
	assert.True(t, len(kvs) == 0)
}

func TestDiceCM_GetConfigs(t *testing.T) {
	initDiceCM()
	// delete namespace
	assert.NoError(t, cm.IdempotentDeleteNS(context.Background(), ns))
	// add before delete
	kvs := map[string]apistructs.PipelineCmsConfigValue{
		"c": {Value: "cc", EncryptInDB: true},
		"d": {Value: "dd", EncryptInDB: true},
	}
	err := cm.UpdateConfigs(context.Background(), ns, kvs)
	assert.NoError(t, err)

	kvs, err = cm.GetConfigs(context.Background(), ns, true)
	assert.NoError(t, err)
	assert.True(t, len(kvs) == 2)

	// add before delete
	kvs = map[string]apistructs.PipelineCmsConfigValue{
		"a": {Value: "aa", EncryptInDB: true},
	}
	err = cm.UpdateConfigs(context.Background(), ns, kvs)
	assert.NoError(t, err)

	kvs, err = cm.GetConfigs(context.Background(), ns, true)
	assert.NoError(t, err)
	assert.True(t, len(kvs) == 3)
}
