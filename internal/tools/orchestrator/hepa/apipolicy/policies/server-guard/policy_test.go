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

package serverguard

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/apipolicy"
)

func Test_setMSEIngressAnnotation(t *testing.T) {
	type args struct {
		policyDto          *PolicyDto
		ingressAnnotations *apipolicy.IngressAnnotation
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			args: args{
				policyDto: &PolicyDto{
					BaseDto:        apipolicy.BaseDto{},
					MaxTps:         0,
					Busrt:          2,
					ExtraLatency:   0,
					RefuseCode:     0,
					RefuseResponse: "",
				},
				ingressAnnotations: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setMSEIngressAnnotation(tt.args.policyDto, tt.args.ingressAnnotations)
		})
	}
}
