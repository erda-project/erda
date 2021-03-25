package crondsvc

import (
	"fmt"
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/cron"
)

// Result:
// 1000  个约 0.03s
// 10000 个约 1.7s
func TestReloadSpeed(t *testing.T) {
	d := cron.New()
	d.Start()
	for i := 0; i < 10; i++ {
		if err := d.AddFunc("*/1 * * * *", func() {
			fmt.Println("hello world")
		}); err != nil {
			panic(err)
		}
	}
	time.Sleep(time.Second*2)
}
