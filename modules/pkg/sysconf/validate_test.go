package sysconf

import (
	"testing"
)

func TestIsClusterName(t *testing.T) {
	m := map[string]bool{
		"":              false,
		"terminus":      false,
		"terminus-test": true,
		"terminus-prod": true,
	}
	for k, v := range m {
		if v != isClusterName(k) {
			t.Fatal(k)
		}
	}
}

func TestIsPort(t *testing.T) {
	m := map[int]bool{
		0:     false,
		-1:    false,
		22:    true,
		65535: true,
		65536: false,
	}
	for k, v := range m {
		if v != IsPort(k) {
			t.Fatal(k)
		}
	}
}

func TestIsDNSName(t *testing.T) {
	m := map[string]bool{
		"":            false,
		"*.aa":        false,
		"a.b":         true,
		"192.168.0.1": true,
		"test.com":    true,
	}
	for k, v := range m {
		if v != IsDNSName(k) {
			t.Fatal(k)
		}
	}
}
