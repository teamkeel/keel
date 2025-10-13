package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAcceptsGzip(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		expected       bool
	}{
		{
			name:           "accepts gzip",
			acceptEncoding: "gzip",
			expected:       true,
		},
		{
			name:           "accepts gzip with quality",
			acceptEncoding: "gzip;q=1.0",
			expected:       true,
		},
		{
			name:           "accepts multiple encodings including gzip",
			acceptEncoding: "deflate, gzip, br",
			expected:       true,
		},
		{
			name:           "accepts gzip with spaces",
			acceptEncoding: "deflate, gzip , br",
			expected:       true,
		},
		{
			name:           "does not accept gzip",
			acceptEncoding: "deflate, br",
			expected:       false,
		},
		{
			name:           "empty accept encoding",
			acceptEncoding: "",
			expected:       false,
		},
		{
			name:           "only identity",
			acceptEncoding: "identity",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.acceptEncoding != "" {
				headers.Set("Accept-Encoding", tt.acceptEncoding)
			}

			result := AcceptsGzip(headers)
			if result != tt.expected {
				t.Errorf("AcceptsGzip() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestShouldCompress(t *testing.T) {
	largeBody := make([]byte, MinCompressionSize+1)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	smallBody := make([]byte, MinCompressionSize-1)
	for i := range smallBody {
		smallBody[i] = 'a'
	}

	tests := []struct {
		name           string
		body           []byte
		acceptEncoding string
		contentEnc     string
		expected       bool
	}{
		{
			name:           "large body with gzip support",
			body:           largeBody,
			acceptEncoding: "gzip",
			expected:       true,
		},
		{
			name:           "small body with gzip support",
			body:           smallBody,
			acceptEncoding: "gzip",
			expected:       false,
		},
		{
			name:           "large body without gzip support",
			body:           largeBody,
			acceptEncoding: "deflate",
			expected:       false,
		},
		{
			name:           "large body already compressed",
			body:           largeBody,
			acceptEncoding: "gzip",
			contentEnc:     "gzip",
			expected:       false,
		},
		{
			name:           "empty body",
			body:           []byte{},
			acceptEncoding: "gzip",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.acceptEncoding != "" {
				headers.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEnc != "" {
				headers.Set("Content-Encoding", tt.contentEnc)
			}

			result := ShouldCompress(tt.body, headers)
			if result != tt.expected {
				t.Errorf("ShouldCompress() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCompress(t *testing.T) {
	testData := []byte("This is a test string that should be compressed using gzip compression algorithm. " +
		"It needs to be long enough to see compression benefits and test the functionality properly.")

	compressed, err := Compress(testData)
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("Compress() returned empty result")
	}

	// Compressed data should be smaller than original for this test case
	if len(compressed) >= len(testData) {
		t.Errorf("Compressed data (%d bytes) is not smaller than original (%d bytes)",
			len(compressed), len(testData))
	}

	// Verify we can decompress it back
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed, testData) {
		t.Errorf("Decompressed data does not match original")
	}
}

func TestSetCompressionHeaders(t *testing.T) {
	t.Run("sets headers on empty header map", func(t *testing.T) {
		headers := http.Header{}
		SetCompressionHeaders(headers)

		if headers.Get("Content-Encoding") != "gzip" {
			t.Errorf("Content-Encoding = %v, expected gzip", headers.Get("Content-Encoding"))
		}

		if headers.Get("Vary") != "Accept-Encoding" {
			t.Errorf("Vary = %v, expected Accept-Encoding", headers.Get("Vary"))
		}
	})

	t.Run("appends to existing Vary header", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Vary", "Origin")
		SetCompressionHeaders(headers)

		if headers.Get("Content-Encoding") != "gzip" {
			t.Errorf("Content-Encoding = %v, expected gzip", headers.Get("Content-Encoding"))
		}

		vary := headers.Get("Vary")
		if !strings.Contains(vary, "Origin") || !strings.Contains(vary, "Accept-Encoding") {
			t.Errorf("Vary = %v, expected to contain both Origin and Accept-Encoding", vary)
		}
	})

	t.Run("does not duplicate Accept-Encoding in Vary", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Vary", "Accept-Encoding, Origin")
		SetCompressionHeaders(headers)

		vary := headers.Get("Vary")
		// Should not add Accept-Encoding again
		if vary != "Accept-Encoding, Origin" {
			t.Errorf("Vary = %v, expected Accept-Encoding, Origin", vary)
		}
	})
}

func TestCompressEmptyBody(t *testing.T) {
	compressed, err := Compress([]byte{})
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// Even empty body should produce valid gzip output
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if len(decompressed) != 0 {
		t.Errorf("Expected empty decompressed body, got %d bytes", len(decompressed))
	}
}

func TestCompressionLevel(t *testing.T) {
	testData := make([]byte, 10000)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	compressed, err := Compress(testData)
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// Verify the compression level by checking it produces reasonable compression
	// DefaultCompression should achieve some compression on this data
	compressionRatio := float64(len(compressed)) / float64(len(testData))
	if compressionRatio > 0.95 {
		t.Errorf("Compression ratio %.2f is too high, compression may not be working effectively", compressionRatio)
	}
}
