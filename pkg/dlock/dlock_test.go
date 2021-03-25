package dlock

//import (
//	"context"
//	"fmt"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestCancel(t *testing.T) {
//	l1, err := New("dlock-testcancel", func() {}, WithTTL(5))
//	assert.Nil(t, err)
//	l2, err := New("dlock-testcancel", func() {})
//	assert.Nil(t, err)
//	ctx, cancel := context.WithCancel(context.Background())
//
//	go func() {
//		time.Sleep(1 * time.Second)
//		cancel()
//	}()
//
//	ch := make(chan struct{})
//	ch2 := make(chan struct{})
//	go func() {
//		ti := time.NewTimer(2 * time.Second)
//		select {
//		case <-ch: // fine
//			ch2 <- struct{}{}
//		case <-ti.C:
//			assert.Nil(t, 1)
//			ch2 <- struct{}{}
//		}
//	}()
//
//	l1.Lock(context.Background())
//	l2.Lock(ctx)
//	ch <- struct{}{}
//	<-ch2
//	l1.Unlock()
//	l2.Unlock()
//}
//
//func TestMultiLock(t *testing.T) {
//	l, err := New("dlock-multilock", func() {})
//	assert.Nil(t, err)
//	l.Lock(context.Background())
//	l.Lock(context.Background())
//	l.Lock(context.Background())
//	l.Lock(context.Background())
//	b, err := l.IsOwner()
//	assert.Nil(t, err)
//	assert.True(t, b)
//	l.Unlock()
//	b, err = l.IsOwner()
//	assert.Nil(t, err)
//	assert.False(t, b)
//}
//
//func TestLostLock(t *testing.T) {
//	l, err := New("dlock-lost-key", func() { fmt.Println("dlock lost") }, WithTTL(5))
//	assert.NoError(t, err)
//	err = l.Lock(context.Background())
//	assert.NoError(t, err)
//	time.Sleep(time.Second * 20)
//	// stop etcd to test dlock lost
//	// will print "dlock lost"
//}
//
//func TestDLock_UnlockAndClose(t *testing.T) {
//	l, err := New("my-test-dlock", func() { fmt.Println("should not print these") }, WithTTL(5))
//	assert.NoError(t, err)
//	err = l.Lock(context.Background())
//	assert.NoError(t, err)
//	defer l.UnlockAndClose()
//	// will not lost dlock
//}
