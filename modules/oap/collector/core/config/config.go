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

import (
	"time"
)

// The global common config for pipeline component
type GlobalConfig struct {
	BatchLimit    int           `file:"batch_limit" default:"10"`
	FlushInterval time.Duration `file:"flush_interval" default:"1s"`
	FlushJitter   time.Duration `file:"flush_jitter" default:"1s"`
}

type Config struct {
	GlobalConfig GlobalConfig `file:"global_config"`
	Pipelines    []Pipeline   `file:"pipelines" desc:"compose of components"`
}

type Pipeline struct {
	Receivers  []string `file:"receivers"`
	Processors []string `file:"processors"`
	Exporters  []string `file:"exporters"`
}
