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

package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/olivere/elastic"

	"github.com/erda-project/erda/modules/core/monitor/storekit/elasticsearch/index/loader"
)

type clearRequest struct {
	waitCh chan struct{}
	list   []string
}

func (p *provider) runCleanIndices(ctx context.Context) {
	p.loader.WaitAndGetIndices(ctx)
	p.Log.Infof("run indices clean with interval(%v)", p.Cfg.CheckInterval)
	defer p.Log.Infof("exit indices clean")
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			err := p.CleanIndices(ctx, func(*loader.IndexEntry) bool { return true })
			if err != nil {
				p.Log.Errorf("failed to CleanIndices: %s", err)
			}
		case req := <-p.clearCh:
			p.deleteIndices(req.list)
			if req.waitCh != nil {
				close(req.waitCh)
			}
		case <-ctx.Done():
			return
		}
		timer.Reset(p.Cfg.CheckInterval)
	}
}

// CleanIndices .
func (p *provider) CleanIndices(ctx context.Context, filter loader.Matcher) error {
	indices := p.loader.AllIndices()
	if indices == nil {
		return nil
	}
	now := time.Now()
	removeList := p.getRemoveList(ctx, indices, filter, now)
	if len(removeList) > 0 {
		err := p.deleteIndices(removeList)
		if err != nil {
			return err
		}
		p.loader.ReloadIndices()
	}
	return nil
}

func (p *provider) getRemoveList(ctx context.Context, indices *loader.IndexGroup, filter loader.Matcher, now time.Time) (list []string) {
	for _, entry := range indices.List {
		if filter(entry) && p.needToDelete(entry, now) {
			list = append(list, entry.Index)
		}
	}
	for _, ig := range indices.Groups {
		list = append(list, p.getRemoveList(ctx, ig, filter, now)...)
	}
	return nil
}

func (p *provider) needToDelete(entry *loader.IndexEntry, now time.Time) bool {
	if entry.MaxT.IsZero() || (entry.Num >= 0 && entry.Active) {
		return false
	}
	duration := p.retentions.GetTTL(entry)
	if duration <= 0 {
		return false
	}
	return now.After(entry.MaxT.Add(duration))
}

func (p *provider) deleteIndices(removeList []string) error {
	const size = 10 // delete too many at once and the request will be rejected
	for len(removeList) >= size {
		err := p.deleteIndex(removeList[:size])
		if err != nil {
			return err
		}
		removeList = removeList[size:]
	}
	if len(removeList) > 0 {
		err := p.deleteIndex(removeList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) deleteIndex(indices []string) error {
	if p.Cfg.PrintOnly {
		p.Log.Infof("clean indices %d, %v", len(indices), indices)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.RequestTimeout)
	defer cancel()
	resp, err := p.loader.Client().DeleteIndex(indices...).Do(ctx)
	if err != nil {
		if e, ok := err.(*elastic.Error); ok {
			if e.Status == 404 {
				return nil
			}
		}
		return err
	}
	if !resp.Acknowledged {
		return fmt.Errorf("delete indices Acknowledged=false")
	}
	p.Log.Infof("clean indices %d, %v", len(indices), indices)
	return nil
}
