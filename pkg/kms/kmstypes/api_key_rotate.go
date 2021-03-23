package kmstypes

import "fmt"

type RotateKeyVersionRequest struct {
	KeyID string `json:"keyID,omitempty"`
}

func (req *RotateKeyVersionRequest) ValidateRequest() error {
	if req.KeyID == "" {
		return fmt.Errorf("missing keyID")
	}
	return nil
}

type RotateKeyVersionResponse struct {
	KeyMetadata KeyMetadata `json:"keyMetadata,omitempty"`
}
