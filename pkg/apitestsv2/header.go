package apitestsv2

import "net/http"

const headerAcceptEncoding = "Accept-Encoding"

// polishHeadersForCompression 优化用于压缩的 header
func polishHeadersForCompression(headers http.Header) http.Header {
	// polish header for compression
	headers.Del(headerAcceptEncoding)
	headers.Add(headerAcceptEncoding, "identity")
	return headers
}
