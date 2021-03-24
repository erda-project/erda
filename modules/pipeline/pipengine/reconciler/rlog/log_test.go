package rlog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPErrorAndReturn(t *testing.T) {
	err := fmt.Errorf("failed to get schedulable tasks, err: %s", "internal error")
	err = PErrorAndReturn(1, err)
	assert.Error(t, err)
	assert.True(t, true, strings.HasPrefix(err.Error(), pErrorFormat))
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	s := "start watching"
	Infof(s)
	assert.True(t, strings.Contains(buf.String(), fmt.Sprintf(errorFormat, s)))
}
