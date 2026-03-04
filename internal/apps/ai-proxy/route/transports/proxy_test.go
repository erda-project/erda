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

package transports

import (
	"context"
	"testing"
	"time"
)

func TestWithForwardDialTimeout(t *testing.T) {
	ctx := context.Background()
	next := WithForwardDialTimeout(ctx, 5*time.Second)
	got, ok := getForwardDialTimeoutFromContext(next)
	if !ok {
		t.Fatal("expected forward dial timeout in context")
	}
	if got != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %s", got)
	}
}

func TestWithForwardDialTimeoutInvalid(t *testing.T) {
	ctx := context.Background()
	next := WithForwardDialTimeout(ctx, 0)
	if _, ok := getForwardDialTimeoutFromContext(next); ok {
		t.Fatal("expected no timeout in context when input is invalid")
	}
}

func TestWithForwardTLSHandshakeTimeout(t *testing.T) {
	ctx := context.Background()
	next := WithForwardTLSHandshakeTimeout(ctx, 5*time.Second)
	got, ok := getForwardTLSHandshakeTimeoutFromContext(next)
	if !ok {
		t.Fatal("expected forward tls handshake timeout in context")
	}
	if got != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %s", got)
	}
}

func TestWithForwardTLSHandshakeTimeoutInvalid(t *testing.T) {
	ctx := context.Background()
	next := WithForwardTLSHandshakeTimeout(ctx, 0)
	if _, ok := getForwardTLSHandshakeTimeoutFromContext(next); ok {
		t.Fatal("expected no tls handshake timeout in context when input is invalid")
	}
}
