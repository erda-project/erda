package etcd

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

func TestGetKeyVersion(t *testing.T) {
	etcdclient, err := etcd.New()
	assert.NoError(t, err)
	s := Store{etcdClient: etcdclient}
	keyVersionInfo, err := s.GetKeyVersion("9367fcdeeed94a809004b3f228c05a08", "ac643c07a95a433ca080ef58c04bc357")
	assert.NoError(t, err)
	spew.Dump(keyVersionInfo)
}
