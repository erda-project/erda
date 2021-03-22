package pexpr

import (
	"regexp"

	"github.com/erda-project/erda/pkg/strutil"
)

// phRe 占位符正则表达式:
//   ${{ configs.key }}
//   ${{ dirs.preTaskName.fileName }}
//   ${{ outputs.preTaskName.key }}
//   ${{ params.key }}
//   ${{ (echo hello world) }}
var PhRe = regexp.MustCompile(`\${{[ ]{1}([^{}\s]+)[ ]{1}}}`) // [ ]{1} 强调前后均有且仅有一个空格

// loosePhRe 宽松的正则表达式:
var LoosePhRe = regexp.MustCompile(`\${{[^{}]+}}`)

// FindInvalidPlaceholders 找到表达式中不合规范的占位符
func FindInvalidPlaceholders(exprStr string) []string {
	// 合法的占位符列表
	validPhs := PhRe.FindAllString(exprStr, -1)
	// 宽松的占位符列表
	loosePhs := LoosePhRe.FindAllString(exprStr, -1)

	// 不在合法占位符列表中的即为非法占位符
	var invalidPhs []string
	for _, loose := range loosePhs {
		if !strutil.Exist(validPhs, loose) {
			invalidPhs = append(invalidPhs, loose)
		}
	}
	return invalidPhs
}
