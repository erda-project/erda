package impl

import (
	"testing"
)

func Test_encodeTenantGroup(t *testing.T) {
	type args struct {
		projectId      string
		env            string
		clusterName    string
		tenantGroupKey string
	}
	tests := []struct {
		name string
		args args
	}{
		{"case1", args{projectId: "1", env: "DEV", clusterName: "test", tenantGroupKey: "test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeTenantGroup(tt.args.projectId, tt.args.env, tt.args.clusterName, tt.args.tenantGroupKey)
			if got == "" {
				t.Errorf("encodeTenantGroup() = %v", got)
			}
		})
	}
}
