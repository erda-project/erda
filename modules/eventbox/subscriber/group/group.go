package group

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/eventbox/conf"
	dispatchererror "github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/template"
)

type routerI interface {
	Route(m *types.Message) *dispatchererror.DispatchError
}

type GroupSubscriber struct {
	bundle *bundle.Bundle
	router routerI
}

type Option func(*GroupSubscriber)

func New(bundle *bundle.Bundle) *GroupSubscriber {
	subscriber := &GroupSubscriber{
		bundle: bundle,
	}
	return subscriber
}

func (d *GroupSubscriber) SetRoute(route routerI) {
	d.router = route
}

func (d *GroupSubscriber) Publish(dest string, content string, time int64, msg *types.Message) []error {
	errs := []error{}
	var groupID int64
	var groupNotifyContent apistructs.GroupNotifyContent
	err := json.Unmarshal([]byte(dest), &groupID)
	if err != nil {
		return []error{err}
	}
	err = json.Unmarshal([]byte(content), &groupNotifyContent)
	if err != nil {
		return []error{err}
	}
	groupDetail, err := d.bundle.GetNotifyGroupDetail(groupID, groupNotifyContent.OrgID, conf.BundleUserID())
	if err != nil {
		return []error{err}
	}

	createHistoryRequest := &apistructs.CreateNotifyHistoryRequest{
		NotifyName:            groupNotifyContent.NotifyName,
		NotifyItemDisplayName: groupNotifyContent.NotifyItemDisplayName,
		NotifySource: apistructs.NotifySource{
			Name:       groupNotifyContent.SourceName,
			SourceType: groupNotifyContent.SourceType,
			SourceID:   groupNotifyContent.SourceID,
		},
		NotifyTargets: groupDetail.Targets,
		Status:        "success",
		OrgID:         groupNotifyContent.OrgID,
		Label:         groupNotifyContent.Label,
		ClusterName:   groupNotifyContent.ClusterName,
	}

	for _, channel := range groupNotifyContent.Channels {
		chr := *createHistoryRequest
		chr.NotifySource.Params = channel.Params
		chr.Channel = channel.Name
		// 通知历史记录的label本来和通知项的label保持一致，这意味着每多1个租户，就会多n条通知项，即使每个租户属于同一级别
		// fdp有了用户可创建的项目后，为了让每个项目可以重复使用通知项，通知历史记录需要从别处获取租户信息，比如pipelineDetail里
		fdpNotifyLabel, ok := channel.Params["fdpNotifyLabel"]
		if ok {
			logrus.Debugf("fdpNotifyLabel is: %s", fdpNotifyLabel)
			chr.Label = fdpNotifyLabel
		}
		_, existTitle := channel.Params["title"]
		if !existTitle {
			channel.Params["title"] = groupNotifyContent.NotifyItemDisplayName
		}
		// 监控的sms，vms需要
		channel.Params["message"] = groupNotifyContent.NotifyItemDisplayName
		request := map[string]interface{}{
			"template":         channel.Template,
			"type":             channel.Type,
			"params":           channel.Params,
			"orgID":            groupNotifyContent.OrgID,
			"label":            groupNotifyContent.Label,
			"calledShowNumber": groupNotifyContent.CalledShowNumber,
		}

		if channel.Name == "email" {
			emails := []string{}
			for _, user := range groupDetail.Users {
				email := strings.TrimSpace(user.Email)
				if email != "" {
					emails = append(emails, email)
				}
			}
			msg := &types.Message{
				Content: request,
				Time:    time,
				Labels: map[types.LabelKey]interface{}{
					"EMAIL": emails,
				},
			}
			if len(emails) > 0 {
				d.routeMessage(msg, &chr)
			}
		} else if channel.Name == "sms" {
			mobiles := []string{}
			for _, user := range groupDetail.Users {
				mobile := strings.TrimSpace(user.Mobile)
				if mobile != "" {
					mobiles = append(mobiles, mobile)
				}
			}
			msg := &types.Message{
				Sender:  msg.Sender,
				Content: request,
				Time:    time,
				Labels: map[types.LabelKey]interface{}{
					"SMS": mobiles,
				},
			}
			if len(mobiles) > 0 {
				d.routeMessage(msg, &chr)
			}
		} else if channel.Name == "vms" {
			mobiles := []string{}
			for _, user := range groupDetail.Users {
				mobile := strings.TrimSpace(user.Mobile)
				if mobile != "" {
					mobiles = append(mobiles, mobile)
				}
			}
			msg := &types.Message{
				Sender:  msg.Sender,
				Content: request,
				Time:    time,
				Labels: map[types.LabelKey]interface{}{
					"VMS": mobiles,
				},
			}
			if len(mobiles) > 0 {
				d.routeMessage(msg, &chr)
			}
		} else if channel.Name == "dingding" {
			var atMobiles []string
			if k, ok := channel.Params["atUserIDs"]; ok {
				var users []apistructs.UserInfo
				userIDs := strutil.Split(k, ",", true)
				if len(userIDs) > 0 {
					if r, err := d.bundle.ListUsers(apistructs.UserListRequest{UserIDs: userIDs, Plaintext: true}); err != nil {
						logrus.Warnf("fail to fetch user, err: %v", err)
					} else {
						users = r.Users
					}
				}
				for _, u := range users {
					if u.Phone != "" {
						atMobiles = append(atMobiles, u.Phone)
					}
				}
			}
			var atMobilesTail string
			if len(atMobiles) > 0 {
				atMobilesTail = "\n\n"
			}
			for _, m := range atMobiles {
				atMobilesTail += "@" + m + " "
			}
			msg := &types.Message{
				Content: template.Render(channel.Template, channel.Params) + atMobilesTail,
				Time:    time,
				Labels: map[types.LabelKey]interface{}{
					"DINGDING": groupDetail.DingdingList,
					"MARKDOWN": map[string]string{
						"title": template.Render(channel.Params["title"], channel.Params),
					},
					"AT": map[string]interface{}{
						"atMobiles": atMobiles,
					},
				},
			}
			if len(groupDetail.DingdingList) > 0 {
				d.routeMessage(msg, &chr)
			}
		} else if channel.Name == "mbox" {
			userIDs := []string{}
			for _, user := range groupDetail.Users {
				userIDs = append(userIDs, user.ID)
			}
			msg := &types.Message{
				Content: request,
				Time:    time,
				Labels: map[types.LabelKey]interface{}{
					"MBOX": userIDs,
				},
			}
			if len(userIDs) > 0 {
				d.routeMessage(msg, &chr)
			}
		} else if channel.Name == "webhook" {
			msg := &types.Message{
				Content: map[string]string{
					"message": template.Render(channel.Template, channel.Params),
					"title":   template.Render(channel.Params["title"], channel.Params),
					"tag":     channel.Tag,
				},
				Time: time,
				Labels: map[types.LabelKey]interface{}{
					"HTTP": groupDetail.WebHookList,
				},
			}
			if len(groupDetail.WebHookList) > 0 {
				d.routeMessage(msg, &chr)
			}
		}
	}
	return errs
}

func (d *GroupSubscriber) Status() interface{} {
	return nil
}

func (d *GroupSubscriber) Name() string {
	return "GROUP"
}

func (d *GroupSubscriber) routeMessage(msg *types.Message, createHistoryRequest *apistructs.CreateNotifyHistoryRequest) {
	go func() {
		d.router.Route(msg)
		_, err := d.bundle.CreateNotifyHistory(createHistoryRequest)
		if err != nil {
			logrus.Errorf("创建通知历史记录失败: %v", err)
		}
	}()
}
