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

package token

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda/pkg/secret"
)

func TestGenerateFromAkskPair(t *testing.T) {
	type args struct {
		pair *secret.AkSkPair
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				pair: &secret.AkSkPair{
					AccessKeyID: "IQ9E2Buhd8z2h7njPaxeGxq8",
					SecretKey:   "0O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6",
				},
			},
			want: "IQ9E2Buhd8z2h7njPaxeGxq80O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6",
		},
		{
			name: "invalid",
			args: args{
				pair: &secret.AkSkPair{
					AccessKeyID: "",
					SecretKey:   "0O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeFromAkskPair(tt.args.pair); got != tt.want {
				t.Errorf("EncodeFromAkskPair() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeWithAkskPair(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want *secret.AkSkPair
	}{
		{
			name: "",
			args: args{
				token: "IQ9E2Buhd8z2h7njPaxeGxq80O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6",
			},
			want: &secret.AkSkPair{
				AccessKeyID: "IQ9E2Buhd8z2h7njPaxeGxq8",
				SecretKey:   "0O2Hn0TrTrRwrds1q0un0p9AvX4JB8V6",
			},
		},
		{
			name: "invalid",
			args: args{
				token: "IQ9E2Buhd8z2h7njPaxeGxq80O2Hn0TrTrRwrds1q0un0p9AvX4JB8V",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodeToAkskPair(tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeToAkskPair() = %v, want %v", got, tt.want)
			}
		})
	}
}
