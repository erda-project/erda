package session

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/erda-project/erda-infra/providers/redis"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type RemoteManager struct {
	rds   redis.Interface
	cache gcache.Cache
}

func NewRemoteManager(rds redis.Interface) *RemoteManager {
	return &RemoteManager{
		rds: rds,
		cache: gcache.New(5).LRU().Expiration(1 * time.Hour).LoaderExpireFunc(func(sessionId interface{}) (interface{}, *time.Duration, error) {
			client := rds.DB()
			result, err := client.Get(fmt.Sprintf("%v", sessionId)).Result()
			if err != nil {
				return nil, nil, err
			}
			split := strings.Split(result, " ")
			if len(split) != 2 {
				return nil, nil, fmt.Errorf("invalid result: %s", result)
			}
			host := split[0]
			scheme := split[1]
			duration := 1 * time.Hour
			return &ServerInfo{Host: host, Scheme: scheme}, &duration, nil

		}).Build(),
	}
}

func (r *RemoteManager) Load(sessionId string) (*ServerInfo, error) {
	raw, err := r.cache.Get(sessionId)
	if err != nil {
		return nil, err
	}
	info, ok := raw.(*ServerInfo)
	if !ok {
		return nil, fmt.Errorf("invalid server info: %v", raw)
	}
	return info, nil
}

func (r *RemoteManager) Save(sessionId string, info *ServerInfo) error {
	if err := r.cache.Set(sessionId, info); err != nil {
		return err
	}
	go func() {
		client := r.rds.DB()
		err := client.Set(sessionId, fmt.Sprintf("%v %v", info.Host, info.Scheme), 0).Err()
		if err != nil {
			logrus.Errorf("[Mcp Proxy] save server info error: %v", err)
		}
	}()
	return nil
}
