package semver

import (
	"fmt"
	"regexp"
)

// Semantic Version
// see: https://semver.org
// see: https://github.com/semver/semver/issues/232#issuecomment-405596809
var SemverRegexp = regexp.MustCompile(`^(v)?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func Valid(ver string) bool {
	return SemverRegexp.MatchString(ver)
}

func New(major int, vers ...int) string {
	minor := 0
	patch := 0
	if len(vers) > 0 {
		if len(vers) > 0 {
			minor = vers[0]
		}
		if len(vers) > 1 {
			patch = vers[1]
		}
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
