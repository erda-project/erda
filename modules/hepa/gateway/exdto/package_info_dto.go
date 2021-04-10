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

package exdto

type PackageInfoDto struct {
	Id       string `json:"id"`
	CreateAt string `json:"createAt"`
	PackageDto
}

type SortBySceneList []PackageInfoDto

func (list SortBySceneList) Len() int { return len(list) }

func (list SortBySceneList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list SortBySceneList) Less(i, j int) bool {
	if list[i].Scene == "unity" {
		return true
	}
	if list[j].Scene == "unity" {
		return false
	}
	return list[i].CreateAt > list[j].CreateAt
}
