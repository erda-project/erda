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

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/orchestrator/conf"
)

type QueueEnum string

const (
	DEPLOY_CONTINUING QueueEnum = "DEPLOY_CONTINUING"
	RUNTIME_DELETING  QueueEnum = "RUNTIME_DELETING"
)

type PusherQueue struct {
	redisClient *redis.Client
}

func NewPusherQueue() *PusherQueue {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    conf.RedisMasterName(),
		SentinelAddrs: strings.Split(conf.RedisSentinels(), ","),
		Password:      conf.RedisPassword(),
	})
	return &PusherQueue{
		redisClient: redisClient,
	}
}

func (q *PusherQueue) Push(queue QueueEnum, item string) error {
	if ok, err := q.Lock(queue, item); err != nil {
		logrus.Errorf("[alert] failed to lock %v/%v, err: %v", queue, item, err)
		return err
	} else if !ok {
		logrus.Warnf("failed to lock %v/%v, already locked", queue, item)
		return nil
	}
	key := buildQueueNameInRedis(queue)
	_, err := q.redisClient.ZIncrBy(key, 1.0, item).Result()
	return err
}

func (q *PusherQueue) Pop(queue QueueEnum) (string, error) {
	key := buildQueueNameInRedis(queue)
	if items, err := q.redisClient.ZRevRange(key, 0, 0).Result(); err != nil {
		return "", err
	} else {
		if len(items) <= 0 {
			return "", nil
		}
		item := items[0]
		if cnt, err := q.redisClient.ZRem(key, item).Result(); err != nil {
			return "", nil
		} else {
			if cnt > 0 {
				return item, nil
			}
			return q.Pop(queue)
		}
	}
}

func (q *PusherQueue) List(queue QueueEnum) ([]string, error) {
	key := buildQueueNameInRedis(queue)
	if ret, err := q.redisClient.ZRange(key, 0, -1).Result(); err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}

func (q *PusherQueue) Lock(queue QueueEnum, item string) (bool, error) {
	key := buildLockKeyInRedis(queue, item)
	if tm, err := q.redisClient.TTL(key).Result(); err != nil {
		return false, err
	} else if tm > 0 {
		// already locked
		return false, nil
	}
	if _, err := q.redisClient.Set(key, item, 30*time.Second).Result(); err != nil {
		return false, err
	}
	return true, nil
}

func (q *PusherQueue) Unlock(queue QueueEnum, item string) (bool, error) {
	key := buildLockKeyInRedis(queue, item)
	if _, err := q.redisClient.Del(key).Result(); err != nil {
		return false, err
	}
	return true, nil
}

func buildQueueNameInRedis(queue QueueEnum) string {
	return "pqExq-" + string(queue)
}

func buildLockKeyInRedis(queue QueueEnum, item string) string {
	return "pqExq-" + string(queue) + "-lock-" + item
}
