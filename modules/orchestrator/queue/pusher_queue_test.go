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
