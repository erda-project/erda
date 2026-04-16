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

package status

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteSessionInfoFromMap(t *testing.T) {
	sessions := map[string]StatusInfo{
		"https://openapi.erda.cloud":  {Token: "Bearer token"},
		"https://openapi.example.com": {SessionID: "legacy"},
	}

	result := deleteSessionInfoFromMap(sessions, "https://openapi.erda.cloud")
	require.Len(t, result, 1)
	_, exists := result["https://openapi.erda.cloud"]
	require.False(t, exists)
	require.Equal(t, "legacy", result["https://openapi.example.com"].SessionID)
}
