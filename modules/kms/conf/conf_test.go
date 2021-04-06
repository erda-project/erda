package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/kms/kmstypes"
)

func TestLoad(t *testing.T) {
	// panic
	shouldLoadPanic(t)

	// normal
	normalLoad(t)
}

func shouldLoadPanic(t *testing.T) {
	defer func() { recover() }()
	// panic logic
	Load()
	// should panic here, cannot reach the line below
	t.Errorf("shold have panicked")
}

const (
	envKeyKmsStoreKind  = "KMS_STORE_KIND"
	envKeyEtcdEndpoints = "ETCD_ENDPOINTS"
)

func normalLoad(t *testing.T) {
	_ = os.Setenv(envKeyKmsStoreKind, kmstypes.StoreKind_ETCD.String())
	_ = os.Setenv(envKeyEtcdEndpoints, "fake")
	defer func() { _ = os.Unsetenv(envKeyKmsStoreKind) }()

	Load()

	assert.Equal(t, KmsStoreKind(), kmstypes.StoreKind_ETCD)
	assert.Equal(t, EtcdEndpoints(), "fake")
	assert.Equal(t, ListenAddr(), ":3082")
	assert.False(t, Debug())
}
