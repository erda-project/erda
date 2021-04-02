package marathon

import (
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/httpclient"
)

func TestMarathon_SuspendApp(t *testing.T) {
	m := &Marathon{
		name:   "MARATHONFORTERMINUSTEST",
		addr:   "dcos.test.terminus.io/service/marathon",
		client: httpclient.New().BasicAuth("admin", "Terminus1234"),
	}

	ch := make(chan string, 10)
	for i := 0; i < 1; i++ {
		ch <- "/runtimes/v1/addon-redis/d140959eb05b4960846e7b0fb3cc42a6/redis-master-d140959eb05b4960846e7b0fb3cc42a6"
		time.Sleep(1 * time.Second)
	}
	go m.SuspendApp(ch)

	time.Sleep(1 * time.Second)
	close(ch)
}
