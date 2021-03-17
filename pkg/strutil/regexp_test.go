package strutil

import (
	"regexp"
	"testing"

	"gotest.tools/assert"
)

func TestReplaceAllStringSubmatchFunc(t *testing.T) {
	s := "${java:OUTPUT:image} ${js:OUTPUT:image}"
	m := map[string]map[string]string{
		"java": {
			"image": "openjdk:8",
		},
		"js": {
			"image": "herd:1.3",
		},
	}
	re := regexp.MustCompile(`\${([^:]+):OUTPUT:([^:]+)}`)
	replaced := ReplaceAllStringSubmatchFunc(re, s, func(sub []string) string {
		return m[sub[1]][sub[2]]
	})
	assert.Equal(t, "openjdk:8 herd:1.3", replaced)
}
