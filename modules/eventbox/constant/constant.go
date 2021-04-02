package constant

import (
	"path/filepath"
)

var (
	EventboxDir = "/eventbox"
	MessageDir  = filepath.Join(EventboxDir, "messages")

	// schemon input
	// ScheMonServiceLockKey  = filepath.Join(EventboxDir, "schemonservicelock")
	// ScheMonServiceWatchDir = filepath.Join("/dice/service")
	// ScheMonServiceQueryURL = "/v1/notify/runtime/%s/%s" // /v1/notify/runtime/<namespace>/<name>
	ScheMonJobLockKey  = filepath.Join(EventboxDir, "schemonjoblock")
	ScheMonJobWatchDir = filepath.Join("/dice/job/")
	ScheMonJobQueryURL = "/v1/notify/job/%s/%s" // /v1/notify/job/<namespace>/<name>

	// register
	RegisterDir      = filepath.Join(EventboxDir, "register")
	RegisterLabelKey = "/REGISTERED_LABEL"

	// webhook
	WebhookLabelKey = "/WEBHOOK"
	WebhookDir      = filepath.Join(EventboxDir, "webhook")
)
