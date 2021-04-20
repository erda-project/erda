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

package etcd

//import (
//	"context"
//	"fmt"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/pkg/jsonstore/stm"
//)
//
//func TestCancelWatch(t *testing.T) {
//	s, err := New()
//	assert.Nil(t, err)
//	ctx, cancel := context.WithCancel(context.Background())
//	ch, err := s.Watch(ctx, "/etcd_test", true, false)
//	assert.Nil(t, err)
//	c := make(chan struct{})
//	go func() {
//		for range ch {
//
//		}
//		c <- struct{}{}
//	}()
//	time.Sleep(1 * time.Second)
//	cancel()
//	<-c
//}
//
//type stmtest struct {
//	A int
//	B string
//}
//
//func testSTMNormal(t *testing.T) {
//	s, err := New()
//	assert.Nil(t, err)
//	assert.Nil(t, s.NewSTM(func(stm stm.JSONStoreSTMOP) error {
//		stm.Put("/stmtest", stmtest{1, "2"})
//		var o stmtest
//		stm.Get("/stmtest", &o)
//		fmt.Printf("%+v\n", o) // debug print
//		stm.Remove("/stmtest")
//		return nil
//	}))
//}
//
//func testSTMRepeat(t *testing.T) {
//	s, err := New()
//	assert.Nil(t, err)
//	r := 0
//	ch0 := make(chan struct{}, 5)
//	ch1 := make(chan struct{}, 5)
//	ch3 := make(chan struct{}, 5)
//
//	f1 := func() {
//		assert.Nil(t, s.NewSTM(func(stm stm.JSONStoreSTMOP) error {
//			var o stmtest
//			stm.Get("/stmtest", &o)
//			fmt.Printf("%+v\n", o) // debug print
//			ch0 <- struct{}{}
//			<-ch1
//			stm.Put("/stmtest", stmtest{2, "3"})
//			stm.Get("/stmtest", &o)
//			r = o.A
//			return nil
//		}))
//		ch3 <- struct{}{}
//	}
//
//	f2 := func() {
//		assert.Nil(t, s.NewSTM(func(stm stm.JSONStoreSTMOP) error {
//			<-ch0
//			stm.Put("/stmtest", stmtest{3, "4"})
//			ch1 <- struct{}{}
//			ch1 <- struct{}{}
//			return nil
//		}))
//	}
//	go f1()
//	f2()
//	<-ch3
//	assert.Equal(t, 2, r)
//	s.Remove(context.Background(), "/stmtest")
//}
//
//func TestSTM(t *testing.T) {
//	t.Run("normal", testSTMNormal)
//	t.Run("repeat", testSTMRepeat)
//}
