package expression

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/Knetic/govaluate.v3"

	"github.com/erda-project/erda/modules/pipeline/pexpr"
)

// 匹配 ${{ xxx }}
var Re = pexpr.PhRe
var OldRe = regexp.MustCompile(`\${([^{}]+)}`)

const (
	Dirs    = "dirs"
	Outputs = "outputs"
	Random  = "random"
	Params  = "params"
	Globals = "globals"
	Configs = "configs"
)

const (
	TaskNotJumpOver  SignType = 1 // task 不跳过
	TaskJumpOver     SignType = 0 // task 跳过
	LeftPlaceholder           = "${{"
	RightPlaceholder          = "}}"

	OldLeftPlaceholder  = "${"
	OldRightPlaceholder = "}"
)

type SignType int

type ExpressionExecSign struct {
	Sign      SignType
	Msg       string
	Err       error
	Condition string
}

func Reconcile(condition string) ExpressionExecSign {

	// 表达式为空，不跳过
	if condition == "" {
		return ExpressionExecSign{
			Sign: TaskNotJumpOver,
		}
	}

	condition = ReplacePlaceholder(condition)
	condition = strings.Trim(condition, " ")

	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return ExpressionExecSign{
			Sign: TaskJumpOver,
			Err:  fmt.Errorf("new expression %s error %v, action skip", condition, err),
		}
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		return ExpressionExecSign{
			Sign: TaskJumpOver,
			Err:  fmt.Errorf("expression %s exec error %v, action skip", condition, err),
		}
	}

	done, _ := result.(bool)
	if !done {
		return ExpressionExecSign{
			Sign: TaskJumpOver,
			Msg:  fmt.Sprintf("run %s expression fail, action skip", condition),
		}
	}

	return ExpressionExecSign{
		Sign: TaskNotJumpOver,
	}
}

func ReplacePlaceholder(condition string) string {
	condition = strings.TrimPrefix(condition, LeftPlaceholder)
	condition = strings.TrimSuffix(condition, RightPlaceholder)
	return condition
}

func AppendPlaceholder(condition string) string {
	return fmt.Sprintf("%s %s %s", LeftPlaceholder, condition, RightPlaceholder)
}

// str 整段包含 ${{}} 的字符串
// placeholderParams eval 方法的参数
// matchType 匹配那些类型
// 该方法是将 str 进行正则获取全部匹配的表达式，然后根据匹配的表达式执行 eval 方法，最后将表达式替换成执行的值
func MatchEval(str string, placeholderParams map[string]string, matchType ...string) (string, error) {
	matchSlice := pexpr.LoosePhRe.FindAllString(str, -1)
	for _, v := range matchSlice {
		var find = false
		for _, match := range matchType {
			if strings.HasPrefix(v, LeftPlaceholder+" "+match) || strings.HasPrefix(v, OldLeftPlaceholder+match) {
				find = true
				break
			}
		}
		if !find {
			continue
		}

		result, err := pexpr.Eval(v, placeholderParams)
		if err != nil {
			return str, err
		}
		evalStr, err := EvalResultToString(result)
		if err != nil {
			return str, fmt.Errorf("format result %v to string error: %v", result, err)
		}
		str = strings.ReplaceAll(str, v, evalStr)
	}
	return str, nil
}

func EvalResultToString(result interface{}) (string, error) {
	switch result.(type) {
	case string:
		return result.(string), nil
	case int:
		return strconv.Itoa(result.(int)), nil
	case float64:
		return strconv.FormatFloat(result.(float64), 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(result.(float64), 'f', -1, 32), nil
	case bool:
		return strconv.FormatBool(result.(bool)), nil
	}

	jsonByte, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(jsonByte), nil
}

// GenConfigParams 生成全局参数的表达式
func GenConfigParams(key string) string {
	return LeftPlaceholder + " configs.autotest." + key + " " + RightPlaceholder
}
