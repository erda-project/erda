package etcd

// func TestRestart(t *testing.T) {
// 	etcd, err := New()
// 	assert.Nil(t, err)
// 	etcd2, err := New()
// 	assert.Nil(t, err)

// 	etcdcount := 0
// 	etcd2count := 0

// 	go etcd.Start(func(m *types.Message) *errors.DispatchError {
// 		if m.Content != "restore" {
// 			etcdcount++
// 			logrus.Info("etcd:", m)
// 		}
// 		return nil
// 	})
// 	time.Sleep(1 * time.Second)
// 	go etcd2.Start(func(m *types.Message) *errors.DispatchError {
// 		if m.Content != "restore" {
// 			etcd2count++
// 			logrus.Info("etcd2:", m)

// 		}
// 		return nil
// 	})

// 	time.Sleep(3 * time.Second)

// 	content := GenContent("test1")
// 	assert.Nil(t, InputEtcd(content))
// 	time.Sleep(500 * time.Millisecond)
// 	assert.True(t, IsCleanEtcd())
// 	etcd.Stop()

// 	content = GenContent("test1")
// 	assert.Nil(t, InputEtcd(content))
// 	time.Sleep(2 * time.Second)
// 	assert.True(t, IsCleanEtcd())
// 	etcd2.Stop()

// 	assert.Equal(t, 2, etcd2count+etcdcount)
// }

// func TestCancelLock(t *testing.T) {
// 	etcd, _ := New()
// 	etcd2, _ := New()
// 	go etcd.Start(func(m *types.Message) *errors.DispatchError {
// 		return nil
// 	})
// 	time.Sleep(1 * time.Second)
// 	go etcd2.Start(func(m *types.Message) *errors.DispatchError {
// 		return nil
// 	})
// 	etcd2.Stop()
// 	go etcd2.Start(func(m *types.Message) *errors.DispatchError {
// 		return nil
// 	})
// 	etcd2.Stop()
// 	etcd.Stop()
// }

// func TestFilter(t *testing.T) {
// 	lru, err := jsonstore.New(jsonstore.UseLruStore(10), jsonstore.UseMemStore())
// 	assert.Nil(t, err)
// 	m := &types.Message{Sender: "sender", Content: "dedede", Labels: map[types.LabelKey]interface{}{}}
// 	m2 := &types.Message{Sender: "sender", Content: "dedede2", Labels: map[types.LabelKey]interface{}{}}
// 	assert.True(t, filter(lru, m))
// 	assert.False(t, filter(lru, m))
// 	assert.True(t, filter(lru, m2))
// 	time.Sleep(6 * time.Second)
// 	assert.True(t, filter(lru, m))
// }

// func TestFilterConsecutive(t *testing.T) {
// 	lru, err := jsonstore.New(jsonstore.UseLruStore(10), jsonstore.UseMemStore())
// 	assert.Nil(t, err)
// 	// 消息到达时间点分别为1s, 3s, 7s, 9s, 应该把3s和9s的过滤掉
// 	m1 := &types.Message{Sender: "x", Content: "y", Labels: map[types.LabelKey]interface{}{}}
// 	time.Sleep(1 * time.Second)
// 	assert.True(t, filter(lru, m1))

// 	m2 := &types.Message{Sender: "x", Content: "y", Labels: map[types.LabelKey]interface{}{}}
// 	time.Sleep(2 * time.Second)
// 	assert.False(t, filter(lru, m2))

// 	m3 := &types.Message{Sender: "x", Content: "y", Labels: map[types.LabelKey]interface{}{}}
// 	time.Sleep(4 * time.Second)
// 	assert.True(t, filter(lru, m3))

// 	m4 := &types.Message{Sender: "x", Content: "y", Labels: map[types.LabelKey]interface{}{}}
// 	time.Sleep(2 * time.Second)
// 	assert.False(t, filter(lru, m4))
// }
