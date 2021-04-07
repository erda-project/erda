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

package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/pkg/uuid"
//)
//
//var (
//	b        *Bundle
//	ticketID int64
//	err      error
//)
//
//// TestTicket 工单测试集
//func TestTicket(t *testing.T) {
//	// setup
//	os.Setenv("CMDB_ADDR", "http://cmdb.marathon.l4lb.thisdcos.directory:9093")
//	b = New(WithCMDB())
//
//	t.Run("ticket abnormal create", TestAbnormalCreateTicket)
//	t.Run("ticket normal create", TestNormalCreateTicket)
//	t.Run("ticket normal list", TestNormalListTicket)
//	t.Run("comment normal create", TestNormalCreateComment)
//	// tear-down
//	t.Run("ticket normal delete", TestDeleteNormalTicket)
//	os.Unsetenv("CMDB_ADDR")
//}
//
//// TestAbnormalCreateTicket Ticket异常创建场景
//func TestAbnormalCreateTicket(t *testing.T) {
//	abnormalReq := apistructs.TicketCreateRequest{
//		Title:      "unit test ticket",
//		Content:    "unit test content",
//		Type:       "fix",
//		Priority:   "low",
//		UserID:     "2",
//		TargetType: "project",
//		TargetID:   "1",
//	}
//
//	ticketID, err := b.CreateTicket(uuid.UUID(), &abnormalReq)
//	assert.True(t, ticketID == 0)
//	assert.True(t, err != nil)
//}
//
//// TestNormalCreateTicket Ticket正常创建场景
//func TestNormalCreateTicket(t *testing.T) {
//	req := apistructs.TicketCreateRequest{
//		Title:      "unit test ticket",
//		Content:    "unit test content",
//		Type:       "task",
//		Priority:   "low",
//		UserID:     "2",
//		TargetType: "project",
//		TargetID:   "1",
//	}
//
//	ticketID, err = b.CreateTicket(uuid.UUID(), &req)
//	assert.True(t, ticketID > 0)
//	assert.Nil(t, err)
//}
//
//// TestNormalCloseTicket Ticket正常关闭场景
//func TestNormalCloseTicket(t *testing.T) {
//	ticketID, err = b.CloseTicket(ticketID, "2")
//	assert.True(t, ticketID > 0)
//	assert.Nil(t, err)
//}
//
//// TestNormalListTicket Ticket列表正常场景
//func TestNormalListTicket(t *testing.T) {
//	req := apistructs.TicketListRequest{
//		//Type:       "task",
//		Priority:   "low",
//		Status:     "open",
//		TargetType: "project",
//		TargetID:   "1",
//	}
//	tickets, err := b.ListTicket(req)
//	assert.Nil(t, err)
//	assert.True(t, tickets.Total > 0)
//	assert.True(t, len(tickets.Tickets) > 0)
//}
//
//// TestNormalCreateComment Comment正常创建场景
//func TestNormalCreateComment(t *testing.T) {
//	req := apistructs.CommentCreateRequest{
//		TicketID: ticketID,
//		Content:  "comment test content",
//		UserID:   "2",
//	}
//
//	commentID, err := b.CreateComment(&req)
//	assert.True(t, commentID > 0)
//	assert.Nil(t, err)
//}
//
//// TestDeleteNormalTicket Ticket删除正常场景
//func TestDeleteNormalTicket(t *testing.T) {
//	id, err := b.DeleteTicket(ticketID)
//	assert.Nil(t, err)
//	assert.Equal(t, ticketID, id)
//}
