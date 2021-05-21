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

package testcase

import (
	"sort"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/i18n"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/xmind"
)

func (svc *Service) convert2XMind(tcs []apistructs.TestCaseWithSimpleSetInfo, locale string) (xmind.XMLContent, error) {
	l := svc.bdl.GetLocale(locale)
	// 定义临时结构体
	type TmpTestCases struct {
		TestSetID  uint64
		TestSetDir string
		TestCases  []apistructs.TestCase
	}
	// 按照父测试集 ID 进行分类
	tcGroupMapByTsID := make(map[uint64]*TmpTestCases)
	for _, tc := range tcs {
		tmp := tcGroupMapByTsID[tc.TestSetID]
		if tmp == nil {
			tmp = &TmpTestCases{}
		}
		tmp.TestSetID = tc.TestSetID
		tmp.TestCases = append(tmp.TestCases, tc.TestCase)
		tmp.TestSetDir = tc.Directory
		tcGroupMapByTsID[tc.TestSetID] = tmp
	}
	// 将 map 转换为 list，并用 testSetID 倒序排序
	tmpTcsList := make([]*TmpTestCases, 0, len(tcGroupMapByTsID))
	for _, tmp := range tcGroupMapByTsID {
		tmpTcsList = append(tmpTcsList, tmp)
	}
	sort.Slice(tmpTcsList, func(i, j int) bool {
		// 倒序排序
		return tmpTcsList[i].TestSetID < tmpTcsList[j].TestSetID
	})

	var content xmind.XMLContent
	rootTopic := &xmind.XMLTopic{Title: l.Get(i18n.I18nKeyTestCaseSheetName)}
	content.Sheet.Topic = rootTopic

	for _, tmpTcs := range tmpTcsList {
		// 插入目录节点
		currentTopic := rootTopic
		for _, dir := range strutil.Split(tmpTcs.TestSetDir, "/", true) {
			currentTopic = currentTopic.AddAttachedChildTopic(dir, true)
		}
		// 插入测试用例节点
		for _, tc := range tmpTcs.TestCases {
			insertTestCaseTopic(currentTopic, tc)
		}
	}

	return content, nil
}
