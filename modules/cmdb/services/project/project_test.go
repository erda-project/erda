package project

import "testing"

func TestClass_genProjectNamespace(t *testing.T) {
	namespaces := genProjectNamespace("1")
	expected := map[string]string{"DEV": "project-1-dev", "TEST": "project-1-test", "STAGING": "project-1-staging", "PROD": "project-1-prod"}
	for env, expectedNS := range expected {
		actuallyNS, ok := namespaces[env]
		if !ok {
			t.Errorf("env not existd: %s", env)
		}
		if expectedNS != actuallyNS {
			t.Errorf("expected: [%s:%s], actually: [%s:%s]", env, expectedNS, env, actuallyNS)
		}
	}
}
