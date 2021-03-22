package apitestsv2

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/jsonpath"
)

// JudgeAsserts 执行断言测试
func (at *APITest) JudgeAsserts(outParams map[string]interface{}, asserts []apistructs.APIAssert) (bool, []*apistructs.APITestsAssertData) {
	var results []*apistructs.APITestsAssertData
	for _, assert := range asserts {
		// 出参里的值
		actualValue := outParams[assert.Arg]
		//var buffer bytes.Buffer
		//enc := json.NewEncoder(&buffer)
		//enc.SetEscapeHTML(false)
		//enc.Encode(actualValue)
		//actualValueString := buffer.String()
		//if len(actualValueString) > 0 {
		//	actualValueString = actualValueString[:len(actualValueString)-1]
		//}
		succ, err := doAssert(actualValue, assert.Operator, assert.Value)
		result := apistructs.APITestsAssertData{
			Arg:         assert.Arg,
			Operator:    assert.Operator,
			Value:       assert.Value,
			Success:     succ,
			ActualValue: actualValue,
			ErrorInfo: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}
		results = append(results, &result)
	}
	// 判断结果
	globalSuccess := true
	for _, result := range results {
		if !result.Success {
			globalSuccess = false
			break
		}
	}
	return globalSuccess, results
}

// Assert 断言操作匹配
func doAssert(value interface{}, op string, expect string) (bool, error) {
	switch op {
	case "=":
		return isEqual(value, expect), nil
	case "!=":
		return !isEqual(value, expect), nil
	case ">=":
		return moreThan(value, expect, true)
	case "<=":
		return moreThan(expect, value, true)
	case ">":
		return moreThan(value, expect, false)
	case "<":
		return moreThan(expect, value, false)
	case "contains":
		valMashal, err := json.Marshal(value)
		if err != nil {
			return false, errors.Errorf("failed to marshal, value:%+v, (%+v)", value, err)
		}
		isMatch, err := regexp.MatchString(expect, string(valMashal))
		if err != nil {
			return false, err
		}
		return isMatch, nil
	case "not_contains":
		valMashal, err := json.Marshal(value)
		if err != nil {
			return false, errors.Errorf("failed to marshal, value:%+v, (%+v)", value, err)
		}
		isMatch, err := regexp.MatchString(expect, string(valMashal))
		if err != nil {
			return false, err
		}
		return !isMatch, nil
	case "belong":
		return isBelong(value, expect)
	case "not_belong":
		ret, err := isBelong(value, expect)
		if err != nil {
			return false, err
		}
		return !ret, nil
	case "empty":
		return isEmpty(value)
	case "not_empty":
		ret, err := isEmpty(value)
		if err != nil {
			return false, err
		}
		return !ret, nil
	case "exist":
		return isExist(value, expect), nil
	case "not_exist":
		return !isExist(value, expect), nil
	default:
		return false, fmt.Errorf("invalid operator")
	}
}

func isEqual(value interface{}, expect string) bool {
	if fmt.Sprint(value) == expect {
		return true
	}

	valDigital, err := strconv.ParseFloat(fmt.Sprint(value), 64)
	if err != nil {
		return false
	}
	expectDigital, err := strconv.ParseFloat(fmt.Sprint(expect), 64)
	if err != nil {
		return false
	}
	return valDigital == expectDigital
}

func isBelong(value interface{}, expect string) (bool, error) {
	// 匹配区间, 跳过错误
	isBelong, isMatch, err := dealInterval(value, expect)
	if err != nil {
		return false, err
	}
	if isBelong {
		return true, nil
	}

	if isMatch {
		return false, nil
	}

	isBelong, err = dealDataSets(value, expect)
	if err != nil {
		return false, err
	}
	if isBelong {
		return true, nil
	}

	return false, nil
}

func dealDataSets(value interface{}, expect string) (bool, error) {
	var (
		p          = "^\\{.+\\}$" // [11,2)
		compareVal string
	)

	switch value.(type) {
	case string:
		compareVal = fmt.Sprint(value)
	default:
		valMashal, err := json.Marshal(value)
		if err != nil {
			return false, errors.Errorf("failed to marshal, value:%+v, (%+v)", value, err)
		}

		compareVal = string(valMashal)
	}

	matched, err := regexp.MatchString(p, expect)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("not support pattern, pattern:%s", expect)
	}

	expList := strings.Split(expect[1:len(expect)-1], ",")
	for i, length := 0, len(expList); i < length; i++ {
		exp := expList[i]
		if len(exp) > 1 {
			if exp[len(exp)-1] == ')' || exp[len(exp)-1] == ']' {
				return false, errors.Errorf("pattern error, value:%s", exp)
			}
			if exp[0] == '(' || exp[0] == '[' {
				if i == len(expList)-1 {
					return false, errors.Errorf("pattern error, value:%s", exp)
				}

				next := expList[i+1]
				if next[len(next)-1] != ')' && next[len(next)-1] != ']' {
					return false, errors.Errorf("pattern error, should like: %s,%s", exp, next)
				}

				interval := fmt.Sprint(exp, ",", next)

				// 跳过下次循环，因为已经取得集合右边的元素
				i++

				// 匹配区间, 跳过错误
				isBelong, _, _ := dealInterval(value, interval)
				if isBelong {
					return true, nil
				}
			}
		}

		if exp == compareVal {
			return true, nil
		}
	}

	return false, nil
}

