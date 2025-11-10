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

package handler_client

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "google.golang.org/protobuf/types/known/structpb"

    clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
    metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
    "github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestDesensitizeClient(t *testing.T) {
    type want struct {
        accessKeyId string
        secretKeyId string
        secretIsNil bool
    }

    tests := []struct {
        name         string
        setupCtx     func(context.Context)
        buildPayload func() *clientpb.Client
        want         want
    }{
        {
            name:     "non-admin hides sensitive fields",
            setupCtx: nil,
            buildPayload: func() *clientpb.Client {
                return &clientpb.Client{
                    Id:          "client-1",
                    AccessKeyId: "AK-should-hide",
                    SecretKeyId: "SK-should-hide",
                    Metadata: &metadatapb.Metadata{
                        Secret: map[string]*structpb.Value{
                            "token": structpb.NewStringValue("value"),
                        },
                    },
                }
            },
            want: want{
                accessKeyId: "",
                secretKeyId: "",
                secretIsNil: true,
            },
        },
        {
            name: "admin keeps sensitive fields",
            setupCtx: func(ctx context.Context) {
                ctxhelper.PutIsAdmin(ctx, true)
            },
            buildPayload: func() *clientpb.Client {
                return &clientpb.Client{
                    Id:          "client-1",
                    AccessKeyId: "AK-keep",
                    SecretKeyId: "SK-keep",
                    Metadata: &metadatapb.Metadata{
                        Secret: map[string]*structpb.Value{
                            "token": structpb.NewStringValue("value"),
                        },
                    },
                }
            },
            want: want{
                accessKeyId: "AK-keep",
                secretKeyId: "SK-keep",
                secretIsNil: false,
            },
        },
        {
            name: "client querying itself keeps sensitive fields",
            setupCtx: func(ctx context.Context) {
                ctxhelper.PutClient(ctx, &clientpb.Client{Id: "client-1"})
            },
            buildPayload: func() *clientpb.Client {
                return &clientpb.Client{
                    Id:          "client-1",
                    AccessKeyId: "AK-keep",
                    SecretKeyId: "SK-keep",
                    Metadata: &metadatapb.Metadata{
                        Secret: map[string]*structpb.Value{
                            "token": structpb.NewStringValue("value"),
                        },
                    },
                }
            },
            want: want{
                accessKeyId: "AK-keep",
                secretKeyId: "SK-keep",
                secretIsNil: false,
            },
        },
        {
            name: "client querying others hides sensitive fields",
            setupCtx: func(ctx context.Context) {
                ctxhelper.PutClient(ctx, &clientpb.Client{Id: "client-1"})
            },
            buildPayload: func() *clientpb.Client {
                return &clientpb.Client{
                    Id:          "client-2",
                    AccessKeyId: "AK-should-hide",
                    SecretKeyId: "SK-should-hide",
                    Metadata: &metadatapb.Metadata{
                        Secret: map[string]*structpb.Value{
                            "token": structpb.NewStringValue("value"),
                        },
                    },
                }
            },
            want: want{
                accessKeyId: "",
                secretKeyId: "",
                secretIsNil: true,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
            if tt.setupCtx != nil {
                tt.setupCtx(ctx)
            }

            client := tt.buildPayload()
            desensitizeClient(ctx, client)

            assert.Equal(t, tt.want.accessKeyId, client.AccessKeyId)
            assert.Equal(t, tt.want.secretKeyId, client.SecretKeyId)
            if client.Metadata != nil {
                if tt.want.secretIsNil {
                    assert.Nil(t, client.Metadata.Secret)
                } else {
                    assert.NotNil(t, client.Metadata.Secret)
                }
            }
        })
    }
}

