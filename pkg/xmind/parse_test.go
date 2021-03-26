package xmind

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestParseByXMindparser(t *testing.T) {
//	f, err := os.Open("./examples/content.json")
//	assert.NoError(t, err)
//	content, err := Parse(f)
//	assert.NoError(t, err)
//	b, _ := json.MarshalIndent(content, "", "  ")
//	fmt.Println(string(b))
//}

func TestParseByJSON(t *testing.T) {
	f, err := os.Open("./examples/content.json")
	assert.NoError(t, err)
	var content Content
	err = json.NewDecoder(f).Decode(&content)
	assert.NoError(t, err)

	b, _ := xml.MarshalIndent(&content, "", "  ")
	fmt.Println(string(b))
}
