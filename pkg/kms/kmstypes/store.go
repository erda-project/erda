package kmstypes

// Store the key information storage interface
type Store interface {
	// PluginKind is key store type
	GetKind() StoreKind

	// Create create and store new CMK
	CreateKey(info KeyInfo) error

	// GetKey use keyID to find CMK
	GetKey(keyID string) (KeyInfo, error)

	// ListByKind use plugin type to list CMKs
	ListKeysByKind(kind PluginKind) ([]string, error)

	// DeleteByKeyID use keyID to delete CMK
	DeleteByKeyID(keyID string) error

	// GetKeyVersion use keyID and keyVersionID to find keyVersion
	GetKeyVersion(keyID, keyVersionID string) (KeyVersionInfo, error)

	// RotateKeyVersion rotate key version
	RotateKeyVersion(keyID string, newKeyVersionInfo KeyVersionInfo) (KeyVersionInfo, error)
}