func dealInterval(value interface{}, expect string) (bool, bool, error) {
	var (
		p  = "^\\-?\\d+$"                      // 数字
		p1 = "^\\[\\-?[0-9]+\\,\\-?[0-9]+\\]$" // [-11,-2]
		p2 = "^\\(\\-?[0-9]+\\,\\-?[0-9]+\\]$" // (11,2]
		p3 = "^\\[\\-?[0-9]+\\,\\-?[0-9]+\\)$" // [11,2)
		p4 = "^\\(\\-?[0-9]+\\,\\-?[0-9]+\\)$" // (11,2)

		isMatch bool
	)

	valMashal, err := json.Marshal(value)
	if err != nil {
		return false, false, errors.Errorf("failed to marshal, value:%+v, (%+v)", value, err)
	}

	matched, err := regexp.MatchString(p, string(valMashal))
	if !matched {
		return false, false, nil
	}

	digitalVal, err := strconv.Atoi(string(valMashal))
	if err != nil {
		return false, false, nil
	}

	// 处理全闭区间 [1,2]
	l, r, isMatch, err := getIntervalValue(p1, expect)
	if err != nil {
		return false, isMatch, err
	}
	if digitalVal >= l && digitalVal <= r {
		return true, isMatch, nil
	}

	if isMatch {
		return false, true, nil
	}

	// 处理左开右闭区间 (1,2]
	l, r, isMatch, err = getIntervalValue(p2, expect)
	if err != nil {
		return false, isMatch, err
	}
	if digitalVal > l && digitalVal <= r {
		return true, isMatch, nil
	}

	if isMatch {
		return false, true, nil
	}

	// 处理左闭右开区间 [1,2）
	l, r, isMatch, err = getIntervalValue(p3, expect)
	if err != nil {
		return false, isMatch, err
	}
	if digitalVal >= l && digitalVal < r {
		return true, isMatch, nil
	}

	if isMatch {
		return false, true, nil
	}

	// 处理全开区间 (1,2）
	l, r, isMatch, err = getIntervalValue(p4, expect)
	if err != nil {
		return false, isMatch, err
	}
	if digitalVal > l && digitalVal < r {
		return true, isMatch, nil
	}

	return false, isMatch, nil
}

func getIntervalValue(p, str string) (int, int, bool, error) {
	matched, err := regexp.MatchString(p, str)
	if err != nil {
		return 0, 0, false, err
	}
	if matched {
		re := regexp.MustCompile("\\-?\\d+")
		strList := re.FindAllString(str, 2)
		left, err := strconv.Atoi(strList[0])
		if err != nil {
			return 0, 0, true, err
		}

		right, err := strconv.Atoi(strList[1])
		if err != nil {
			return 0, 0, true, err
		}

		return left, right, true, nil
	}

	return 0, 0, false, nil
}

func isEmpty(value interface{}) (bool, error) {
	if value == nil {
		return true, nil
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		if reflect.ValueOf(value).Len() == 0 {
			return true, nil
		}
	case reflect.String:
		if fmt.Sprint(value) == "" {
			return true, nil
		}
	default:
		return false, errors.Errorf("not support this type, value:%v", reflect.ValueOf(value))
	}

	return false, nil
}

func isExist(value interface{}, expect string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	_, err := jsonpath.Get(value, expect)
	if err != nil {
		return false
	}
	return true
}

func moreThan(value interface{}, expect interface{}, isEqual bool) (bool, error) {
	valDigital, err := strconv.ParseFloat(fmt.Sprint(value), 64)
	if err != nil {
		return false, err
	}
	expectDigital, err := strconv.ParseFloat(fmt.Sprint(expect), 64)
	if err != nil {
		return false, err
	}

	if isEqual {
		return valDigital >= expectDigital, nil
	}

	return valDigital > expectDigital, nil
}
