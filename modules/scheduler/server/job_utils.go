package server

import (
	"regexp"
)

const jobNameNamespaceFormat = `^[a-zA-Z0-9_\-]+$`

var jobFormater *regexp.Regexp = regexp.MustCompile(jobNameNamespaceFormat)

func validateJobName(name string) bool {
	return jobFormater.MatchString(name)
}

func validateJobNamespace(namespace string) bool {
	return jobFormater.MatchString(namespace)
}

func validateJobFlowID(id string) bool {
	return jobFormater.MatchString(id)
}

func makeJobKey(namespce, name string) string {
	return "/dice/job/" + namespce + "/" + name
}

func makeJobFlowKey(namespace, id string) string {
	return "/dice/jobflow/" + namespace + "/" + id
}
