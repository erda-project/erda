// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iam

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/user/common"
)

const (
	EventAccountUpdate  = "ACCOUNT_UPDATE"
	EventAccountDestroy = "ACCOUNT_DESTROY"
	EventBindEmail      = "ACCOUNT_BIND_EMAIL"
	EventUnbindEmail    = "ACCOUNT_UNBIND_EMAIL"
	EventChangeEmail    = "ACCOUNT_CHANGE_EMAIL"
	EventChangeMobile   = "ACCOUNT_CHANGE_MOBILE"
	EventBindMobile     = "ACCOUNT_BIND_MOBILE"
	EventUnbindMobile   = "ACCOUNT_UNBIND_MOBILE"
)

func (p *provider) UserEventWebhook(_ context.Context, req *pb.UserEventWebhookRequest) (*pb.UserEventWebhookResponse, error) {
	if req.EventType != pb.EventType_EVENT_IAM {
		return nil, errors.New("event type must be iam")
	}

	event := req.GetData()
	if event == nil {
		return nil, errors.New("nil event payload")
	}

	l := p.Log.Set("event", event.EventName)

	jsonBytes, err := event.Data.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data to json")
	}

	l.Debugf("event payload, %s", string(jsonBytes))

	var userEvent UserEventDto
	if err := json.Unmarshal(jsonBytes, &userEvent); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload to IAMUserDto")
	}

	switch event.EventName {
	case EventUnbindMobile, EventUnbindEmail, EventAccountUpdate, EventChangeMobile,
		EventChangeEmail, EventBindMobile, EventBindEmail:
		// TODO: The IAMUserEventDto payload does not contain the most up-to-date user information.
		// For example, in the EventUnbindMobile case, the payload may include:
		// - email = null (no email data provided)
		//  - mobile = the mobile number that was previously bound (before unbinding)
		// So, the logic should be updated to fetch the latest user information first.
		newUser, err := p.getUser(cast.ToString(userEvent.ID), true)
		if err != nil {
			return nil, err
		}
		if err := p.doUpdateUser(l, newUser); err != nil {
			return nil, err
		}
	case EventAccountDestroy:
		if err := p.doDestroyUser(l, &userEvent); err != nil {
			return nil, err
		}
	default:
		l.Warnf("not support event type %s", event.EventName)
	}

	return &pb.UserEventWebhookResponse{}, nil
}

func (p *provider) doDestroyUser(l logs.Logger, userEvent *UserEventDto) error {
	l.Infof("user %s (id: %d) destroy", userEvent.Username, userEvent.ID)
	if err := p.bdl.DestroyUsers(apistructs.MemberDestroyRequest{
		UserIDs: []string{cast.ToString(userEvent.ID)},
	}); err != nil {
		return errors.Wrap(err, "failed to call bdl.DestroyUsers")
	}
	return nil
}

func (p *provider) doUpdateUser(l logs.Logger, newUser *common.User) error {
	l.Infof("user %s (id: %s) updated info", newUser.Name, newUser.ID)
	if err := p.bdl.UpdateMemberUserInfo(apistructs.MemberUserInfoUpdateRequest{
		Members: []apistructs.Member{
			{
				UserID: cast.ToString(newUser.ID),
				Email:  newUser.Email,
				Mobile: newUser.Phone,
				Name:   newUser.Name,
				Nick:   newUser.Nick,
				Avatar: newUser.AvatarURL,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "failed to call bdl.UpdateMemberUserInfo")
	}
	return nil
}
