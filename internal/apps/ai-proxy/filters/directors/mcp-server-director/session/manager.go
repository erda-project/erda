package session

type Manager interface {
	Load(sessionId string) (*ServerInfo, error)
	Save(sessionId string, info *ServerInfo) error
}

type ServerInfo struct {
	Host   string
	Scheme string
}
