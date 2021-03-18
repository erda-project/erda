package kmstypes

import (
	"encoding/base64"
	"fmt"
)

type EncryptRequest struct {
	KeyID string `json:"keyID,omitempty"`
	// Required. The data to encrypt. Must be no larger than 64KiB.
	// A base64-encoded string.
	PlaintextBase64 string `json:"plaintextBase64,omitempty"`
}

func (req *EncryptRequest) ValidateRequest() error {
	if req.KeyID == "" {
		return fmt.Errorf("missing keyID")
	}
	if len(req.PlaintextBase64) == 0 {
		return fmt.Errorf("missing plaintextBase64")
	}
	if _, err := base64.StdEncoding.DecodeString(string(req.PlaintextBase64)); err != nil {
		return fmt.Errorf("cannot decode base64 plaintext, err: %v", err)
	}
	return nil
}

type EncryptResponse struct {
	KeyID string `json:"keyID,omitempty"`
	// The encrypted data.
	// A base64-encoded string.
	CiphertextBase64 string `json:"ciphertextBase64,omitempty"`
}

type DecryptRequest struct {
	KeyID string `json:"keyID,omitempty"`
	// The encrypted data.
	// A base64-encoded string.
	CiphertextBase64 string `json:"ciphertextBase64,omitempty"`
}

func (req *DecryptRequest) ValidateRequest() error {
	if req.KeyID == "" {
		return fmt.Errorf("missing keyID")
	}
	if len(req.CiphertextBase64) == 0 {
		return fmt.Errorf("missing ciphertextBase64")
	}
	if _, err := base64.StdEncoding.DecodeString(string(req.CiphertextBase64)); err != nil {
		return fmt.Errorf("cannot decode base64 ciphertext, err: %v", err)
	}
	return nil
}

type DecryptResponse struct {
	PlaintextBase64 string `json:"plaintextBase64,omitempty"`
}
