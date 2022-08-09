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

package types

const (
	// ReleaseCallbackPath ReleaseCallback的路径
	ReleaseCallbackPath     = "/api/actions/release-callback"
	CDPCallbackPath         = "/api/actions/cdp-callback"
	GitCreateMrCallback     = "/api/actions/git-mr-create-callback"
	GitMergeMrCallback      = "/api/actions/git-mr-merge-callback"
	GitCloseMrCallback      = "/api/actions/git-mr-close-callback"
	GitCommentMrCallback    = "/api/actions/git-mr-comment-callback"
	GitDeleteBranchCallback = "/api/actions/git-branch-delete-callback"
	GitDeleteTagCallback    = "/api/actions/git-tag-delete-callback"
	IssueCallback           = "/api/actions/issue-callback"
	MrCheckRunCallback      = "/api/actions/check-run-callback"
	DevFlowCallback         = "/api/devflow/actions/callback"
)

type EventCallback struct {
	Name   string
	Path   string
	Events []string
}

var EventCallbacks = []EventCallback{
	{Name: "git_push_release", Path: ReleaseCallbackPath, Events: []string{"git_push"}},
	{Name: "cdp_pipeline", Path: CDPCallbackPath, Events: []string{"pipeline"}},
	{Name: "git_create_mr", Path: GitCreateMrCallback, Events: []string{"git_create_mr"}},
	{Name: "git_merge_mr", Path: GitMergeMrCallback, Events: []string{"git_merge_mr"}},
	{Name: "git_close_mr", Path: GitCloseMrCallback, Events: []string{"git_close_mr"}},
	{Name: "git_comment_mr", Path: GitCommentMrCallback, Events: []string{"git_comment_mr"}},
	{Name: "git_delete_branch", Path: GitDeleteBranchCallback, Events: []string{"git_delete_branch"}},
	{Name: "git_delete_tag", Path: GitDeleteTagCallback, Events: []string{"git_delete_tag"}},
	{Name: "issue", Path: IssueCallback, Events: []string{"issue"}},
	{Name: "check-run", Path: MrCheckRunCallback, Events: []string{"check-run"}},
	{Name: "qa_git_mr_create", Path: "/api/callbacks/git-mr-create", Events: []string{"git_create_mr"}},
	{Name: "dev_flow", Path: DevFlowCallback, Events: []string{"dev_flow"}},
}
