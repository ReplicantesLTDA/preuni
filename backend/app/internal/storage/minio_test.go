package storage

import (
	"context"
	"strings"
	"testing"
)

func testConfig() Config {
	return Config{
		Endpoint:  "localhost:9000",
		AccessKey: "test-access",
		SecretKey: "test-secret",
		UseSSL:    false,
		Bucket:    "test-bucket",
	}
}

func TestNew_SucceedsWithValidConfig(t *testing.T) {
	c, err := New(testConfig())
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if c == nil {
		t.Fatal("New() returned nil client")
	}
}

func TestNew_ErrorsOnEmptyEndpoint(t *testing.T) {
	cfg := testConfig()
	cfg.Endpoint = ""
	if _, err := New(cfg); err == nil {
		t.Fatal("New() error = nil, want an error for an empty endpoint")
	}
}

func TestPresignedPutURL_ReturnsSignedURLWithoutNetworkAccess(t *testing.T) {
	// No MinIO server needs to be running: the fixed region (see minio.go's
	// `region` constant) makes this a purely local signing operation. If
	// that ever regresses, this test starts hanging/erroring on a
	// connection attempt instead of returning immediately.
	c, err := New(testConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	url, err := c.PresignedPutURL(context.Background(), "avatars/user-1/123.webp")
	if err != nil {
		t.Fatalf("PresignedPutURL() error = %v, want nil", err)
	}
	if url == "" {
		t.Fatal("PresignedPutURL() returned an empty URL")
	}
	if !strings.Contains(url, "test-bucket") {
		t.Errorf("PresignedPutURL() = %q, want it to contain the bucket name", url)
	}
	if !strings.Contains(url, "avatars/user-1/123.webp") {
		t.Errorf("PresignedPutURL() = %q, want it to contain the object key", url)
	}
	if !strings.Contains(url, "X-Amz-Signature=") {
		t.Errorf("PresignedPutURL() = %q, want a signed URL (X-Amz-Signature present)", url)
	}
}

func TestPublicURL_ConstructsExpectedFormat(t *testing.T) {
	cfg := testConfig()
	cfg.UseSSL = false
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := c.PublicURL("avatars/user-1/123.webp")
	want := "http://localhost:9000/test-bucket/avatars/user-1/123.webp"
	if got != want {
		t.Errorf("PublicURL() = %q, want %q", got, want)
	}
}

func TestPublicURL_UsesHTTPSWhenUseSSLIsTrue(t *testing.T) {
	cfg := testConfig()
	cfg.UseSSL = true
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := c.PublicURL("avatars/user-1/123.webp")
	if !strings.HasPrefix(got, "https://") {
		t.Errorf("PublicURL() = %q, want an https:// URL when UseSSL is true", got)
	}
}
