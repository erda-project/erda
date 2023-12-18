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

package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/apistructs"
	aiproxyclient "github.com/erda-project/erda/internal/apps/ai-proxy/sdk/client"
	"github.com/erda-project/erda/internal/tools/gittar/ai/cr/util/mrutil"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/gitmodule"
	"github.com/erda-project/erda/internal/tools/gittar/pkg/util/guid"
	"github.com/erda-project/erda/internal/tools/gittar/uc"
)

type NoteRole string

var (
	NoteRoleAI   = NoteRole("AI")
	NoteRoleUser = NoteRole("USER")
)

type NoteRequest struct {
	Note         string `json:"note"`
	Type         string `json:"type"`
	DiscussionId string `json:"discussionId"` //讨论id
	OldCommitId  string `json:"oldCommitId"`
	NewCommitId  string `json:"newCommitId"`
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldLine      int    `json:"oldLine"`
	NewLine      int    `json:"newLine"`
	OldLineTo    int    `json:"oldLineTo"`
	NewLineTo    int    `json:"newLineTo"` // these four index number must in a same section
	Score        int    `json:"score" default:"-1"`

	Role             NoteRole         `json:"role"`
	AICodeReviewType AICodeReviewType `json:"aiCodeReviewType,omitempty"`
	StartAISession   bool             `json:"startAISession,omitempty"`
}

const MAX_NOTE_DIFF_LINE_COUNT = 15

var (
	NoteTypeNormal        = "normal"
	NoteTypeDiffNote      = "diff_note"
	NoteTypeDiffNoteReply = "diff_note_reply"
)

// Note struct
type Note struct {
	ID           int64                   `json:"id"`
	RepoID       int64                   `json:"repoId"`
	Type         string                  `json:"type" gorm:"size:150;index:idx_type"` // normal diff_note
	DiscussionId string                  `json:"discussionId"`                        //讨论id
	OldCommitId  string                  `json:"oldCommitId"`
	NewCommitId  string                  `json:"newCommitId"`
	MergeId      int64                   `json:"-" gorm:"index:idx_merge_id"`
	Note         string                  `json:"note" gorm:"type:text"`
	Data         string                  `json:"-" gorm:"type:text"`
	DataResult   NoteData                `json:"data" gorm:"-"`
	AuthorId     string                  `json:"authorId"`
	AuthorUser   *apistructs.UserInfoDto `json:"author" gorm:"-"`
	CreatedAt    time.Time               `json:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt"`
	Score        int                     `json:"score" gorm:"size:150;index:idx_score"`
	Role         NoteRole                `json:"role"`
}

type NoteData struct {
	DiffLines   []*gitmodule.DiffLine `json:"diffLines"`
	OldPath     string                `json:"oldPath"`
	NewPath     string                `json:"newPath"`
	OldLine     int                   `json:"oldLine"`
	NewLine     int                   `json:"newLine"`
	OldLineTo   int                   `json:"oldLineTo"`
	NewLineTo   int                   `json:"newLineTo"`
	OldCommitId string                `json:"oldCommitId"`
	NewCommitId string                `json:"newCommitId"`

	AICodeReviewType AICodeReviewType `json:"aiCodeReviewType,omitempty"`
	AISessionID      string           `json:"aiSessionID,omitempty"`
}

