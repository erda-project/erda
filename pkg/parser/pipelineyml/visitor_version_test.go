package pipelineyml

import (
	"testing"
)

func TestVersionVisitor_Visit(t *testing.T) {
	v := NewVersionVisitor()

	validVersionTestCases := []string{
		`version: 1.1`,
		`version: '1.1'`,
		`version: "1.1"`,
	}
	for _, tc := range validVersionTestCases {
		y, err := New([]byte(tc))
		if err != nil {
			t.Fatal(err)
		}
		y.s.Accept(v)
		if len(y.s.errs) > 0 {
			t.Fatal(y.s.mergeErrors())
		}
	}

	invalidVersionTestCases := []string{
		`version: 1.2`,
		`version: 2`,
		`version: 1.1.alpha`,
		`version: "1`,
		`version: test`,
	}
	for _, tc := range invalidVersionTestCases {
		_, err := New([]byte(tc))
		if err == nil {
			t.Fatalf("should error: `%s`", tc)
		}
	}
}

func TestGetVersion(t *testing.T) {
	validVersionTestCases := []string{
		`version: 1`,
		`version: '1'`,
		`version: "1"`,
		`version: 1.0`,
		`version: '1.0'`,
		`version: "1.0"`,
	}
	for _, tc := range validVersionTestCases {
		version, err := GetVersion([]byte(tc))
		if err != nil {
			t.Fatal(err)
		}
		if !(version == "1" || version == "1.0") {
			t.Fatalf("invalid version: %s", version)
		}
	}
}
