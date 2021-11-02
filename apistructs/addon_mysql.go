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

package apistructs

type MySQLAccount struct {
	ID         string `json:"id"`
	InstanceID string `json:"instanceId"`
	Creator    string `json:"creator"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type MySQLAccountListData struct {
	List []MySQLAccount `json:"list"`
}

type ListMySQLAccountsRequest struct {
	InstanceID string `json:"instanceId"`
	ProjectID  uint64 `json:"projectId"`
}

type ListMySQLAccountResponse struct {
	Header
	Data MySQLAccountListData `json:"data"`
}
