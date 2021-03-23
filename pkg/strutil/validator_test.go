package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxLenValidator(t *testing.T) {
	sLen6 := "123456"
	err := Validate(sLen6, MaxLenValidator(1))
	assert.Error(t, err)
	err = Validate(sLen6, MaxLenValidator(10))
	assert.NoError(t, err)
}

func TestMinLenValidator(t *testing.T) {
	sLen0 := ""
	err := Validate(sLen0, MinLenValidator(1))
	fmt.Println(err)
	assert.Error(t, err)

	sLen1 := "1"
	err = Validate(sLen1, MinLenValidator(2))
	fmt.Println(err)
	assert.Error(t, err)
}

func TestMaxRuneCountValidator(t *testing.T) {
	s := "测试中文 _1"
	fmt.Println(len(s))
	err := Validate(s, MaxRuneCountValidator(7))
	assert.NoError(t, err)
	err = Validate(s, MaxRuneCountValidator(6))
	assert.Error(t, err)
}

func TestEnvKeyValidator(t *testing.T) {
	assert.NoError(t, Validate("a", EnvKeyValidator))
	assert.NoError(t, Validate("_", EnvKeyValidator))
	assert.NoError(t, Validate("_1", EnvKeyValidator))
	assert.NoError(t, Validate("__1", EnvKeyValidator))
	assert.Error(t, Validate("1", EnvKeyValidator))
	assert.Error(t, Validate("1a", EnvKeyValidator))
}

func TestNoChineseValidator(t *testing.T) {
	fmt.Println(Validate("是，hello 接", NoChineseValidator))
	assert.Error(t, Validate("hello，世界", NoChineseValidator))
	assert.NoError(t, Validate("hello，dice", NoChineseValidator))
}

func TestEnvValueLenValidator(t *testing.T) {
	assert.NoError(t, Validate("s", EnvValueLenValidator))
	b := make([]byte, 1024*128)
	assert.NoError(t, Validate(string(b), EnvValueLenValidator))
	b = append(b, byte(1))
	assert.Error(t, Validate(string(b), EnvValueLenValidator))
}

func TestAlphaNumericDashUnderscoreValidator(t *testing.T) {
	assert.NoError(t, Validate("s", AlphaNumericDashUnderscoreValidator))
	assert.NoError(t, Validate("s-_0", AlphaNumericDashUnderscoreValidator))
	assert.NoError(t, Validate("A0-_0", AlphaNumericDashUnderscoreValidator))
	assert.NoError(t, Validate("0s", AlphaNumericDashUnderscoreValidator))
	assert.NoError(t, Validate("0s.s", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("_s", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("-s", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("s ", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate(".s", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("s0-_Z/", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("s-", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("s_", AlphaNumericDashUnderscoreValidator))
	assert.Error(t, Validate("s.", AlphaNumericDashUnderscoreValidator))
}
