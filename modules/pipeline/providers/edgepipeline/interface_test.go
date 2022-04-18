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

package edgepipeline

import (
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/bundle"
)

func TestSourceWhiteList(t *testing.T) {
	p := &provider{
		Cfg: &config{
			AllowedSources: []string{"cdp-", "recommend-"},
		},
	}
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "cdp source",
			src:  "cdp-123",
			want: true,
		},
		{
			name: "default source",
			src:  "default",
			want: false,
		},
		{
			name: "dice source",
			src:  "dice",
			want: false,
		},
		{
			name: "valid source with prefix",
			src:  "recommend-123",
			want: true,
		},
		{
			name: "invalid source with prefix",
			src:  "invalid-123",
			want: false,
		},
	}
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(p.bdl), "IsClusterDialerClientRegistered", func(_ *bundle.Bundle, _ string, _ string) (bool, error) {
		return true, nil
	})
	defer patch.Unpatch()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.ShouldDispatchToEdge(tt.src, "dev"); got != tt.want {
				t.Errorf("sourceWhiteList() = %v, want %v", got, tt.want)
			}
		})
	}
}
