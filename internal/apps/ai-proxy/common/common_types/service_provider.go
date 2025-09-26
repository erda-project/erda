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

package common_types

// ServiceProviderType defines the service provider that hosts the model runtime.
type ServiceProviderType string

const (
	ServiceProviderTypeVolcengineArk  ServiceProviderType = "volcengine-ark"
	ServiceProviderTypeAliyunBailian  ServiceProviderType = "aliyun-bailian"
	ServiceProviderTypeAWSBedrock     ServiceProviderType = "aws-bedrock"
	ServiceProviderTypeAzureAIFoundry ServiceProviderType = "azure-ai-foundry"
	ServiceProviderTypeOpenAI         ServiceProviderType = "openai"
	ServiceProviderTypeAnthropic      ServiceProviderType = "anthropic"
)

func (m ServiceProviderType) String() string { return string(m) }
