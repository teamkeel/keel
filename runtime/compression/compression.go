package compression

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"
)

const (
	// MinCompressionSize is the minimum response size in bytes before compression is applied
	// Responses smaller than this are not worth compressing due to overhead.
	MinCompressionSize = 1024

	// DefaultCompressionLevel provides a good balance between speed and compression ratio.
	DefaultCompressionLevel = gzip.DefaultCompression
)

// AcceptsGzip checks if the client accepts gzip encoding based on the Accept-Encoding header.
func AcceptsGzip(headers http.Header) bool {
	acceptEncoding := headers.Get("Accept-Encoding")
	if acceptEncoding == "" {
		return false
	}

	// Check if gzip is in the list of accepted encodings
	encodings := strings.Split(acceptEncoding, ",")
	for _, encoding := range encodings {
		encoding = strings.TrimSpace(encoding)
		// Handle quality values like "gzip;q=1.0"
		if strings.HasPrefix(encoding, "gzip") {
			return true
		}
	}
	return false
}

// ShouldCompress determines if a response should be compressed based on size and client support.
func ShouldCompress(body []byte, headers http.Header) bool {
	// Check if the response is large enough to benefit from compression
	if len(body) < MinCompressionSize {
		return false
	}

	// Check if client accepts gzip
	if !AcceptsGzip(headers) {
		return false
	}

	// Check if content is already compressed
	contentEncoding := headers.Get("Content-Encoding")
	return contentEncoding == ""
}

// Compress compresses the given byte slice using gzip.
func Compress(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, DefaultCompressionLevel)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(body)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SetCompressionHeaders sets the appropriate headers for a compressed response.
// It does NOT set Content-Length - this is handled automatically by http.ResponseWriter.
func SetCompressionHeaders(headers http.Header) {
	headers.Set("Content-Encoding", "gzip")
	// Note: We deliberately do NOT set Content-Length here.
	// The http.ResponseWriter will automatically set it based on the actual
	// bytes written in the Write() call.
	// Indicate that the response varies based on Accept-Encoding
	// This is important for caching proxies
	vary := headers.Get("Vary")
	if vary == "" {
		headers.Set("Vary", "Accept-Encoding")
	} else if !strings.Contains(vary, "Accept-Encoding") {
		headers.Set("Vary", vary+", Accept-Encoding")
	}
}
