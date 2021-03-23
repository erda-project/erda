package retry

import (
	"time"

	"github.com/hashicorp/go-multierror"
)

func Do(fn func() error, n int) (err error) {
	return DoWithInterval(fn, n, 0)
}

func DoWithInterval(fn func() error, n int, interval time.Duration) error {
	var me *multierror.Error
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		err := fn()
		if err == nil {
			me = nil
			break
		}
		me = multierror.Append(me, err)
		time.Sleep(interval)
	}
	if me != nil {
		return me.ErrorOrNil()
	}
	return nil
}
