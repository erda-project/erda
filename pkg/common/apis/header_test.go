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

package apis

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func Test_GetInternalClient(t *testing.T) {
	type args struct {
		ctx func() context.Context
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1: get internal client from incoming context",
			args: args{
				ctx: func() context.Context {
					return metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
						headerInternalClient: "fake-service",
					}))
				},
			},
			want: "fake-service",
		},
		{
			name: "case2: get non-existing internal client from context",
			args: args{
				ctx: func() context.Context {
					return context.Background()
				},
			},
			want: "",
		},
		{
			name: "case3: get internal client from outgoing context",
			args: args{
				ctx: func() context.Context {
					ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
						headerInternalClient: "fake-service-from-outgoing",
					}))
					return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
						"x-portal-dest": "fake-host",
					}))
				},
			},
			want: "fake-service-from-outgoing",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := GetInternalClient(test.args.ctx()); got != test.want {
				t.Errorf("GetInternalClient() = %v, want %v", got, test.want)
			}
		})
	}
}
