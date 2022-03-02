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

package leaderworker

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func (p *provider) ListenPrefix(ctx context.Context, prefix string, putHandler, deleteHandler func(context.Context, *clientv3.Event)) {
	for func() bool {
		wctx, wcancel := context.WithCancel(ctx)
		defer wcancel()
		wch := p.EtcdClient.Watch(wctx, prefix, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return false
			case resp, ok := <-wch:
				if !ok {
					return true
				} else if resp.Err() != nil {
					p.Log.Errorf("failed to watch etcd prefix %s, error: %v", prefix, resp.Err())
					return true
				}
				for _, ev := range resp.Events {
					if ev.Kv == nil {
						continue
					}
					switch ev.Type {
					case mvccpb.PUT:
						if putHandler != nil {
							putHandler(wctx, ev)
						}
					case mvccpb.DELETE:
						if deleteHandler != nil {
							deleteHandler(wctx, ev)
						}
					}
				}
			}
		}
	}() {
	}
}
