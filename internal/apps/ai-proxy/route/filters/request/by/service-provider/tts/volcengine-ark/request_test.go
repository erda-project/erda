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

package volcengine_ark

import (
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/tts/ttsutil"
)

func TestMapVoice(t *testing.T) {
	tests := []struct {
		name     string
		input    openai.SpeechVoice
		expected string
	}{
		{
			name:     "Female Voice - Alloy",
			input:    openai.VoiceAlloy,
			expected: FixedFemaleVoice,
		},
		{
			name:     "Female Voice - Nova",
			input:    openai.VoiceNova,
			expected: FixedFemaleVoice,
		},
		{
			name:     "Female Voice - Shimmer",
			input:    openai.VoiceShimmer,
			expected: FixedFemaleVoice,
		},
		{
			name:     "Custom Female Voice - woman",
			input:    "woman",
			expected: FixedFemaleVoice,
		},
		{
			name:     "Custom Female Voice - female",
			input:    "female",
			expected: FixedFemaleVoice,
		},
		{
			name:     "Male Voice - Echo",
			input:    openai.VoiceEcho,
			expected: FixedMaleVoice,
		},
		{
			name:     "Male Voice - Fable",
			input:    openai.VoiceFable,
			expected: FixedMaleVoice,
		},
		{
			name:     "Male Voice - Onyx",
			input:    openai.VoiceOnyx,
			expected: FixedMaleVoice,
		},
		{
			name:     "Custom Male Voice - man",
			input:    "man",
			expected: FixedMaleVoice,
		},
		{
			name:     "Custom Male Voice - male",
			input:    "male",
			expected: FixedMaleVoice,
		},
		{
			name:     "Unknown Voice",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "Empty Voice",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ttsutil.MapVoice(tt.input, FixedMaleVoice, FixedFemaleVoice)
			assert.Equal(t, tt.expected, result)
		})
	}
}
