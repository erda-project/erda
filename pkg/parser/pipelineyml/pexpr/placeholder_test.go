package pexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindInvalidPlaceholders(t *testing.T) {
	invalidExprs := []string{
		"${{configs.key}}",    // 前后均缺失空格
		"${{ configs.key}}",   // 后面缺失空格
		"${{configs.key }}",   // 前面缺失空格
		"${{  configs.key }}", // 前面有两个空格
		"${{ configs.key  }}", // 后面有两个空格
	}

	for _, expr := range invalidExprs {
		fmt.Println(expr)
		v := FindInvalidPlaceholders(expr)
		assert.True(t, len(v) == 1)
		assert.True(t, v[0] == expr)
	}

	// 正确
	v := FindInvalidPlaceholders("${{ configs.key }}")
	assert.True(t, len(v) == 0)
}
