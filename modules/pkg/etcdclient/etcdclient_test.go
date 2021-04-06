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

package etcdclient

//import (
//	"sync"
//	"testing"
//
//	"github.com/coreos/etcd/clientv3"
//)
//
//func TestSingleInstance(t *testing.T) {
//	var instance1 *clientv3.Client
//	var instance2 *clientv3.Client
//	waitGroup := sync.WaitGroup{}
//	waitGroup.Add(2)
//	go func() {
//		for i := 0; i < 100; i++ {
//			instance1, _ = NewEtcdClientSingleInstance()
//		}
//		waitGroup.Done()
//	}()
//	go func() {
//		for i := 0; i < 100; i++ {
//			instance2, _ = NewEtcdClientSingleInstance()
//		}
//		waitGroup.Done()
//	}()
//	waitGroup.Wait()
//	if instance1 != instance2 {
//		t.Failed()
//	}
//
//}
