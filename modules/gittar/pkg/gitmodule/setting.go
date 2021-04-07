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

package gitmodule

var Setting = struct {
	MaxGitDiffLines          int
	MaxGitDiffLineCharacters int
	MaxGitDiffFiles          int
	MaxGitDiffSize           int
	ContextLineCount         int
	RepoStatsCache           Cache
	RepoSubmoduleCache       Cache
	PathCommitCache          Cache
	CommitCache              Cache
	ObjectSizeCache          Cache
}{
	MaxGitDiffLines:          200,
	MaxGitDiffLineCharacters: 1500, //diff line最大长度
	MaxGitDiffFiles:          100,
	MaxGitDiffSize:           5000, //大于指定大小就当二进制文件不做diff处理,单位Byte,单文件diff不做限制
	ContextLineCount:         3,
	RepoStatsCache:           NewMemCache(1000, "Repo-"),
	RepoSubmoduleCache:       NewMemCache(1000, "Repo-Submodule-"),
	PathCommitCache:          NewMemCache(5000, "path_commit_"),
	CommitCache:              NewMemCache(5000, "commit_"),
	ObjectSizeCache:          NewMemCache(10000, "object_size_"),
}
