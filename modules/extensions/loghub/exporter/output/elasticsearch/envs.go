package elasticsearch

import (
	"os"
	"strings"
)

func setEnvDefault(key, val string) {
	if len(val) > 0 {
		_, ok := os.LookupEnv(key)
		if ok {
			return
		}
		os.Setenv(key, val)
	}
}

func init() {
	env := os.Getenv("MONITOR_LOG_OUTPUT_CONFIG")
	if len(env) > 0 {
		lines := strings.Split(env, " ")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) <= 0 {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 1 {
				if strings.HasPrefix(kv[0], "http://") || strings.HasPrefix(kv[0], "https://") {
					setEnvDefault("ES_URLS", kv[0])
				}
			} else if len(kv) == 2 {
				setEnvDefault(strings.ToUpper(kv[0]), kv[1])

			}
		}
	}
}
