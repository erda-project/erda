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

package notify

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/i18n"
)

func (o *NotifyGroup) GetNotifyItemsByNotifyID(notifyID int64) ([]*apistructs.NotifyItem, error) {
	return o.db.GetNotifyItemsByNotifyID(notifyID)
}

func (o *NotifyGroup) QueryNotifyItems(locale *i18n.LocaleResource, request *apistructs.QueryNotifyItemRequest) (*apistructs.QueryNotifyItemData, error) {
	itemData, err := o.db.QueryNotifyItems(request)
	if err != nil {
		return nil, err
	}
	for _, item := range itemData.List {
		o.LocaleItem(locale, item)
	}
	return itemData, nil
}

func (o *NotifyGroup) UpdateNotifyItem(request *apistructs.UpdateNotifyItemRequest) error {
	return o.db.UpdateNotifyItem(request)
}

func (o *NotifyGroup) LocaleItem(locale *i18n.LocaleResource, item *apistructs.NotifyItem) {
	item.DisplayName = locale.Get("notify." + item.Category + "." + item.Name)
	item.MarkdownTemplate = locale.Get("notify."+item.Category+"."+item.Name+".markdown_template", "")
	item.EmailTemplate = locale.Get("notify."+item.Category+"."+item.Name+".email", "")
	item.DingdingTemplate = locale.Get("notify."+item.Category+"."+item.Name+".dingding", "")
	item.MBoxTemplate = locale.Get("notify."+item.Category+"."+item.Name+".mbox", "")
}
