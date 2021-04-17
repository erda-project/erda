package exporter

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/recallsong/go-utils/encoding"
	"github.com/recallsong/go-utils/reflectx"
)

// consumer .
type consumer struct {
	log     logs.Logger
	filters map[string]string
	output  Output
}

type content struct {
	// encoding.RawBytes 不去解析具体内容
	ID        encoding.RawBytes            `json:"id"`
	Timestamp encoding.RawBytes            `json:"timestamp"`
	Source    encoding.RawBytes            `json:"source"`
	Content   encoding.RawBytes            `json:"content"`
	Offset    encoding.RawBytes            `json:"offset"`
	Stream    encoding.RawBytes            `json:"stream"`
	Tags      map[string]encoding.RawBytes `json:"tags"`
}
type labelscontent struct {
	content
	Labels map[string]string `json:"labels,omitempy"`
}

// Invoke .
func (c *consumer) Invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	var data labelscontent
	err := json.Unmarshal(value, &data)
	if err != nil || data.Tags == nil || data.Labels == nil {
		c.log.Warnf("invalid log message: %s", err)
		return nil
	}

	// 兼容
	key, ok := data.Tags["monitor_log_key"]
	if !ok {
		key, ok = data.Tags["terminus_log_key"]
		if ok {
			data.Tags["monitor_log_key"] = key
		}
	}
	if len(key) <= 2 {
		return nil
	}

	// do filter
	if len(c.filters) <= 0 {
		return nil
	}
	for k, v := range c.filters {
		val, ok := data.Tags[k]
		if !ok {
			return nil
		}
		if len(v) > 0 && v != reflectx.BytesToString([]byte(val)) {
			return nil
		}
	}

	delete(data.Labels, "monitor_log_output")
	delete(data.Labels, "monitor_log_output_config")
	value, err = json.Marshal(&data)
	if err != nil {
		return err
	}
	return c.output.Write(reflectx.BytesToString(key[1:len(key)-1]), value)
}
