// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package docker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type client interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
}

type Watcher struct {
	sync.RWMutex
	ctx       context.Context
	stop      context.CancelFunc // 发送结束信号，停止监听
	startOnce sync.Once
	stopOnce  sync.Once

	client client
	cfg    *Config

	containers map[string]*container
	deleted    map[string]time.Time // 死亡容器上一次访问的时间map，用于定时清理内存中的容器

	lastValidTimestamp int64 // 上一次有效时间，用于监听docker事件
	lastEventTimestamp time.Time
}

var (
	instance *Watcher
)

func GetInstance() (*Watcher, error) {
	if instance == nil {
		return nil, errors.Errorf("docker instance is nil, create watcher first")
	}
	return instance, nil
}

func NewWatcher(cfg *Config) (*Watcher, error) {
	client, err := NewClient(cfg.Host, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}

	instance = newWatchWithClient(cfg, client)
	return instance, nil
}

func newWatchWithClient(cfg *Config, client client) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		client:     client,
		cfg:        cfg,
		ctx:        ctx,
		stop:       cancel,
		containers: make(map[string]*container),
		deleted:    make(map[string]time.Time),
	}
}

// 根据容器ID获取容器，容器不存在为nil，返回即内部存储的容器
func (w *Watcher) Container(ID string) *container {
	w.RLock()
	container := w.containers[ID]
	if container == nil {
		w.RUnlock()
		return nil
	}

	// 如果容器已死亡，更新访问时间
	_, ok := w.deleted[container.id]
	w.RUnlock()
	if ok {
		w.Lock()
		w.deleted[container.id] = time.Now()
		w.Unlock()
	}
	return container
}

// 返回全部容器map，返回与内部map无关
func (w *Watcher) Containers() map[string]*container {
	w.RLock()
	defer w.RUnlock()
	containers := make(map[string]*container)
	for k, v := range w.containers {
		containers[k] = v
	}
	return containers
}

func (w *Watcher) Start() error {
	var err error
	w.startOnce.Do(func() {
		w.lastValidTimestamp = time.Now().Unix()

		w.Lock()
		defer w.Unlock()
		// 获取全部容器
		var containers []*container
		containers, err = w.listContainers()
		if err != nil {
			return
		}

		for _, c := range containers {
			w.containers[c.id] = c
		}

		go w.watchWorker()
		go w.cleanupWorker()
	})
	return err
}

func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		w.stop()
	})
}

// 监听docker事件
func (w *Watcher) watchWorker() {
	for {
		select {
		case <-w.ctx.Done():
			logrus.Infof("docker: stop watch event")
			return
		default:
			logrus.Debugf("docker: start watch event")
			w.watch()
		}
	}
}

// 监听docker事件
func (w *Watcher) watch() {
	options := types.EventsOptions{
		Since:   fmt.Sprintf("%d", w.lastValidTimestamp),
		Filters: filters.NewArgs(),
	}
	options.Filters.Add("type", "container")

	logrus.Debugf("docker: watch events since %s", options.Since)
	// 请求超时
	ctx, cancel := context.WithTimeout(w.ctx, w.cfg.EventTimeout)
	defer cancel()
	events, errs := w.client.Events(ctx, options)

	// 创建一个定时ticket去判断是否长时间无event时，是则重启监听
	w.lastEventTimestamp = time.Now()
	intervalTicker := time.NewTicker(w.cfg.WatchInterval)
	defer intervalTicker.Stop()

	for {
		select {
		case event := <-events:
			// 监听到event
			logrus.Debugf("docker: get event: %v", event)

			w.lastValidTimestamp = event.Time
			w.lastEventTimestamp = time.Now()

			if event.Action == "start" || event.Action == "update" {
				container, err := w.getContainer(event.Actor.ID)
				if err != nil {
					logrus.Errorf("docker: fail to get container %s: %s", event.Actor.ID, err)
					continue
				}

				w.Lock()
				w.containers[event.Actor.ID] = container
				// 取消该容器的删除
				delete(w.deleted, event.Actor.ID)
				w.Unlock()
			} else if event.Action == "die" {
				w.Lock()
				w.deleted[event.Actor.ID] = time.Now()
				w.Unlock()
			}

		case err := <-errs:
			// docker event报错，重启监听
			if err == context.DeadlineExceeded {
				logrus.Infof("docker: request event timeout, restart")
			} else {
				logrus.Errorf("docker: fail to watch event: %s", err)
			}
			time.Sleep(1 * time.Second)
			return

		case <-intervalTicker.C:
			// 无event事件过长，重启监听
			if time.Since(w.lastEventTimestamp) > w.cfg.WatchTimeout {
				logrus.Infof("docker: no event received within %s, restart", w.cfg.WatchTimeout)
				time.Sleep(1 * time.Second)
				return
			}

		case <-w.ctx.Done():
			// 上下文结束，终结监听
			logrus.Infof("docker: stop watch")
			return
		}
	}
}

func (w *Watcher) listContainers() ([]*container, error) {
	logrus.Debugf("docker: list containers")

	ctx, cancel := context.WithTimeout(w.ctx, w.cfg.RequestTimeout)
	defer cancel()

	containers, err := w.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var result []*container
	for _, c := range containers {
		info, err := w.getContainer(c.ID)
		if err != nil {
			logrus.Errorf("docker: fail to get container %s: %s", c.ID, err)
			continue
		}
		result = append(result, info)
	}
	return result, nil
}

func (w *Watcher) getContainer(id string) (*container, error) {
	logrus.Debugf("docker: inspect container %s", id)

	ctx, cancel := context.WithTimeout(w.ctx, w.cfg.RequestTimeout)
	defer cancel()

	info, err := w.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	// name为空的容器过滤
	if len(info.Name) == 0 {
		return nil, errors.Errorf("docker: container %s no name", id)
	}

	envs := make(map[string]string)
	for _, env := range info.Config.Env {
		kv := strings.Split(env, "=")
		if len(kv) == 2 {
			envs[kv[0]] = kv[1]
		}
	}
	return &container{
		id:     info.ID,
		name:   info.Name,
		image:  info.Image,
		envs:   envs,
		labels: info.Config.Labels,
	}, nil
}

// 清理没有被使用的已删除容器
func (w *Watcher) cleanupWorker() {
	ticker := time.NewTicker(w.cfg.CleanupTimeout)
	for {
		select {
		case <-w.ctx.Done():
			logrus.Infof("docker: stop clean up")
			ticker.Stop()
			return
		case <-ticker.C:
			timeout := time.Now().Add(-w.cfg.CleanupTimeout)
			w.Lock()
			for id, lastAccess := range w.deleted {
				if lastAccess.Before(timeout) {
					logrus.Debugf("docker: deleted container %s timeout", id)
					delete(w.deleted, id)
					delete(w.containers, id)
				}
			}
			w.Unlock()

		}
	}
}
