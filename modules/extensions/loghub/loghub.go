package loghub

import (
	"os"
)

// Init .
func Init() {
	setEnvDefault("LOGS_ES_URL", "ES_URL")
	setEnvDefault("LOGS_ES_SECURITY_ENABLE", "ES_SECURITY_ENABLE")
	setEnvDefault("LOGS_ES_SECURITY_USERNAME", "ES_SECURITY_USERNAME")
	setEnvDefault("LOGS_ES_SECURITY_PASSWORD", "ES_SECURITY_PASSWORD")

	_, ok := os.LookupEnv("CONF_NAME")
	if !ok {
		val, ok := os.LookupEnv("MONITOR_LOG_OUTPUT")
		if ok {
			os.Setenv("CONF_NAME", "output/"+val)
		}
	}
}

func setEnvDefault(key string, from ...string) {
	_, ok := os.LookupEnv(key)
	if ok {
		return
	}
	for _, item := range from {
		val, ok := os.LookupEnv(item)
		if ok {
			os.Setenv(key, val)
			break
		}
	}
}
