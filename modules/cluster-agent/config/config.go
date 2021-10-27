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

package config

type Config struct {
	Debug               bool   `default:"false" desc:"enable debug logging"`
	CollectClusterInfo  bool   `default:"true" desc:"enable collect cluster info"`
	ClusterDialEndpoint string `desc:"cluster dialer endpoint"`
	ClusterKey          string `desc:"cluster key"`
	ErdaNamespace       string `desc:"erda namespace"`
	ClusterAccessKey    string `desc:"cluster access key, if specified will doesn't start watcher"`
	K8SApiServerAddr    string `desc:"kube-apiserver address in cluster"`
}
