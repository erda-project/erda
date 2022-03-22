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

package reconciler

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestMakePipelineGCKey(t *testing.T) {
	namespace := "pipeline-1"
	gcKey := makePipelineGCKey(namespace)
	assert.Equal(t, "/devops/pipeline/gc/reconciler/pipeline-1", gcKey)
}

func TestGetPipelineNamespaceFromGCWatchedKey(t *testing.T) {
	gckey := "/devops/pipeline/gc/reconciler/pipeline-1"
	namespace := getPipelineNamespaceFromGCWatchedKey(gckey)
	assert.Equal(t, "pipeline-1", namespace)
}

func TestMakePipelineGCKeyWithSlash(t *testing.T) {
	namespace := "pipeline-1"
	gcKey := makePipelineGCKeyWithSlash(namespace)
	assert.Equal(t, "/devops/pipeline/gc/reconciler/pipeline-1/", gcKey)
}

func TestMakePipelineGCSubKey(t *testing.T) {
	subKey := makePipelineGCSubKey("pipeline-1", 1)
	assert.Equal(t, "/devops/pipeline/gc/reconciler/pipeline-1/1", subKey)
}

func TestListenGC(t *testing.T) {
	r := &Reconciler{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(r), "ListenGC", func(r *Reconciler, ctx context.Context) {
		return
	})
	defer pm1.Unpatch()
	t.Run("listenGC", func(t *testing.T) {
		r.ListenGC(context.Background())
	})
}