func (svc *Service) CreateNote(repo *gitmodule.Repository, user *User, mr *apistructs.MergeRequestInfo, req NoteRequest) (note *Note, err error) {
	switch req.Type {
	case NoteTypeNormal:
		note, err = svc.constructNormalNote(repo, user, mr.Id, req)
	case NoteTypeDiffNote:
		note, err = svc.constructDiscussionNote(repo, user, mr.Id, req)
	case NoteTypeDiffNoteReply:
		note, err = svc.constructReplyDiscussionNote(repo, user, mr.Id, req)
	default:
		return nil, fmt.Errorf("not support note type: %s", req.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to construct note, err: %v", err)
	}

	// role
	if note.Role == "" {
		note.Role = NoteRoleUser
	}

	// AI
	if err := svc.handleAIRelatedNote(repo, user, mr, req, note); err != nil {
		return nil, err
	}

	// marshal note data
	noteDataBytes, err := json.Marshal(note.DataResult)
	if err != nil {
		return nil, err
	}
	note.Data = string(noteDataBytes)

	// create note
	err = svc.db.Create(note).Error
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (svc *Service) handleAIRelatedNote(repo *gitmodule.Repository, user *User, mr *apistructs.MergeRequestInfo, req NoteRequest, note *Note) error {
	// role
	if req.AICodeReviewType != "" {
		note.Role = NoteRoleAI
	}

	// AI code review
	if req.AICodeReviewType != "" {
		note.DataResult.AICodeReviewType = req.AICodeReviewType

		// get AI code review result as note content
		crReq := AICodeReviewNoteRequest{Type: req.AICodeReviewType, NoteLocation: req}
		switch req.AICodeReviewType {
		case AICodeReviewTypeMR:
		case AICodeReviewTypeMRFile:
			crReq.FileRelated = &AICodeReviewRequestForFile{}
		case AICodeReviewTypeMRCodeSnippet:
			selectedCode, _ := mrutil.ConvertDiffLinesToSnippet(note.DataResult.DiffLines)
			crReq.CodeSnippetRelated = &AICodeReviewRequestForCodeSnippet{SelectedCode: selectedCode}
		}
		reviewer, err := NewCodeReviewer(crReq, repo, user, mr)
		if err != nil {
			return fmt.Errorf("failed to create code reviewer err: %v", err)
		}
		suggestions := reviewer.CodeReview()
		note.Note = suggestions
	}

	// start AI session
	if req.StartAISession {
		aiSessionID, err := svc.createAISession(*note, user)
		if err != nil {
			return err
		}
		note.DataResult.AISessionID = aiSessionID
	}

	return nil
}

const (
	lineTypeOld = "old"
	lineTypeNew = "new"
)

func (svc *Service) constructDiscussionNote(repo *gitmodule.Repository, user *User, mergeId int64, request NoteRequest) (*Note, error) {
	if request.OldLineTo != 0 {
		if request.OldLineTo < request.OldLine {
			return nil, errors.New("invalid oldLineTo")
		}
	} else {
		request.OldLineTo = request.OldLine
	}
	if request.NewLineTo != 0 {
		if request.NewLineTo < request.NewLine {
			return nil, errors.New("invalid newLineTo")
		}
	} else {
		request.NewLineTo = request.NewLine
	}
	commitTo, err := repo.GetCommit(request.OldCommitId)
	if err != nil {
		return nil, err
	}
	commitFrom, err := repo.GetCommit(request.NewCommitId)
	if err != nil {
		return nil, err
	}

	diff, err := repo.GetDiffFile(commitFrom, commitTo, request.OldPath, request.NewPath)
	if err != nil {
		return nil, err
	}
	lineType := ""
	lineFrom := -1
	lineTo := -1
	if request.OldLine > 0 {
		lineType = lineTypeOld
		lineFrom = request.OldLine
		lineTo = request.OldLineTo
	} else if request.NewLine > 0 {
		lineType = lineTypeNew
		lineFrom = request.NewLine
		lineTo = request.NewLineTo
	}

	//找出section
	var diffSection *gitmodule.DiffSection
	for _, section := range diff.Sections {
		for _, diffLine := range section.Lines {
			if lineType == lineTypeOld {
				if diffLine.OldLineNo == lineFrom {
					diffSection = section
					break
				}
			} else {
				if diffLine.NewLineNo == lineFrom {
					diffSection = section
					break
				}
			}
		}
		if diffSection != nil {
			break
		}
	}

	if diffSection == nil {
		return nil, errors.New("invalid lineNo")
	}

	var lines []*gitmodule.DiffLine
	for _, v := range diffSection.Lines {
		if lineType == lineTypeOld {
			if v.OldLineNo <= lineTo {
				lines = append(lines, v)
			} else {
				break
			}
		}
		if lineType == lineTypeNew {
			if v.NewLineNo <= lineTo {
				lines = append(lines, v)
			} else {
				break
			}
		}
	}

	lineCount := len(lines)
	if lineCount > MAX_NOTE_DIFF_LINE_COUNT {
		lines = lines[lineCount-MAX_NOTE_DIFF_LINE_COUNT : lineCount]
	}

	noteData := NoteData{
		DiffLines:   lines,
		OldLine:     request.OldLine,
		NewLine:     request.NewLine,
		OldPath:     request.OldPath,
		NewPath:     request.NewPath,
		OldCommitId: request.OldCommitId,
		NewCommitId: request.NewCommitId,
	}

	note := Note{
		RepoID:       repo.ID,
		MergeId:      mergeId,
		Note:         request.Note,
		AuthorId:     user.Id,
		OldCommitId:  request.OldCommitId,
		NewCommitId:  request.NewCommitId,
		DiscussionId: guid.NewString(),
		Type:         NoteTypeDiffNote,
		DataResult:   noteData,
	}
	return &note, nil
}

func (svc *Service) constructReplyDiscussionNote(repo *gitmodule.Repository, user *User, mergeId int64, request NoteRequest) (*Note, error) {
	var oldDiffNote Note
	err := svc.db.Where("repo_id = ? and merge_id=? and discussion_id= ?",
		repo.ID, mergeId, request.DiscussionId).
		Limit(1).First(&oldDiffNote).Error

	//是否已存在,不存在已有DiscussionId不能新增
	if err != nil {
		return nil, err
	}

	note := Note{
		RepoID:       repo.ID,
		MergeId:      mergeId,
		Type:         NoteTypeDiffNoteReply,
		Note:         request.Note,
		AuthorId:     user.Id,
		DiscussionId: oldDiffNote.DiscussionId,
		OldCommitId:  oldDiffNote.OldCommitId,
		NewCommitId:  oldDiffNote.NewCommitId,
	}

	return &note, nil
}

func (svc *Service) constructNormalNote(repo *gitmodule.Repository, user *User, mergeId int64, req NoteRequest) (*Note, error) {
	note := Note{
		MergeId:  mergeId,
		Note:     req.Note,
		AuthorId: user.Id,
		RepoID:   repo.ID,
		Type:     NoteTypeNormal,
		Score:    req.Score,
	}
	return &note, nil
}

func (svc *Service) QueryAllNotes(repo *gitmodule.Repository, mergeId int64) ([]Note, error) {
	var notes []Note
	err := svc.db.Where("repo_id = ?  and merge_id=?", repo.ID, mergeId).Order("updated_at desc").Find(&notes).Error
	if err != nil {
		return nil, err
	}
	result := []Note{}
	for _, v := range notes {
		if v.AuthorId != "" {
			dto, err := uc.FindUserById(v.AuthorId)
			if err == nil {
				v.AuthorUser = dto
			} else {
				logrus.Errorf("get user from uc error: %v", err)
			}
		}
		if v.Data != "" {
			var data NoteData
			err := json.Unmarshal([]byte(v.Data), &data)
			if err != nil {
				logrus.Error(err)
			}
			v.DataResult = data
		}
		result = append(result, v)
	}
	return result, nil
}

func (svc *Service) QueryDiffNotes(repo *gitmodule.Repository, mergeId int64, oldCommitId string, newCommitId string) ([]Note, error) {
	var notes []Note
	err := svc.db.Where("repo_id = ?and merge_id=? and oldCommitId=? and newCommitId= ?",
		repo.ID, mergeId, oldCommitId, newCommitId).Find(&notes).Error
	if err != nil {
		return nil, err
	}
	for _, v := range notes {
		if v.AuthorId != "" {
			dto, err := uc.FindUserById(v.AuthorId)
			if err == nil {
				v.AuthorUser = dto
			} else {
				logrus.Errorf("get user from uc error: %v", err)
			}
		}
	}
	return notes, nil
}

func (svc *Service) createAISession(note Note, user *User) (string, error) {
	if !aiproxyclient.Instance.AIEnabled() {
		return "", aiproxyclient.ErrorAINotEnabled
	}
	sessionName := fmt.Sprintf("mr#%d-author#%s", note.MergeId, user.NickName)
	req := sessionpb.SessionCreateRequest{
		UserId:      user.Id,
		Scene:       "ai-cr-session",
		Name:        sessionName,
		Topic:       "This is a code review session, please answer professionally. \nRelated Code And Review: \n" + note.Note,
		NumOfCtxMsg: 100,
	}
	aiSession, err := aiproxyclient.Instance.Session().Create(aiproxyclient.Instance.Context(), &req)
	if err != nil {
		return "", err
	}
	return aiSession.Id, nil
}
