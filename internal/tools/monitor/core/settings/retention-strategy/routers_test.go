package retention

import (
	"fmt"
	"github.com/erda-project/erda/pkg/router"
	"testing"
)

func TestRouter(t *testing.T) {
	matcher := router.New()
	matcher.Add("application_*", []*router.KeyValue{
		{
			Key:   "application_key",
			Value: "application",
		},
	}, "application")
	matcher.Add("*", []*router.KeyValue{
		{
			Key:   "terminus_key",
			Value: "terminus",
		},
	}, "")

	find := matcher.Find("applicatdfsdfffion_sss*", map[string]string{
		"terminus_key": "terminus",
		//"terminus_value": "123",
	})

	fmt.Println(find)
}
