package settings

import (
	"encoding/json"
	"time"
)

type ttl struct {
	TTL    int64 `json:"ttl"`
	HotTTL int64 `json:"hot_ttl"`
}

type ttlConfigMap struct {
	TTL    string `json:"ttl"`
	HotTTL string `json:"hot_ttl"`
}

func (t *ttl) MarshalJSON() ([]byte, error) {
	res := ttlConfigMap{
		TTL:    time.Duration(t.TTL * 24 * int64(time.Hour)).String(),
		HotTTL: time.Duration(t.HotTTL * 24 * int64(time.Hour)).String(),
	}

	return json.Marshal(&res)
}
