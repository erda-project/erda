package trace

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

const (
	compressionFlag    = "0"
	compressionBigText = 1 << 20 // 1MB
)

func BigStringAttribute(key string, data interface{}) attribute.KeyValue {
	text := fmt.Sprintf("%v", data)

	if len(text) < compressionBigText {
		return attribute.String(key, text)
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(text)); err != nil {
		fmt.Println("big string attribute is zip for error:", err)
		return attribute.String(key, text)
	}
	if err := gz.Close(); err != nil {
		fmt.Println("big string attribute is close gzip for error", err)
		return attribute.String(key, text)
	}
	return attribute.String(key, compressionFlag+b.String())
}
