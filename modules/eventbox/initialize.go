package eventbox

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/erda-project/erda/modules/eventbox/dispatcher"
)

func Initialize() error {
	dp, err := dispatcher.New()
	if err != nil {
		panic(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for range sig {
			dp.Stop()
			os.Exit(0)
		}
	}()

	dp.Start()
	return nil
}
