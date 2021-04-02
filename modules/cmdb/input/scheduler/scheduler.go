package scheduler

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/event"
)

var instanceEventName = "instances-status"
var cmdbWebhookName = "cmdb_hook_scheduler_instance"

// NewWebhookForInstanceEvents 创建 webhook 用于获取 scheduler 实例变更事件
func NewWebhookForInstanceEvents(addr string, url string) error {
	if !isValidURL(url) {
		return errors.Errorf("invalid url: %s", url)
	}

	s, err := event.NewWebhook(addr)
	if err != nil {
		return err
	}

	hooks, err := s.List()
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if hook.Name == cmdbWebhookName && strings.Contains(hook.URL, url) && strings.Compare(hook.Events[0], instanceEventName) == 0 {
			logrus.Infof("eventbox webhook already exists, name: %s, url: %s", cmdbWebhookName, hook.URL)
			return nil
		}
	}

	spec := apistructs.WebhookCreateRequest{
		Name:   cmdbWebhookName,
		Events: []string{instanceEventName},
		URL:    url,
		Active: true,
		HookLocation: apistructs.HookLocation{
			Org:         "-1",
			Project:     "-1",
			Application: "-1",
		},
	}

	return s.Create(spec)
}

func isValidURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}

	return true
}
