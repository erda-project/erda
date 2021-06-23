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

package autoscanner

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/ops/dbclient"
)

const NoticeContent = "「SYSTEM INFO」: Your current company has been using the ERDA platform for more than 1 year!!"

type AutoScanner struct {
	bdl *bundle.Bundle
	db  *dbclient.DBClient
}

type Option func(as *AutoScanner)

func New(db *dbclient.DBClient, bdl *bundle.Bundle) *AutoScanner {
	return &AutoScanner{
		bdl: bdl,
		db:  db,
	}
}

func (as *AutoScanner) Run() {
	for {
		logrus.Info("start checking the notice exists")
		as.CheckNoticeExist()

		logrus.Info("start checking the notices expired")
		as.CheckNoticeExpired()
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("start checking")
		}
	}
}

func (as *AutoScanner) CheckNoticeExist() {
	// list all org
	orgInfo, err := as.bdl.ListDopOrgs(&apistructs.OrgSearchRequest{})
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, org := range orgInfo.List {
		// get notice list
		noticeList, err := as.bdl.ListNoticeByOrgID(org.ID)
		if err != nil {
			logrus.Error(err)
			return
		}

		// find notice exist
		var noticeExist bool
		for _, notice := range noticeList.Data.List {
			if notice.Content == NoticeContent {
				noticeExist = true
				break
			}
		}

		if !noticeExist {
			as.PublishNotice(org.ID)
		}
	}
}

func (as *AutoScanner) CheckNoticeExpired() {
	// list all org
	orgInfo, err := as.bdl.ListDopOrgs(&apistructs.OrgSearchRequest{})
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, org := range orgInfo.List {
		// get notice list
		noticeList, err := as.bdl.ListNoticeByOrgID(org.ID)
		if err != nil {
			logrus.Error(err)
			return
		}
		currentTime := time.Now()
		for _, notice := range noticeList.Data.List {
			if notice.Content == NoticeContent {
				timeSub := currentTime.Sub(notice.CreatedAt.AddDate(0, 0, 7))
				if timeSub > 0 && notice.Status == apistructs.NoticePublished {
					err := as.UnPublishNotice(org.ID, notice.ID)
					if err != nil {
						logrus.Error(err)
					}
					break
				}
			}
		}
	}

}

func (as *AutoScanner) UnPublishNotice(orgID, noticeID uint64) error {
	err := as.bdl.PublishORUnPublishNotice(orgID, noticeID, "unpublish")
	if err != nil {
		return err
	}
	return nil
}

func (as *AutoScanner) PublishNotice(orgID uint64) {
	// list all cluster from the org
	clusters, err := as.bdl.ListClusters("", orgID)
	if err != nil {
		logrus.Error(err)
		return
	}

	currentTime := time.Now()
	// find cluster if expired 1 year
	for _, cluster := range clusters {
		timeSub := currentTime.Sub(cluster.CreatedAt.AddDate(1, 0, 0))
		if timeSub > 0 {
			nr := &apistructs.NoticeCreateRequest{
				Content: NoticeContent,
				IdentityInfo: apistructs.IdentityInfo{
					UserID:         "1",
					InternalClient: "",
				},
			}

			// create notice
			resp, err := as.bdl.CreateNoticeRequest(nr, orgID)
			if err != nil {
				logrus.Error(err)
				break
			}

			// publish notice
			err = as.bdl.PublishORUnPublishNotice(orgID, resp.Data.ID, "publish")
			if err != nil {
				logrus.Error(err)
			}
			break
		}
	}
}
