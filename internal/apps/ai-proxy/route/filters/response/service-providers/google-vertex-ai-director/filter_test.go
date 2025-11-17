package google_vertex_ai_director

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/genai"
)

func TestConvertGenAIResponseToOpenAIImageGenerationResponse(t *testing.T) {
	resp := genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{InlineData: &genai.Blob{Data: []byte{0x01, 0x02}}},
						{FileData: &genai.FileData{FileURI: "https://example.com/image.png"}},
					},
				},
			},
		},
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal stub response: %v", err)
	}

	outBytes, err := convertGenAIResponseToOpenAIImageGenerationResponse(raw)
	if err != nil {
		t.Fatalf("convertGenAIResponseToOpenAIImageGenerationResponse error: %v", err)
	}

	var openAIResp openai.ImageResponse
	if err := json.Unmarshal(outBytes, &openAIResp); err != nil {
		t.Fatalf("failed to unmarshal converted response: %v", err)
	}

	if len(openAIResp.Data) != 2 {
		t.Fatalf("expected 2 data items, got %d", len(openAIResp.Data))
	}
	if openAIResp.Data[0].B64JSON != base64.StdEncoding.EncodeToString([]byte{0x01, 0x02}) {
		t.Fatalf("unexpected base64 data: %s", openAIResp.Data[0].B64JSON)
	}
	if openAIResp.Data[1].URL != "https://example.com/image.png" {
		t.Fatalf("unexpected image url: %s", openAIResp.Data[1].URL)
	}
}
