package encryption

//Hash for crypto Hash
type Hash uint

const (
	MD5 Hash = iota
	SHA1
	SHA224
	SHA256
	SHA384
	SHA512
	SHA512_224
	SHA512_256
)

//Encode defines the type of bytes encoded to string
type Encode uint

const (
	String Encode = iota
	HEX
	Base64
)

//Secret defines the private key type
type Secret uint

const (
	PKCS1 Secret = iota
	PKCS8
)

//Crypt defines crypt types
type Crypt uint

const (
	RSA Crypt = iota
)

type Padding uint

const (
	PaddingPKCS5 = iota
	PaddingPKCS7
)

type Cipher uint

const (
	ECB = iota
	CBC
	CFB
	OFB
)
