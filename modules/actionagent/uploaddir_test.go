package actionagent

import (
	"fmt"
	"os"
	"testing"
)

func TestAgent_UploadDir(t *testing.T) {
	os.Setenv("DICE_OPENAPI_TOKEN", "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwaXBlbGluZSIsInBheWxvYWQiOnsiYWNjZXNzVG9rZW5FeHBpcmVkSW4iOiIwIiwiYWNjZXNzaWJsZUFQSXMiOlt7InBhdGgiOiIvYXBpL3BpcGVsaW5lcy9cdTAwM2NwaXBlbGluZUlEXHUwMDNlL3Rhc2tzL1x1MDAzY3Rhc2tJRFx1MDAzZS9hY3Rpb25zL2dldC1ib290c3RyYXAtaW5mbyIsIm1ldGhvZCI6IkdFVCIsInNjaGVtYSI6Imh0dHAifSx7InBhdGgiOiIvYXBpL3BpcGVsaW5lcy9cdTAwM2NwaXBlbGluZUlEXHUwMDNlL2FjdGlvbnMvcnVuIiwibWV0aG9kIjoiUE9TVCIsInNjaGVtYSI6Imh0dHAifSx7InBhdGgiOiIvYXBpL2ZpbGVzIiwibWV0aG9kIjoiUE9TVCIsInNjaGVtYSI6Imh0dHAifV0sIm1ldGFkYXRhIjp7IlVzZXItSUQiOiIyIiwicGlwZWxpbmVJRCI6IjEwMDAwMDk4IiwidGFza0lEIjoiMzQwIn19fQ.g2Ht8F-Rs3ly2BAYdjZFCxxDjAblK-xRAfbnbtx4P2iXcxxm4FxsZHukz33yAXXZHeNlkOuOiooBfPQk-KyORQ")
	agent := Agent{
		EasyUse: EasyUse{
			ContainerUploadDir: "/tmp/uploaddir",
			OpenAPIAddr:        "openapi.default.svc.cluster.local:9529",
		},
	}
	agent.uploadDir()
	for _, err := range agent.Errs {
		fmt.Println(err)
	}
}
