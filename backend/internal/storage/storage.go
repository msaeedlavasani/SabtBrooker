package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// FileStorage wraps MinIO operations
type FileStorage struct {
	client     *minio.Client
	bucketName string
}

// NewFileStorage creates a new MinIO client and ensures bucket exists
func NewFileStorage(ctx context.Context, cfg config.MinIOConfig) (*FileStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	fs := &FileStorage{
		client:     client,
		bucketName: cfg.Bucket,
	}

	// Ensure bucket exists
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket %s: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
		}
		slog.Info("created MinIO bucket", "bucket", cfg.Bucket)
	}

	slog.Info("connected to MinIO", "endpoint", cfg.Endpoint, "bucket", cfg.Bucket)
	return fs, nil
}

// Upload stores a file in MinIO and returns its object key
func (fs *FileStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := fs.client.PutObject(ctx, fs.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %w", objectKey, err)
	}
	slog.Info("file uploaded", "key", objectKey, "size", size)
	return nil
}

// Download retrieves a file from MinIO
func (fs *FileStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := fs.client.GetObject(ctx, fs.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", objectKey, err)
	}
	return obj, nil
}

// Delete removes a file from MinIO
func (fs *FileStorage) Delete(ctx context.Context, objectKey string) error {
	err := fs.client.RemoveObject(ctx, fs.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", objectKey, err)
	}
	slog.Info("file deleted", "key", objectKey)
	return nil
}

// GeneratePresignedURL creates a time-limited download URL
func (fs *FileStorage) GeneratePresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := fs.client.PresignedGetObject(ctx, fs.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for %s: %w", objectKey, err)
	}
	return url.String(), nil
}

// PresignedUploadURL creates a time-limited upload URL
func (fs *FileStorage) PresignedUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := fs.client.PresignedPutObject(ctx, fs.bucketName, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL for %s: %w", objectKey, err)
	}
	return url.String(), nil
}

// Close is a no-op — MinIO client doesn't need explicit close
func (fs *FileStorage) Close() error {
	return nil
}
