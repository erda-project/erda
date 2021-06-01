/*
Copyright [2018] [Spotahome Ltd.]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"errors"
	"fmt"
)

const (
	maxNameLength = 48
)

// Validate set the values by default if not defined and checks if the values given are valid
func (r *RedisFailover) Validate() error {
	if len(r.Name) > maxNameLength {
		return fmt.Errorf("name length can't be higher than %d", maxNameLength)
	}

	initialRedisCustomConfig := defaultRedisCustomConfig
	if r.Bootstrapping() {
		if r.Spec.BootstrapNode.Host == "" {
			return errors.New("BootstrapNode must include a host when provided")
		}

		if r.Spec.BootstrapNode.Port == "" {
			r.Spec.BootstrapNode.Port = defaultRedisPort
		}
		initialRedisCustomConfig = bootstrappingRedisCustomConfig
	}

	r.Spec.Redis.CustomConfig = append(initialRedisCustomConfig, r.Spec.Redis.CustomConfig...)

	if r.Spec.Redis.Image == "" {
		r.Spec.Redis.Image = defaultImage
	}

	if r.Spec.Sentinel.Image == "" {
		r.Spec.Sentinel.Image = defaultImage
	}

	if r.Spec.Redis.Replicas <= 0 {
		r.Spec.Redis.Replicas = defaultRedisNumber
	}

	if r.Spec.Sentinel.Replicas <= 0 {
		r.Spec.Sentinel.Replicas = defaultSentinelNumber
	}

	if r.Spec.Redis.Exporter.Image == "" {
		r.Spec.Redis.Exporter.Image = defaultExporterImage
	}

	if r.Spec.Sentinel.Exporter.Image == "" {
		r.Spec.Sentinel.Exporter.Image = defaultSentinelExporterImage
	}

	if len(r.Spec.Sentinel.CustomConfig) == 0 {
		r.Spec.Sentinel.CustomConfig = defaultSentinelCustomConfig
	}

	return nil
}
