package common

type UCResponse[T any] struct {
	Success bool   `json:"success"`
	Result  T      `json:"result"`
	Error   string `json:"error"`
}

type ListLoginTypeResult struct {
	RegistryType []string `json:"registryType"`
}

type UCCurrentUser struct {
	ID          USERID `json:"id"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	LastLoginAt uint64 `json:"lastLoginAt"`
}
