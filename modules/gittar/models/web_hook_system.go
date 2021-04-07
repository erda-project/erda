// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package models

func (svc *Service) GetSystemHooksByEvent(event string, active bool) ([]WebHook, error) {
	var hooks []WebHook
	err := svc.db.Where("hook_type=? AND "+event+"=1 AND is_active=?",
		HOOK_TYPE_SYSTEM, active).Find(&hooks).Error

	if err != nil {
		return nil, err
	}
	return hooks, nil
}

func (svc *Service) AddSystemHook(hook *WebHook) (*WebHook, error) {
	hook.HookType = HOOK_TYPE_SYSTEM
	//命名hook,检测同名,存在只做更新
	if hook.Name != "" {
		var currentHook WebHook
		err := svc.db.Where("hook_type=?  AND name=?",
			HOOK_TYPE_SYSTEM, hook.Name).First(&currentHook).Error
		if err != nil {
			err := svc.db.Create(hook).Error
			return hook, err
		} else {
			currentHook.Url = hook.Url
			currentHook.PushEvents = hook.PushEvents
			currentHook.IsActive = hook.IsActive
			err := svc.db.Save(&currentHook).Error
			return &currentHook, err
		}
	} else {
		//匿名hook直接添加
		err := svc.db.Create(hook).Error
		return hook, err
	}
}
