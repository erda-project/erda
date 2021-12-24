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
	"sync"

	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata"
	apiv1 "k8s.io/api/core/v1"
)

type Key struct {
	Name      string
	Namespace string
}

type Value map[string]string

type Cache struct {
	store map[Key]Value
	mu    sync.RWMutex
}

func NewCache(podList []apiv1.Pod) *Cache {
	store := make(map[Key]Value, len(podList))
	for _, pod := range podList {
		store[Key{Namespace: pod.Namespace, Name: pod.Name}] = extractPodMetadata(&pod)
	}
	return &Cache{
		store: store,
	}
}

func (c *Cache) AddOrUpdate(pod *apiv1.Pod) {
	c.mu.Lock()
	c.store[Key{Namespace: pod.Namespace, Name: pod.Name}] = extractPodMetadata(pod)
	c.mu.Unlock()
}

func (c *Cache) Delete(name, namespace string) {
	c.mu.Lock()
	delete(c.store, Key{Namespace: namespace, Name: name})
	c.mu.Unlock()
}

func (c *Cache) Get(name, namespace string) (map[string]string, bool) {
	c.mu.RLock()
	c.mu.RUnlock()
	val, ok := c.store[Key{Namespace: namespace, Name: name}]
	if !ok {
		return nil, false
	}

	return val, true
}

func extractPodMetadata(pod *apiv1.Pod) map[string]string {
	m := make(map[string]string, 10)
	m[metadata.PrefixPod+"name"] = pod.Name
	m[metadata.PrefixPod+"namespace"] = pod.Namespace
	m[metadata.PrefixPod+"uid"] = string(pod.UID)

	// labels
	for k, v := range pod.Labels {
		m[metadata.PrefixPodLabels+k] = v
	}
	// annotations
	for k, v := range pod.Annotations {
		m[metadata.PrefixPodAnnotations+k] = v
	}
	return m
}
