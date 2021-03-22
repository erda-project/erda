package pipe

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintStdoutStderr(t *testing.T) {
	p := Line(
		ReadFile("/etc/hosts"),
		System("echo hello 1>&2"),
		Exec("ping", "-c", "3", "baidu.com"),
	)
	outb, errb, err := PrintStdoutStderr(p)
	require.NoError(t, err)
	fmt.Println("===========================")
	fmt.Println(len(outb))
	fmt.Println(len(errb))

	fmt.Println("===========================stdout")
	fmt.Println(string(outb))

	fmt.Println("===========================stderr")
	fmt.Println(string(errb))

	p = Line(
		ReadFile("/etc/hosts"),
		Exec("false"),
	)
	_, _, err = PrintStdoutStderr(p)
	require.Error(t, err)
}
