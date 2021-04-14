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

package cache

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/erda-project/erda/modules/gittar/models"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

type AtomicInt int64

type MysqlCache struct {
	mutex      sync.RWMutex
	typeName   string
	hits, gets AtomicInt
	db         *models.DBClient
}

func NewMysqlCache(typeName string, db *models.DBClient) *MysqlCache {
	return &MysqlCache{
		typeName: typeName,
		db:       db,
	}
}

//Status return the status of cache
func (c *MysqlCache) Status() *gitmodule.CacheStatus {
	status := &gitmodule.CacheStatus{
		MaxItemSize: -1,
		Hits:        c.hits.Get(),
		Gets:        c.gets.Get(),
	}
	currentSize := -1
	c.db.Model(&models.RepoCache{}).Where("type_name = ?", c.typeName).Count(&currentSize)
	status.CurrentSize = currentSize
	return status
}

//Get value with key
func (c *MysqlCache) Get(key string, outValue interface{}) error {
	c.gets.Add(1)
	var repoCache models.RepoCache
	err := c.db.Where("type_name = ? and key_name = ? ", c.typeName, key).First(&repoCache).Error
	if err == nil {
		c.hits.Add(1)
		err := json.Unmarshal([]byte(repoCache.Value), outValue)
		return err
	}
	return errors.New("key not found")
}

//Set a value with key
func (c *MysqlCache) Set(key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var repoCache models.RepoCache
	err = c.db.Where("type_name = ? and key_name = ? ", c.typeName, key).First(&repoCache).Error
	if err != nil {
		//不存在 新增
		repoCache := models.RepoCache{
			TypeName:  c.typeName,
			KeyName:   key,
			Value:     string(bytes),
			CreatedAt: time.Now(),
		}
		err := c.db.Create(&repoCache).Error
		if err != nil {
			return err
		}
	} else {
		//已经存在 更新
		now := time.Now()
		repoCache.Value = string(bytes)
		repoCache.UpdatedAt = &now
		err := c.db.Save(&repoCache).Error
		if err != nil {
			return err
		}
	}
	return nil
}

//Delete delete the key
func (c *MysqlCache) Delete(key string) error {
	err := c.db.Where("type_name = ? and key_name = ? ", c.typeName, key).Delete(models.RepoCache{}).Error
	return err
}

// Add atomically adds n to i.
func (i *AtomicInt) Add(n int64) {
	atomic.AddInt64((*int64)(i), n)
}

// Get atomically gets the value of i.
func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}
