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

package apistructs

type CloudResourceMysqlListAccountRequest struct {
	Vendor string `query:"vendor"`
	Region string `query:"region"`
	// get from request path
	InstanceID string `query:"instanceID"`
}

type CloudResourceMysqlListAccountResponse struct {
	Header
	Data CloudResourceMysqlListAccountData `json:"data"`
}

type CloudResourceMysqlListAccountData struct {
	List []CloudResourceMysqlListAccountItem `json:"list"`
}

type CloudResourceMysqlListAccountItem struct {
	AccountName        string                                `json:"accountName"`
	AccountStatus      string                                `json:"accountStatus"`
	AccountType        string                                `json:"accountType"`
	AccountDescription string                                `json:"accountDescription"`
	DatabasePrivileges []CloudResourceMysqlAccountPrivileges `json:"databasePrivileges"`
}

type CloudResourceMysqlAccountPrivileges struct {
	DBName           string `json:"dBName"`
	AccountPrivilege string `json:"accountPrivilege"`
}

type CloudResourceMysqlListDatabaseRequest struct {
	Vendor string `query:"vendor"`
	Region string `query:"region"`
	// get from request path
	InstanceID string `query:"instanceID"`
}

type CloudResourceMysqlListDatabaseResponse struct {
	Header
	Data CloudResourceMysqlListDatabaseData `json:"data"`
}

type CloudResourceMysqlListDatabaseData struct {
	List []CloudResourceMysqlListDatabaseItem `json:"list"`
}

type CloudResourceMysqlListDatabaseItem struct {
	DBName           string                                  `json:"dBName"`
	DBStatus         string                                  `json:"dBStatus"`
	CharacterSetName string                                  `json:"characterSetName"`
	DBDescription    string                                  `json:"dBDescription"`
	Accounts         []CloudResourceMysqlListDatabaseAccount `json:"accounts"`
}

type CloudResourceMysqlListDatabaseAccount struct {
	Account string `query:"account"`
}
