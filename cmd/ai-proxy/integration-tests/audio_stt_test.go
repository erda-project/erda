// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// AudioTranscriptionRequest represents the audio transcription request structure
type AudioTranscriptionRequest struct {
	Model    string `json:"model"`
	Language string `json:"language,omitempty"`
	Prompt   string `json:"prompt,omitempty"`
}

// AudioTranscriptionResponse represents the audio transcription response structure
type AudioTranscriptionResponse struct {
	Text string `json:"text"`
}

func TestAudioTranscriptions(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Ensure test audio files exist
	audioHelper := NewAudioTestHelper()
	if err := audioHelper.EnsureTestAudioFiles(); err != nil {
		t.Fatalf("Failed to ensure test audio files: %v", err)
	}

	// Get audio models for testing
	audioModels := cfg.AudioSTTModels
	if len(audioModels) == 0 {
		t.Skip("No audio models configured for testing")
	}

	// Define test audio files
	testAudioFiles := []struct {
		filename        string
		expectedContent string // Expected transcription content for basic validation
	}{
		{"test_short.mp3", "hello world"},
		{"test_short.wav", "hello world"},
		{"test_short.m4a", "hello world"},
	}

	for _, model := range audioModels {
		for _, audioFile := range testAudioFiles {
			// Test with header method (existing)
			t.Run(fmt.Sprintf("Header_Model_%s_File_%s", model, audioFile.filename), func(t *testing.T) {
				testAudioTranscriptionForModelWithHeader(t, client, model, audioFile.filename, audioFile.expectedContent)
			})

			// Test with form data method (new)
			t.Run(fmt.Sprintf("FormData_Model_%s_File_%s", model, audioFile.filename), func(t *testing.T) {
				testAudioTranscriptionForModelWithFormData(t, client, model, audioFile.filename, audioFile.expectedContent)
			})
		}
	}
}

