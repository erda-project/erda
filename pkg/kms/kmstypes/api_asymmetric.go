package kmstypes

type AsymmetricDecryptRequest struct {
	KeyID            string `json:"keyID,omitempty"`
	CiphertextBase64 []byte `json:"ciphertextBase64,omitempty"`
}
type AsymmetricDecryptResponse struct {
	PlaintextBase64 []byte `json:"plaintextBase64,omitempty"`
}

type GetPublicKeyRequest struct {
	KeyID string `json:"keyID,omitempty"`
}

type PublicKey struct {
	Pem       string `json:"pem,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}
