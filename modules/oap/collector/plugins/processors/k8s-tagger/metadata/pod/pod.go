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

package pod

import (
	"regexp"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/erda-project/erda/modules/oap/collector/common"
	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata"
)

const (
	// "<pod_namespace>/<pod_name>"
	IndexerPodName = "pod_name"
	// "<pod_uid>"
	IndexerPodUID = "pod_uid"
)

type Key string

type Value map[string]string

type Cache struct {
	podnameIndexer    map[Key]Value
	poduidInddexer    map[Key]Value
	annotationInclude []*regexp.Regexp
	labelInclude      []*regexp.Regexp
	mu                sync.RWMutex
}

func PodName(namespace, name string) Key {
	return Key(strings.Join([]string{namespace, name}, "/"))
}

func PodUID(uid types.UID) Key {
	return Key(uid)
}

func NewCache(podList []apiv1.Pod, aInclude, lInclude []string) *Cache {
	c := &Cache{
		podnameIndexer:    make(map[Key]Value, len(podList)),
		poduidInddexer:    make(map[Key]Value, len(podList)),
		annotationInclude: make([]*regexp.Regexp, len(aInclude)),
		labelInclude:      make([]*regexp.Regexp, len(lInclude)),
	}

	for idx, item := range aInclude {
		c.annotationInclude[idx] = regexp.MustCompile(item)
	}
	for idx, item := range lInclude {
		c.labelInclude[idx] = regexp.MustCompile(item)
	}

	for _, pod := range podList {
		m := c.extractPodMetadata(&pod)
		c.podnameIndexer[PodName(pod.Namespace, pod.Name)] = m
		c.poduidInddexer[PodUID(pod.UID)] = m
	}
	return c
}

func (c *Cache) AddOrUpdate(pod *apiv1.Pod) {
	c.mu.Lock()
	m := c.extractPodMetadata(pod)
	c.podnameIndexer[PodName(pod.Namespace, pod.Name)] = m
	c.poduidInddexer[PodUID(pod.UID)] = m
	c.mu.Unlock()
}

func (c *Cache) Delete(pod *apiv1.Pod) {
	c.mu.Lock()
	delete(c.podnameIndexer, PodName(pod.Namespace, pod.Name))
	delete(c.poduidInddexer, PodUID(pod.UID))
	c.mu.Unlock()
}

func (c *Cache) GetByPodNameIndexer(index Key) (map[string]string, bool) {
	c.mu.RLock()
	c.mu.RUnlock()
	val, ok := c.podnameIndexer[index]
	if !ok {
		return nil, false
	}

	return val, true
}

func (c *Cache) GetByPodUIDIndexer(index Key) (map[string]string, bool) {
	c.mu.RLock()
	c.mu.RUnlock()
	val, ok := c.poduidInddexer[index]
	if !ok {
		return nil, false
	}

	return val, true
}

func (c *Cache) extractPodMetadata(pod *apiv1.Pod) map[string]string {
	m := make(map[string]string, 10)
	m[metadata.PrefixPod+"name"] = pod.Name
	m[metadata.PrefixPod+"namespace"] = pod.Namespace
	m[metadata.PrefixPod+"uid"] = string(pod.UID)
	m[metadata.PrefixPod+"ip"] = pod.Status.PodIP

	// labels
	for _, p := range c.labelInclude {
		for k, v := range pod.Labels {
			if p.Match([]byte(k)) {
				m[metadata.PrefixPodLabels+common.NormalizeKey(k)] = v
			}
		}
	}

	// annotations
	for _, p := range c.annotationInclude {
		for k, v := range pod.Annotations {
			if p.Match([]byte(k)) {
				m[metadata.PrefixPodAnnotations+common.NormalizeKey(k)] = v
			}
		}
	}

	return m
}
