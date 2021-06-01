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

const (
	defaultRedisNumber           = 3
	defaultSentinelNumber        = 3
	defaultSentinelExporterImage = "leominov/redis_sentinel_exporter:1.3.0"
	defaultExporterImage         = "oliver006/redis_exporter:v1.3.5-alpine"
	defaultImage                 = "redis:5.0-alpine"
	defaultRedisPort             = "6379"
)

var (
	defaultSentinelCustomConfig = []string{
		"down-after-milliseconds 5000",
		"failover-timeout 10000",
	}
	defaultRedisCustomConfig = []string{
		"replica-priority 100",
	}
	bootstrappingRedisCustomConfig = []string{
		"replica-priority 0",
	}
)
