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

package models_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
)

func TestAIProxyProviders_ResetAesKey(t *testing.T) {
	var a = new(models.AIProxyProviders)
	a.ResetAesKey()
	t.Logf("aes key: %s, length: %d", a.AesKey, len(a.AesKey))
	if len(a.AesKey) != 16 {
		t.Error("a.AesKey's length should be 16")
	}
}

func TestAIProxyProviders_EncryptAPIKey(t *testing.T) {
	var a = new(models.AIProxyProviders)
	a.SetAPIKeyWithEncrypt("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")
	t.Logf("a.APIKey: %s, raw: %s", a.APIKey, a.GetAPIKeyWithDecrypt())
}

func BenchmarkAIProxyProviders_EncryptAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := make([]byte, 16)
		if _, err := rand.Read(key); err != nil {
			b.Fatal(err)
		}
		apiKey := "sk-" + hex.EncodeToString(key)
		var a = new(models.AIProxyProviders)
		a.SetAPIKeyWithEncrypt(apiKey)
		if a.GetAPIKeyWithDecrypt() != apiKey {
			b.Fatal("apiKey encrypt or decrypt error")
		}
	}
}
