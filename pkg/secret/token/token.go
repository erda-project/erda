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
	"github.com/erda-project/erda/pkg/secret"
)

func EncodeFromAkskPair(pair *secret.AkSkPair) string {
	if pair == nil || pair.AccessKeyID == "" || pair.SecretKey == "" {
		return ""
	}
	return pair.AccessKeyID + pair.SecretKey
}

func DecodeToAkskPair(token string) *secret.AkSkPair {
	if len(token) != 56 {
		return nil
	}
	return &secret.AkSkPair{
		AccessKeyID: token[:24],
		SecretKey:   token[24:],
	}
}
