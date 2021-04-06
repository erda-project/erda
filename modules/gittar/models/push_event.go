package models

import "github.com/erda-project/erda/modules/gittar/pkg/gitmodule"

// PayloadPushEvent struct
// https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#push-events

type PayloadPushEvent struct {
	ObjectKind        string                `json:"object_kind"`
	IsTag             bool                  `json:"is_tag"`
	Ref               string                `json:"ref"`
	After             string                `json:"after"`
	Before            string                `json:"before"`
	IsDelete          bool                  `json:"is_delete"`
	Commits           []PayloadCommit       `json:"commits"`
	TotalCommitsCount int                   `json:"total_commits_count"`
	Pusher            *User                 `json:"pusher"`
	Repository        *gitmodule.Repository `json:"repository"`
}
