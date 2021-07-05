// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package generalability

import (
	"fmt"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type testInterface interface {
	testFunc(arg interface{}) interface{}
}

func (p *provider) testFunc(arg interface{}) interface{} {
	return fmt.Sprintf("%s -> result", arg)
}

func Test_provider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		config   string
		arg      interface{}
		want     interface{}
	}{
		{
			"case 1",
			"erda.msp.resource.deploy.handlers.generalability",
			`
erda.msp.resource.deploy.handlers.generalability:
    message: "hello"
`,

			"test arg",
			"test arg -> result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			<-events.Started()

			p := hub.Provider(tt.provider).(*provider)
			if got := p.testFunc(tt.arg); got != tt.want {
				t.Errorf("provider.testFunc() = %v, want %v", got, tt.want)
			}
			if err := hub.Close(); err != nil {
				t.Errorf("Hub.Close() = %v, want nil", err)
			}
		})
	}
}

func Test_provider_service(t *testing.T) {
	tests := []struct {
		name    string
		service string
		config  string
		arg     interface{}
		want    interface{}
	}{
		{
			"case 1",
			"erda.msp.resource.deploy.handlers.generalability-service",
			`
erda.msp.resource.deploy.handlers.generalability:
    message: "hello"
`,

			"test arg",
			"test arg -> result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			<-events.Started()
			s := hub.Service(tt.service).(testInterface)
			if got := s.testFunc(tt.arg); got != tt.want {
				t.Errorf("(service %q).testFunc() = %v, want %v", tt.service, got, tt.want)
			}
			if err := hub.Close(); err != nil {
				t.Errorf("Hub.Close() = %v, want nil", err)
			}
		})
	}
}
