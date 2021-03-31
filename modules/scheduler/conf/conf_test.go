package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalEnv(t *testing.T) {
	os.Setenv("DEBUG", "true")
	os.Setenv("LISTEN_ADDR", "*:9091")

	debug := os.Getenv("DEBUG")
	listenAddr := os.Getenv("LISTEN_ADDR")
	assert.Equal(t, "true", debug)
	assert.Equal(t, "*:9091", listenAddr)
}
