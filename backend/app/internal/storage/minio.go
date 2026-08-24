// Package storage generates presigned upload URLs against a MinIO (or any
// S3-compatible) object store. The backend never proxies the upload bytes
// themselves -- the client PUTs directly to the presigned URL, so the
// client this package is configured with must be reachable by the client
// (public endpoint), not necessarily the same host the backend itself
// would use for server-side calls.
package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	apperrors "github.com/preuni/pkg/errors"
)

// PresignExpiry bounds how long a presigned PUT URL remains valid.
const PresignExpiry = 15 * time.Minute

// Client wraps a minio.Client scoped to one bucket.
type Client struct {
	mc     *minio.Client
	bucket string
}

// Config carries the connection details for the object store's
// client-reachable (public) endpoint.
type Config struct {
	Endpoint  string // host:port or host, no scheme
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// region is fixed (not user-configurable) purely to make presigning a
// local, network-free operation: without an explicit Region, minio-go's
// PresignedPutObject calls GetBucketLocation over the network on first
// use per bucket. MinIO's own default region is "us-east-1" regardless
// of where it's actually hosted, so this has no real geographic meaning.
const region = "us-east-1"

// New constructs a Client. Endpoint must be the address the uploading
// client (mobile app / browser) can actually reach -- not necessarily an
// internal Docker service name.
func New(cfg Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: constructing minio client: %w", err)
	}
	return &Client{mc: mc, bucket: cfg.Bucket}, nil
}

// PresignedPutURL returns a URL the client can PUT the object bytes to
// directly, valid for PresignExpiry.
func (c *Client) PresignedPutURL(ctx context.Context, objectKey string) (string, error) {
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, objectKey, PresignExpiry)
	if err != nil {
		return "", apperrors.Internal(fmt.Errorf("storage: presigning put url: %w", err))
	}
	return u.String(), nil
}

// PublicURL returns the URL the object is reachable at for reading, once
// uploaded -- the bucket must be configured for public read (see
// infra/docker-compose*.yml's minio-init sidecar, which sets the bucket's
// anonymous-read policy on first boot).
func (c *Client) PublicURL(objectKey string) string {
	u := &url.URL{
		Scheme: c.mc.EndpointURL().Scheme,
		Host:   c.mc.EndpointURL().Host,
		Path:   "/" + c.bucket + "/" + objectKey,
	}
	return u.String()
}
