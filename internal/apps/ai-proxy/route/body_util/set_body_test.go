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

package body_util

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSetBodyWithProtoMessage(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	require.NoError(t, err)

	msg := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"name": structpb.NewStringValue("hello"),
		},
	}

	err = SetBody(req, msg)
	require.NoError(t, err)

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"hello"}`, string(bodyBytes))
	require.Equal(t, int64(len(bodyBytes)), req.ContentLength)
	require.Equal(t, "application/json", req.Header.Get("Content-Type"))
	require.Equal(t, fmt.Sprintf("%d", len(bodyBytes)), req.Header.Get("Content-Length"))
}
