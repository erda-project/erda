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

package protocol_translator_test

import (
	"testing"

	protocol_translator "github.com/erda-project/erda/internal/pkg/ai-proxy/filter/protocol-translator"
)

func TestParseProcessorNameArgs(t *testing.T) {
	var cases = []struct {
		Processor string
		Name      string
		Args      string
	}{
		{
			Processor: "SetAuthorizationIfNotSpecified",
			Name:      "SetAuthorizationIfNotSpecified",
		}, {
			Processor: "ReplaceAuthorizationWithAPIKey",
			Name:      "ReplaceAuthorizationWithAPIKey",
		}, {
			Processor: "SetAPIKeyIfNotSpecified",
			Name:      "SetAPIKeyIfNotSpecified",
		}, {
			Processor: `ReplaceURIPath("/openai/deployments/${ provider.metadata.DEVELOPMENT_NAME }/completions")`,
			Name:      "ReplaceURIPath",
			Args:      `"/openai/deployments/${ provider.metadata.DEVELOPMENT_NAME }/completions"`,
		},
	}
	for _, c := range cases {
		name, args, err := protocol_translator.ParseProcessorNameArgs(c.Processor)
		if err != nil {
			t.Fatalf("failed to ParseProcessorNameArgs, err: %v", err)
		}
		if name != c.Name {
			t.Fatalf("name error, expected: %s, got: %s", c.Name, name)
		}
		if args != c.Args {
			t.Fatalf("args error, expected: %s, got: %s", c.Args, args)
		}
		t.Logf("name: %s, args: %s", name, args)
	}
}
