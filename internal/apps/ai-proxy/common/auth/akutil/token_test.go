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

package akutil

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
)

func TestIsTokenExpired(t *testing.T) {
	now := time.Now()

	t.Run("default behaviour", func(t *testing.T) {
		t.Setenv(EnvKeyEnableTokenExpireCheck, "")

		testCases := []struct {
			name  string
			token *clienttokenpb.ClientToken
			want  bool
		}{
			{
				name:  "nil token",
				token: nil,
				want:  false,
			},
			{
				name:  "no expiration",
				token: &clienttokenpb.ClientToken{},
				want:  false,
			},
			{
				name: "expired token",
				token: &clienttokenpb.ClientToken{
					ExpireAt: timestamppb.New(now.Add(-time.Minute)),
				},
				want: false,
			},
			{
				name: "future token",
				token: &clienttokenpb.ClientToken{
					ExpireAt: timestamppb.New(now.Add(time.Minute)),
				},
				want: false,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if got := isTokenExpired(tc.token); got != tc.want {
					t.Errorf("expected %v, got %v", tc.want, got)
				}
			})
		}
	})

	t.Run("disabled via env", func(t *testing.T) {
		t.Setenv(EnvKeyEnableTokenExpireCheck, "false")
		expiredToken := &clienttokenpb.ClientToken{
			ExpireAt: timestamppb.New(now.Add(-time.Minute)),
		}
		if got := isTokenExpired(expiredToken); got {
			t.Errorf("expected false when check disabled, got %v", got)
		}
	})

	t.Run("case insensitive true", func(t *testing.T) {
		t.Setenv(EnvKeyEnableTokenExpireCheck, "TRUE")
		expiredToken := &clienttokenpb.ClientToken{
			ExpireAt: timestamppb.New(now.Add(-time.Minute)),
		}
		if got := isTokenExpired(expiredToken); !got {
			t.Errorf("expected true when check explicitly enabled, got %v", got)
		}
	})
}
