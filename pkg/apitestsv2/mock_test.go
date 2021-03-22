package apitestsv2

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	s := randString(Integer)
	i, err := strconv.Atoi(s)
	assert.NoError(t, err)
	fmt.Println(s, i)

	s = randString(String)
	fmt.Println(s)
}
