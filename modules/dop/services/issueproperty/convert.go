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

package issueproperty

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func (ip *IssueProperty) Convert(is *dao.IssueProperty) *apistructs.IssuePropertyIndex {
	return &apistructs.IssuePropertyIndex{
		PropertyID:        int64(is.ID),
		ScopeType:         is.ScopeType,
		ScopeID:           is.ScopeID,
		OrgID:             is.OrgID,
		Required:          is.Required,
		PropertyType:      is.PropertyType,
		PropertyName:      is.PropertyName,
		DisplayName:       is.DisplayName,
		PropertyIssueType: is.PropertyIssueType,
		Relation:          is.Relation,
		Index:             is.Index,
	}
}

func (ip *IssueProperty) BatchConvert(properties []dao.IssueProperty) []apistructs.IssuePropertyIndex {
	var response []apistructs.IssuePropertyIndex
	for _, is := range properties {
		response = append(response, apistructs.IssuePropertyIndex{
			PropertyID:        int64(is.ID),
			ScopeType:         is.ScopeType,
			ScopeID:           is.ScopeID,
			OrgID:             is.OrgID,
			Required:          is.Required,
			PropertyType:      is.PropertyType,
			PropertyName:      is.PropertyName,
			DisplayName:       is.DisplayName,
			PropertyIssueType: is.PropertyIssueType,
			Relation:          is.Relation,
			Index:             is.Index,
		})
	}
	return response
}

// []apistructs.IssuePropertyInstance => *apistructs.IssueAndPropertyAndValue
func (ip *IssueProperty) ConvertRelations(issueID int64, relations []apistructs.IssuePropertyInstance) (*apistructs.IssueAndPropertyAndValue, error) {
	res := apistructs.IssueAndPropertyAndValue{
		IssueID: issueID,
	}
	for i, v := range relations {
		var arbitraryValue interface{}
		// 判断出参应该是数字还是字符串
		if v.PropertyType.IsNumber() && v.ArbitraryValue != nil {
			// 空字符串无法转成数字
			if v.ArbitraryValue.(string) == "" {
				arbitraryValue = ""
			} else {
				// 数字类型
				str, err := strconv.ParseInt(v.ArbitraryValue.(string), 10, 64)
				if err != nil {
					return nil, err
				}
				arbitraryValue = str
			}
		} else {
			arbitraryValue = v.ArbitraryValue
		}

		res.Property = append(res.Property, apistructs.IssuePropertyExtraProperty{
			PropertyID:       v.PropertyID,
			PropertyType:     v.PropertyType,
			PropertyName:     v.PropertyName,
			Required:         v.Required,
			DisplayName:      v.DisplayName,
			ArbitraryValue:   arbitraryValue,
			EnumeratedValues: v.IssuePropertyIndex.EnumeratedValues,
		})
		for _, val := range v.EnumeratedValues {
			res.Property[i].Values = append(res.Property[i].Values, val.ID)
		}
	}
	return &res, nil
}
