package expression

import (
	"fmt"
	"testing"
)

func Test_expression(t *testing.T) {

	var testData = []struct {
		express string
		result  string
	}{
		{
			"${{ '${{' == '}}' }}", " '${{' == '}}' ",
		},
		{
			"${{ '}}' == '}}' }}", " '}}' == '}}' ",
		},
		{
			"${{ '1' == '}}' }}", " '1' == '}}' ",
		},
		{
			"${{ '${{' == '2' }}", " '${{' == '2' ",
		},
		{
			"${{ '1' == '2' }}", " '1' == '2' ",
		},
		{
			"${{ '2' == '2' }}", " '2' == '2' ",
		},
		{
			"${{ '' == '2' }}", " '' == '2' ",
		},
		{
			"${{${{ == '2' }}}}", "${{ == '2' }}",
		},
	}

	for _, condition := range testData {
		if ReplacePlaceholder(condition.express) != condition.result {
			fmt.Println(" error express ", condition.express)
			t.Fail()
		}
	}
}
