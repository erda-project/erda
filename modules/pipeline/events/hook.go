package events

type Hook interface {
	HandleWebhook
	HandleWebSocket
	HandleDingDing
	HandleHTTP
}

type HandleWebhook interface{ HandleWebhook() error }
type HandleWebSocket interface{ HandleWebSocket() error }
type HandleDingDing interface{ HandleDingDing() error }
type HandleHTTP interface{ HandleHTTP() error }

type HookType string

const (
	HookTypeWebHook   HookType = "WEBHOOK"
	HookTypeWebSocket HookType = "WEBSOCKET"
	HookTypeDINGDING  HookType = "DINGDING"
	HookTypeHTTP      HookType = "HTTP"
)
