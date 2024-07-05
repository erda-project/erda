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

package sdk

type (
	DsRespStreamChunk struct {
		Output    DsRespStreamChunkOutput `json:"output,omitempty"`
		RequestID string                  `json:"request_id,omitempty"`
		Usage     DsRespStreamUsage       `json:"usage,omitempty"`
	}
	DsRespStreamChunkOutput struct {
		Choices []DsRespStreamChunkOutputChoice `json:"choices,omitempty"`
		Text    string                          `json:"text,omitempty"`
	}
	DsRespStreamChunkOutputChoice struct {
		Message      DsRespStreamChunkOutputChoiceMessage `json:"message,omitempty"`
		FinishReason string                               `json:"finish_reason,omitempty"`
	}
	DsRespStreamChunkOutputChoiceMessage struct {
		Content any    `json:"content,omitempty"` // string or [] DsRespStreamChunkOutputChoiceMessagePart
		Role    string `json:"role,omitempty"`
	}
	DsRespStreamChunkOutputChoiceMessagePart struct {
		Text string `json:"text,omitempty"`
	}
	DsRespStreamUsage struct {
		TotalTokens  uint64 `json:"total_tokens"`
		InputTokens  uint64 `json:"input_tokens"`
		OutputTokens uint64 `json:"output_tokens"`
	}
)
