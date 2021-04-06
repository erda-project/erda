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

package dbclient

type DicePipelineSnippetClient struct {
	ID    uint64                     `json:"id" xorm:"pk autoincr"`
	Name  string                     `json:"name"`
	Host  string                     `json:"host"`
	Extra PipelineSnippetClientExtra `json:"extra" xorm:"json"`
}

func (ps *DicePipelineSnippetClient) TableName() string {
	return "dice_pipeline_snippet_clients"
}

type PipelineSnippetClientExtra struct {
	UrlPathPrefix string `json:"urlPathPrefix"`
}

func (client *Client) FindSnippetClientList() (clients []*DicePipelineSnippetClient, err error) {

	err = client.Find(&clients)
	if err != nil {
		return nil, err
	}

	return clients, err
}
