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

package elasticsearch

import (
	elasticsearchv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/elasticsearch/v1"
	corev1 "k8s.io/api/core/v1"
)

type ElasticsearchAndSecret struct {
	elasticsearchv1.Elasticsearch
	corev1.Secret
	corev1.ConfigMap
}
