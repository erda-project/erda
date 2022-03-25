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

package tagger

import (
	"regexp"
	"strings"

	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata/pod"
)

var matcherPattern = regexp.MustCompile("%{([^%{}]*)}")

func (p *provider) getPodMetadata(tags map[string]interface{}) (pod.Value, bool) {
	f := p.Cfg.Pod.AddMetadata.Finder
	switch f.Indexer {
	case pod.IndexerPodName:
		index := generateIndexByMatcher(f.Matcher, tags)
		return p.podCache.GetByPodNameIndexer(index)
	case pod.IndexerPodNameContainer:
		index := generateIndexByMatcher(f.Matcher, tags)
		return p.podCache.GetByPodNameContainerIndexer(index)
	}
	return pod.Value{}, false
}

// {namespace}/{pod}
func generateIndexByMatcher(matcher string, tags map[string]interface{}) pod.Key {
	matches := matcherPattern.FindAllStringSubmatch(matcher, -1)
	for _, item := range matches {
		if len(item) != 2 {
			continue
		}
		if v, ok := tags[item[1]]; ok {
			sv, ok := v.(string)
			if !ok {
				continue
			}
			matcher = strings.Replace(matcher, item[0], sv, -1)
		}
	}
	return pod.Key(matcher)
}
