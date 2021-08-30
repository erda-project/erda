package monitoring

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
			"monitor-monitoring",
			`
monitor-monitoring:
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
			"monitor-monitoring-service",
			`
monitor-monitoring:
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