func testAudioTranscriptionForModelWithHeader(t *testing.T, client *common.Client, model, filename, expectedContent string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Check if audio file exists
	audioPath := filepath.Join("testdata", filename)
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Skipf("Audio file %s not found, skipping test", audioPath)
		return
	}

	// Read audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		t.Fatalf("✗ Failed to open audio file %s: %v", audioPath, err)
	}
	defer audioFile.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Don't add model field in form, pass via header instead
	t.Logf("Debug: Setting model in header: %s", model)

	// Add audio file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("✗ Failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, audioFile); err != nil {
		t.Fatalf("✗ Failed to copy audio file: %v", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("✗ Failed to close writer: %v", err)
	}

	// Send request with model passed via header
	headers := map[string]string{
		"X-AI-Proxy-Model": model,
	}
	startTime := time.Now()
	resp := client.PostMultipartWithHeaders(ctx, "/v1/audio/transcriptions", &buf, contentType, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var transcriptionResp AudioTranscriptionResponse
	if err := resp.GetJSON(&transcriptionResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if transcriptionResp.Text == "" {
		t.Error("✗ Empty transcription text")
	}

	// Basic content validation (tolerant of transcription differences)
	transcribedText := strings.ToLower(strings.TrimSpace(transcriptionResp.Text))
	if expectedContent != "" {
		// Check if contains expected keywords
		expectedWords := strings.Fields(strings.ToLower(expectedContent))
		foundWords := 0
		for _, word := range expectedWords {
			if strings.Contains(transcribedText, word) {
				foundWords++
			}
		}

		// Require at least half of expected words to be found
		requiredWords := len(expectedWords) / 2
		if requiredWords == 0 {
			requiredWords = 1
		}

		if foundWords < requiredWords {
			t.Logf("⚠ Transcription quality warning: expected words '%s', got '%s' (found %d/%d words)",
				expectedContent, transcribedText, foundWords, len(expectedWords))
		}
	}

	// Check response time reasonableness
	if responseTime > 30*time.Second {
		t.Logf("⚠ Slow transcription response time: %v", responseTime)
	}

	t.Logf("✓ Model %s transcribed %s: '%s' (response time: %v)",
		model, filename, truncateString(transcribedText, 100), responseTime)
}

func testAudioTranscriptionForModelWithFormData(t *testing.T, client *common.Client, model, filename, expectedContent string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Check if audio file exists
	audioPath := filepath.Join("testdata", filename)
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Skipf("Audio file %s not found, skipping test", audioPath)
		return
	}

	// Read audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		t.Fatalf("✗ Failed to open audio file %s: %v", audioPath, err)
	}
	defer audioFile.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add model field in form
	t.Logf("Debug: Setting model in form data: %s", model)
	if err := writer.WriteField("model", model); err != nil {
		t.Fatalf("✗ Failed to write model field: %v", err)
	}

	// Add audio file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("✗ Failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, audioFile); err != nil {
		t.Fatalf("✗ Failed to copy audio file: %v", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("✗ Failed to close writer: %v", err)
	}

	// Send request without using header to pass model
	startTime := time.Now()
	resp := client.PostMultipart(ctx, "/v1/audio/transcriptions", &buf, contentType)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var transcriptionResp AudioTranscriptionResponse
	if err := resp.GetJSON(&transcriptionResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if transcriptionResp.Text == "" {
		t.Error("✗ Empty transcription text")
	}

	// Basic content validation (tolerant of transcription differences)
	transcribedText := strings.ToLower(strings.TrimSpace(transcriptionResp.Text))
	if expectedContent != "" {
		// Check if contains expected keywords
		expectedWords := strings.Fields(strings.ToLower(expectedContent))
		foundWords := 0
		for _, word := range expectedWords {
			if strings.Contains(transcribedText, word) {
				foundWords++
			}
		}

		// Require at least half of expected words to be found
		requiredWords := len(expectedWords) / 2
		if requiredWords == 0 {
			requiredWords = 1
		}

		if foundWords < requiredWords {
			t.Logf("⚠ Transcription quality warning: expected words '%s', got '%s' (found %d/%d words)",
				expectedContent, transcribedText, foundWords, len(expectedWords))
		}
	}

	// Check response time reasonableness
	if responseTime > 30*time.Second {
		t.Logf("⚠ Slow transcription response time: %v", responseTime)
	}

	t.Logf("✓ Model %s transcribed %s via form data: '%s' (response time: %v)",
		model, filename, truncateString(transcribedText, 100), responseTime)
}

// TestAudioTranscriptionsErrorHandling tests error handling
func TestAudioTranscriptionsErrorHandling(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Ensure test audio files exist
	audioHelper := NewAudioTestHelper()
	if err := audioHelper.EnsureTestAudioFiles(); err != nil {
		t.Fatalf("Failed to ensure test audio files: %v", err)
	}

	audioModels := cfg.AudioSTTModels
	if len(audioModels) == 0 {
		t.Skip("No audio models configured for testing")
	}

	model := audioModels[0] // Use first model for error testing

	t.Run("UnsupportedFileFormat", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Create a fake text file as audio file
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		if _, err := part.Write([]byte("This is not an audio file")); err != nil {
			t.Fatalf("Failed to write fake audio content: %v", err)
		}

		contentType := writer.FormDataContentType()
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostMultipartWithHeaders(ctx, "/v1/audio/transcriptions", &buf, contentType, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with unsupported file format, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected unsupported file format (status: %d)", resp.StatusCode)
		}
	})

	t.Run("MissingFile", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Create request without file
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		contentType := writer.FormDataContentType()
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostMultipartWithHeaders(ctx, "/v1/audio/transcriptions", &buf, contentType, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with missing file, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected request with missing file (status: %d)", resp.StatusCode)
		}
	})

	t.Run("InvalidModel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Check if test audio file is available
		audioPath := filepath.Join("testdata", "test_short.mp3")
		if _, err := os.Stat(audioPath); os.IsNotExist(err) {
			t.Skip("No test audio file available")
			return
		}

		audioFile, err := os.Open(audioPath)
		if err != nil {
			t.Fatalf("Failed to open audio file: %v", err)
		}
		defer audioFile.Close()

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		part, err := writer.CreateFormFile("file", "test_short.mp3")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		if _, err := io.Copy(part, audioFile); err != nil {
			t.Fatalf("Failed to copy audio file: %v", err)
		}

		contentType := writer.FormDataContentType()
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		// Use non-existent model
		headers := map[string]string{
			"X-AI-Proxy-Model": "non-existent-model",
		}
		resp := client.PostMultipartWithHeaders(ctx, "/v1/audio/transcriptions", &buf, contentType, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with invalid model, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected invalid model (status: %d)", resp.StatusCode)
		}
	})
}

