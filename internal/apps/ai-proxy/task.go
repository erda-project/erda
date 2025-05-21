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

package ai_proxy

import (
	"context"
	"errors"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (p *provider) CleanupFile(ctx context.Context) error {
	poolSize := runtime.NumCPU() * 4
	logrus.Infof("cleanup file, max files total size is %v, pool size is %v", p.Config.MaxFileSize, poolSize)
	pool := workerpool.New(poolSize)

	var totalSize uint64    // total size
	var keepFileSize uint64 // which files should be kept
	var freedSize uint64    // freed size
	var orphanedSize uint64 // which files not in db

	objects, err := p.CloudStorage.ListObjects(ctx)
	if err != nil {
		return err
	}
	var willDelete = make(map[string]struct{})
	var lock sync.Mutex
	// find file object not in db
	for _, object := range objects {
		pool.Submit(func() {
			if _, err = p.Dao.McpFilesystemClient().GetFileByObjectKey(object.Key); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if object.LastModified != nil && object.LastModified.Add(7*24*time.Hour).After(time.Now()) {
						return
					}
					logrus.Infof("file %v not in db and meets condition, marked for deletion", object.Key)
					lock.Lock()
					defer lock.Unlock()
					willDelete[object.Key] = struct{}{}
					orphanedSize += uint64(object.Size)
				} else {
					logrus.Errorf("list file error: %v", err)
				}
			}
		})
	}
	pool.StopWait()

	files, err := p.Dao.McpFilesystemClient().ListMcpFiles()
	if err == nil && len(files) > 0 {
		sort.Slice(files, func(i, j int) bool {
			return !files[i].CreatedAt.Before(files[j].CreatedAt)
		})

		for _, file := range files {
			if file.Keep == "Y" {
				keepFileSize += uint64(file.FileSize)
				totalSize += uint64(file.FileSize)
				continue
			}
			if totalSize+uint64(file.FileSize) < p.Config.MaxFileSize {
				totalSize += uint64(file.FileSize)
				continue
			}
			freedSize += uint64(file.FileSize)
			willDelete[file.ObjectKey] = struct{}{}
		}
	}

	var keys []string
	for objectKey := range willDelete {
		keys = append(keys, objectKey)

		if err = p.Dao.McpFilesystemClient().DeleteFileByKey(objectKey); err != nil {
			logrus.Errorf("failed to delete file: %v", err)
		}
	}
	logrus.Infof("will delete %v files", len(willDelete))

	if err = p.CloudStorage.DeleteObjects(ctx, keys); err != nil {
		logrus.Errorf("delete file error: %v", err)
	}

	logrus.Infof("cleanup file done! Total size is %vkb, keep file size is %vkb, freed size is %vkb, orphaned size is %vkb", totalSize, keepFileSize, freedSize, orphanedSize)
	return nil
}
