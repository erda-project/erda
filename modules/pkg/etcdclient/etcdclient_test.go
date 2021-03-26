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