// AudioTestHelper provides methods to generate test audio files
type AudioTestHelper struct {
	testDataDir string
}

// NewAudioTestHelper creates a new audio test helper
func NewAudioTestHelper() *AudioTestHelper {
	return &AudioTestHelper{
		testDataDir: "testdata",
	}
}

// EnsureTestAudioFiles creates test audio files if they don't exist
func (h *AudioTestHelper) EnsureTestAudioFiles() error {
	// Ensure testdata directory exists
	if err := os.MkdirAll(h.testDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create testdata directory: %v", err)
	}

	// Define required test files
	testFiles := []struct {
		filename string
		format   string
	}{
		{"test_short.wav", "wav"},
		{"test_short.mp3", "mp3"},
		{"test_short.m4a", "m4a"},
	}

	for _, file := range testFiles {
		filePath := filepath.Join(h.testDataDir, file.filename)

		// Check if file already exists
		if _, err := os.Stat(filePath); err == nil {
			continue // File exists, skip
		}

		// Generate the audio file using say + ffmpeg
		if err := h.generateWithSayAndFFmpeg(filePath, file.format); err != nil {
			return fmt.Errorf("failed to generate %s: %v", file.filename, err)
		}
	}

	return nil
}

// generateWithSayAndFFmpeg uses macOS 'say' command + ffmpeg exactly as described in README
func (h *AudioTestHelper) generateWithSayAndFFmpeg(filePath, format string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("say command only available on macOS")
	}

	// Check if both 'say' and 'ffmpeg' commands exist
	if _, err := exec.LookPath("say"); err != nil {
		return fmt.Errorf("say command not found: %v", err)
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg command not found: %v", err)
	}

	// Step 1: Generate AIFF with 'say' (default format)
	tempAiff := filePath + ".temp.aiff"
	defer os.Remove(tempAiff)

	cmd := exec.Command("say", "hello world", "-o", tempAiff)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate AIFF with say: %v", err)
	}

	// Step 2: Convert AIFF to target format using ffmpeg
	var convertCmd *exec.Cmd
	switch format {
	case "wav":
		// ffmpeg -i temp.aiff test_short.wav
		convertCmd = exec.Command("ffmpeg", "-i", tempAiff, "-y", filePath)
	case "mp3":
		// ffmpeg -i temp.aiff test_short.mp3
		convertCmd = exec.Command("ffmpeg", "-i", tempAiff, "-y", filePath)
	case "m4a":
		// ffmpeg -i temp.aiff test_short.m4a
		convertCmd = exec.Command("ffmpeg", "-i", tempAiff, "-y", filePath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err := convertCmd.Run(); err != nil {
		return fmt.Errorf("failed to convert to %s with ffmpeg: %v", format, err)
	}

	return nil
}

// CheckTestAudioFiles verifies that all required test audio files exist
func (h *AudioTestHelper) CheckTestAudioFiles() ([]string, []string) {
	requiredFiles := []string{"test_short.wav", "test_short.mp3", "test_short.m4a"}
	var existing, missing []string

	for _, filename := range requiredFiles {
		filePath := filepath.Join(h.testDataDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			existing = append(existing, filename)
		} else {
			missing = append(missing, filename)
		}
	}

	return existing, missing
}
