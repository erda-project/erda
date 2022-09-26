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

package edgepipeline_register

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/pkg/transport"
)

func TestCheckAccessTokenFromCtx(t *testing.T) {
	ctx := transport.WithHeader(context.Background(), transport.Header{
		"Authorization": []string{"xxx"},
	})
	p := &provider{
		Cfg: &Config{
			IsEdge: true,
		},
	}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "CheckAccessToken", func(_ *provider, token string) error {
		return nil
	})
	defer pm1.Unpatch()
	err := p.CheckAccessTokenFromCtx(ctx)
	assert.NoError(t, err)
}
