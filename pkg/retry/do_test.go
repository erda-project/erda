package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func testF(text string) error {
	return errors.New(text)
}

func testFF() {
	fmt.Println(time.Now())
}

func TestDoWithInterval(t *testing.T) {
	err := DoWithInterval(func() error {
		testFF()
		return nil
	}, 2, time.Second*3)
	require.Error(t, err)
}

func TestDo(t *testing.T) {
	var i = 0
	err := DoWithInterval(func() error {
		if i == 0 {
			i++
			return fmt.Errorf("1")
		}
		i++
		return nil
	}, 1, 1*time.Second)
	spew.Dump(err)
}
