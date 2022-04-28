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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata/pod"
)

func (p *provider) initCache(ctx context.Context) error {
	pSelector := p.Cfg.Pod.WatchSelector
	pList, err := p.Kubernetes.Client().CoreV1().Pods(pSelector.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: pSelector.LabelSelector,
		FieldSelector: pSelector.FieldSelector,
	})
	if err != nil {
		return fmt.Errorf("list pod err: %w", err)
	}

	cache, err := pod.NewCache(pList.Items, p.Cfg.Pod.AddMetadata.AnnotationInclude, p.Cfg.Pod.AddMetadata.LabelInclude)
	if err != nil {
		return fmt.Errorf("pod cache: %w", err)
	}
	p.podCache = cache

	return nil
}
