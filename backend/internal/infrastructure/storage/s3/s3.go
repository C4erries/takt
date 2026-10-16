package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3Service "github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3ClientAPI abstracts AWS S3 client operations for testing and decoupling.
type S3ClientAPI interface {
	PutObject(ctx context.Context, params *s3Service.PutObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3Service.GetObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3Service.DeleteObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.DeleteObjectOutput, error)
	HeadBucket(ctx context.Context, params *s3Service.HeadBucketInput, optFns ...func(*s3Service.Options)) (*s3Service.HeadBucketOutput, error)
	CreateBucket(ctx context.Context, params *s3Service.CreateBucketInput, optFns ...func(*s3Service.Options)) (*s3Service.CreateBucketOutput, error)
}

// Config contains connection parameters for S3 or MinIO storage.
type Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// Storage provides object storage operations using AWS S3 client.
type Storage struct {
	client S3ClientAPI
	bucket string
}

// NewStorage initializes a new Storage instance connected to S3 or MinIO.
func NewStorage(cfg Config) (*Storage, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               cfg.Endpoint,
			SigningRegion:     cfg.Region,
			HostnameImmutable: true,
		}, nil
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3Service.NewFromConfig(awsCfg, func(o *s3Service.Options) {
		o.UsePathStyle = true
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// NewStorageWithClient creates a Storage instance using an injected S3ClientAPI (e.g. for testing).
func NewStorageWithClient(client S3ClientAPI, bucket string) *Storage {
	return &Storage{
		client: client,
		bucket: bucket,
	}
}

// PutObject streams data into the S3 bucket with the given key, content size and content type.
func (s *Storage) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	input := &s3Service.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("failed to put object to s3: %w", err)
	}
	return nil
}

// GetObject downloads the object content from the S3 bucket.
func (s *Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	input := &s3Service.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get object from s3: %w", err)
	}

	contentType := "application/octet-stream"
	if output.ContentType != nil && *output.ContentType != "" {
		contentType = *output.ContentType
	}

	var size int64
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	return output.Body, contentType, size, nil
}

// DeleteObject deletes an object by key from the S3 bucket.
func (s *Storage) DeleteObject(ctx context.Context, key string) error {
	input := &s3Service.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	if _, err := s.client.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("failed to delete object from s3: %w", err)
	}
	return nil
}

// EnsureBucket checks if the configured bucket exists, and creates it if not.
// Retries with a short backoff in case S3/MinIO is still initializing during startup.
func (s *Storage) EnsureBucket(ctx context.Context) error {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		_, err := s.client.HeadBucket(ctx, &s3Service.HeadBucketInput{
			Bucket: aws.String(s.bucket),
		})
		if err == nil {
			return nil
		}

		_, err = s.client.CreateBucket(ctx, &s3Service.CreateBucketInput{
			Bucket: aws.String(s.bucket),
		})
		if err == nil {
			return nil
		}

		errStr := err.Error()
		if strings.Contains(errStr, "BucketAlreadyOwnedByYou") || strings.Contains(errStr, "BucketAlreadyExists") {
			return nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
	return fmt.Errorf("failed to ensure s3 bucket %s: %w", s.bucket, lastErr)
}

