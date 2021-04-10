package websocket

// 主动离开
type ExitWithDoingNothing struct {
}

func (e ExitWithDoingNothing) Error() string {
	return "exit 0"
}
