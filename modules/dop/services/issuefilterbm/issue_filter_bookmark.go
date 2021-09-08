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

package issuefilterbm

import (
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/strutil"
)

// IssueFilterBookmark service handler.
type IssueFilterBookmark struct {
	db *dao.DBClient
}

// Option of IssueFilterBookmark.
type Option func(*IssueFilterBookmark)

// New IssueFilterBookmark service handler.
func New(options ...Option) *IssueFilterBookmark {
	is := &IssueFilterBookmark{}
	for _, op := range options {
		op(is)
	}
	return is
}

// WithDBClient for IssueFilterBookmark.
func WithDBClient(db *dao.DBClient) Option {
	return func(is *IssueFilterBookmark) {
		is.db = db
	}
}

type MyFilterBmMap map[string][]MyFilterBm

type MyFilterBm struct {
	ID           string
	Name         string
	PageKey      string
	FilterEntity string
}

func (mp *MyFilterBmMap) GetByPageKey(pageKey string) []MyFilterBm {
	return (*mp)[pageKey]
}

func (s *IssueFilterBookmark) ListMyBms(userID, projectID string) (*MyFilterBmMap, error) {
	bms, err := s.db.ListIssueFilterBookmarkByUserIDAndProjectID(userID, projectID)
	if err != nil {
		return nil, err
	}
	mp := make(MyFilterBmMap)
	for _, bm := range bms {
		mp[bm.PageKey] = append(mp[bm.PageKey], MyFilterBm{
			ID:           bm.ID,
			Name:         bm.Name,
			PageKey:      bm.PageKey,
			FilterEntity: bm.FilterEntity,
		})
	}
	return &mp, nil
}

func (s *IssueFilterBookmark) GenPageKey(fixedIteration, fixedIssueType string) string {
	return strutil.Join([]string{fixedIteration, fixedIssueType}, "-", true)
}

func (s *IssueFilterBookmark) Delete(id string) error {
	return s.db.DeleteIssueFilterBookmarkByID(id)
}

func (s *IssueFilterBookmark) Create(bm *dao.IssueFilterBookmark) (string, error) {
	err := s.db.CreateIssueFilterBookmark(bm)
	if err != nil {
		return "", err
	}
	return bm.ID, nil
}
