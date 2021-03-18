package nexus

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNexus_ContentSelectorListRequest(t *testing.T) {
	selectors, err := n.ContentSelectorListRequest(ContentSelectorListRequest{})
	assert.NoError(t, err)
	s, _ := json.MarshalIndent(&selectors, "", "  ")
	fmt.Println(string(s))
}

func TestNexus_ContentSelectorCreateRequest(t *testing.T) {
	err := n.ContentSelectorCreateRequest(ContentSelectorCreateRequest{
		Name:        "test-content-selector",
		Description: "test content selector",
		Expression:  `format == "maven2" and path =^ "/org/sonatype/nexus"`,
	})
	assert.NoError(t, err)
}

func TestNexus_ContentSelectorGetRequest(t *testing.T) {
	selector, err := n.ContentSelectorGetRequest(ContentSelectorGetRequest{
		ContentSelectorName: "test-content-selector",
	})
	assert.NoError(t, err)
	s, _ := json.MarshalIndent(&selector, "", "  ")
	fmt.Println(string(s))
}

func TestNexus_UpdateContentSelector(t *testing.T) {
	err := n.UpdateContentSelector(ContentSelectorUpdateRequest{
		ContentSelectorName: "test-content-selector",
		Description:         "ssssssss",
		Expression:          `format == "maven2" and path =^ "/org/sonatype"`,
	})
	assert.NoError(t, err)
}

func TestNexus_DeleteContentSelector(t *testing.T) {
	err := n.DeleteContentSelector(ContentSelectorDeleteRequest{
		ContentSelectorName: "test-content-selector",
	})
	assert.NoError(t, err)
}
