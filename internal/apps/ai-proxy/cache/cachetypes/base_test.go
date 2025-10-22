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

package cachetypes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
)

func TestSmartCloneProtoMessage(t *testing.T) {
	original := &modelpb.Model{
		ApiKey: "secret",
		Metadata: &metadatapb.Metadata{
			Secret: map[string]*structpb.Value{
				"token": structpb.NewStringValue("value"),
			},
		},
	}

	cloned := smartClone(original).(*modelpb.Model)

	if original == cloned {
		t.Fatalf("expected different pointer after clone")
	}
	assert.Equal(t, original.ApiKey, cloned.ApiKey)

	cloned.ApiKey = "changed"
	assert.Equal(t, "secret", original.ApiKey)
}

func TestSmartCloneSliceOfProtoMessages(t *testing.T) {
	original := []*modelpb.Model{
		{
			ApiKey: "secret-1",
		},
		nil,
	}

	cloned := smartClone(original).([]*modelpb.Model)

	assert.Equal(t, len(original), len(cloned))
	assert.Nil(t, cloned[1])
	if original[0] == cloned[0] {
		t.Fatalf("expected slice element to be cloned")
	}

	cloned[0].ApiKey = "changed"
	assert.Equal(t, "secret-1", original[0].ApiKey)
}

func TestSmartCloneFallbackDeepCopy(t *testing.T) {
	type sample struct {
		Values map[string]int
	}

	original := sample{
		Values: map[string]int{"foo": 1},
	}

	cloned := smartClone(original).(sample)

	if fmt.Sprintf("%p", original.Values) == fmt.Sprintf("%p", cloned.Values) {
		t.Fatalf("expected map to be cloned")
	}
	cloned.Values["foo"] = 2
	assert.Equal(t, 1, original.Values["foo"])
}

func TestSmartCloneNilSlice(t *testing.T) {
	var original []*modelpb.Model

	cloned := smartClone(original)
	assert.Nil(t, cloned)
}
