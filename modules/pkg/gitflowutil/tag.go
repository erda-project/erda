package gitflowutil

import (
	"github.com/erda-project/erda/pkg/semver"
)

func IsReleaseTag(tag string) bool {
	return semver.Valid(tag)
}
