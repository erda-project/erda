package queue

// func TestPusherQueue_Lock(t *testing.T) {
// 	q := &PusherQueue{
// 		redisClient: redis.NewClient(&redis.Options{
// 			Addr: "127.0.0.1:6379",
// 		}),
// 	}
//
// 	// first lock
// 	r, err := q.Lock(DEPLOY_CONTINUING, "123")
// 	if assert.NoError(t, err) {
// 		assert.True(t, r)
// 	}
//
// 	// try lock twice, must fail
// 	r, err = q.Lock(DEPLOY_CONTINUING, "123")
// 	if assert.NoError(t, err) {
// 		assert.False(t, r)
// 	}
//
// 	// unlock
// 	r, err = q.Unlock(DEPLOY_CONTINUING, "123")
// 	if assert.NoError(t, err) {
// 		assert.True(t, r)
// 	}
//
// 	// then lock success
// 	r, err = q.Lock(DEPLOY_CONTINUING, "123")
// 	if assert.NoError(t, err) {
// 		assert.True(t, r)
// 	}
// }
