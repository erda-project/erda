package server

import (
	"path/filepath"
	"regexp"
)

const (
	runtimeNameFormat = `^[a-zA-Z0-9\-]+$`
)

var (
	// record runtime's last restart time
	LastRestartTime = "lastRestartTime"
	//
	runtimeFormater *regexp.Regexp = regexp.MustCompile(runtimeNameFormat)
)

func makeRuntimeKey(namespace, name string) string {
	return filepath.Join("/dice/service/", namespace, name)
}

func validateRuntimeName(name string) bool {
	return len(name) > 0 && runtimeFormater.MatchString(name)
}

func validateRuntimeNamespace(namespace string) bool {
	return len(namespace) > 0 && runtimeFormater.MatchString(namespace)
}

func makePlatformKey(name string) string {
	return "/dice/service/platform/" + name
}
