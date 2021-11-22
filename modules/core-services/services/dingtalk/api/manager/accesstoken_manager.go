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

package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda/modules/core-services/services/dingtalk/api/native"
)

type DingtalkAccessTokenManager interface {
	GetAccessToken(appKey string) (string, error)
}

var appKeySecrets = make(map[string]string)
var requestLocks = make(map[string]*sync.Mutex)

func (p *provider) GetAccessTokenManager(appKey, appSecret string) DingtalkAccessTokenManager {
	if secret, ok := appKeySecrets[appKey]; ok && secret == appSecret {
		return p
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if secret, ok := appKeySecrets[appKey]; ok && secret == appSecret {
		return p
	}
	appKeySecrets[appKey] = appSecret
	requestLocks[appKey] = &sync.Mutex{}
	return p
}

func (p *provider) GetAccessToken(appKey string) (string, error) {
	cacheKey := p.getAccessTokenCacheKey(appKey)
	result, err := p.Redis.Get(cacheKey).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}
	if len(result) > 0 {
		return result, nil
	}

	secret, ok := appKeySecrets[appKey]
	if !ok {
		return "", fmt.Errorf("appSecret not registered")
	}
	requestLock, ok := requestLocks[appKey]
	if !ok {
		return "", fmt.Errorf("request lock is nil")
	}

	// todo: use chan to load accessToken async
	requestLock.Lock()
	defer requestLock.Unlock()

	accessToken, expireIn, err := native.GetAccessToken(appKey, secret)
	if err != nil {
		return "", err
	}
	_, err = p.Redis.Set(cacheKey, accessToken, time.Duration(expireIn)*time.Second).Result()
	return accessToken, nil
}

func (p *provider) getAccessTokenCacheKey(appKey string) string {
	return "erda_dingtalk_ak_" + appKey
}
