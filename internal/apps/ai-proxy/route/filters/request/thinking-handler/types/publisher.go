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

package types

type ModelPublisher string

var (
	ModelPublisherOpenAI    ModelPublisher = "openai"
	ModelPublisherAnthropic ModelPublisher = "anthropic"
	ModelPublisherQwen      ModelPublisher = "qwen"
	ModelPublisherBytedance ModelPublisher = "bytedance" // Doubao
)

// Provider constants for cleaner and encoder
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderQwen      = "qwen"
	ProviderDoubao    = "doubao"
)

// API constants
const (
	APIChat      = "chat"
	APIResponses = "responses"
	APIRealtime  = "realtime"
)
