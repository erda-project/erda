// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

//import (
//	"testing"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/pkg/jsonstore/etcd"
//)
//
//func TestGetKeyVersion(t *testing.T) {
//	etcdclient, err := etcd.New()
//	assert.NoError(t, err)
//	s := Store{etcdClient: etcdclient}
//	keyVersionInfo, err := s.GetKeyVersion("9367fcdeeed94a809004b3f228c05a08", "ac643c07a95a433ca080ef58c04bc357")
//	assert.NoError(t, err)
//	spew.Dump(keyVersionInfo)
//}
