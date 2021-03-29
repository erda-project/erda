package snapshot

import "encoding/json"

type Snapshot interface {
	Export() json.RawMessage
	Import(json.RawMessage) error
}
