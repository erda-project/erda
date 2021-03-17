package kmstypes

type RequestValidator interface {
	ValidateRequest() error
}

const (
	CtxKeyKmsRequestID = "KmsRequestID"
)

const (
	PluginKind_DICE_KMS   PluginKind = "DICE_KMS"
	PluginKind_AWS_KMS    PluginKind = "AWS_KMS"
	PluginKind_ALIYUN_KMS PluginKind = "ALIYUN_KMS"

	StoreKind_ETCD  StoreKind = "ETCD"
	StoreKind_MYSQL StoreKind = "MYSQL"

	CustomerMasterKeySpec_SYMMETRIC_DEFAULT   CustomerMasterKeySpec = "SYMMETRIC_DEFAULT" // AES-256-GCM ; default
	CustomerMasterKeySpec_ASYMMETRIC_RSA_2048 CustomerMasterKeySpec = "RSA_2048"
	CustomerMasterKeySpec_ASYMMETRIC_RSA_3072 CustomerMasterKeySpec = "RSA_3072"
	CustomerMasterKeySpec_ASYMMETRIC_RSA_4096 CustomerMasterKeySpec = "RSA_4096"

	KeyUsage_ENCRYPT_DECRYPT KeyUsage = "ENCRYPT_DECRYPT"
	KeyUsage_SIGN_VERIFY     KeyUsage = "SIGN_VERIFY"

	KeyStateEnabled         KeyState = "Enabled"
	KeyStateDisabled        KeyState = "Disabled"
	KeyStatePendingDeletion KeyState = "PendingDeletion"
	KeyStatePendingImport   KeyState = "PendingImport"
	KeyStateUnavailable     KeyState = "Unavailable"
)
