//go:build !windows
// +build !windows

package main

import (
	"embed"
	"log"
)

//go:embed testdata/v1:0.json
var testFS embed.FS

func main() {
	const path = "testdata/v1:0.json"

	data, err := testFS.ReadFile(path)
	if err != nil {
		log.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if len(data) == 0 {
		log.Fatalf("ReadFile(%q) returned empty content", path)
	}
}
