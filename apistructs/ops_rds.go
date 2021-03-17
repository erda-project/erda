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
