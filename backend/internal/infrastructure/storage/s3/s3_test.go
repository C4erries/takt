package s3_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3Service "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/infrastructure/storage/s3"
)

type mockS3Client struct {
	putObjectFunc    func(ctx context.Context, params *s3Service.PutObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.PutObjectOutput, error)
	getObjectFunc    func(ctx context.Context, params *s3Service.GetObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.GetObjectOutput, error)
	deleteObjectFunc func(ctx context.Context, params *s3Service.DeleteObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.DeleteObjectOutput, error)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3Service.PutObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return &s3Service.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3Service.GetObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.GetObjectOutput, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, params, optFns...)
	}
	return &s3Service.GetObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3Service.DeleteObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.DeleteObjectOutput, error) {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, params, optFns...)
	}
	return &s3Service.DeleteObjectOutput{}, nil
}

func TestStorage_PutObject(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		content     string
		contentType string
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful put",
			key:         "workspaces/1/tasks/2/3_file.txt",
			content:     "hello s3",
			contentType: "text/plain",
			mockErr:     nil,
			wantErr:     false,
		},
		{
			name:        "s3 client error",
			key:         "workspaces/1/tasks/2/3_file.txt",
			content:     "hello s3",
			contentType: "text/plain",
			mockErr:     errors.New("s3 connection timeout"),
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockS3Client{
				putObjectFunc: func(ctx context.Context, params *s3Service.PutObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.PutObjectOutput, error) {
					assert.Equal(t, "test-bucket", *params.Bucket)
					assert.Equal(t, tc.key, *params.Key)
					assert.Equal(t, tc.contentType, *params.ContentType)
					return &s3Service.PutObjectOutput{}, tc.mockErr
				},
			}

			storage := s3.NewStorageWithClient(mockClient, "test-bucket")
			err := storage.PutObject(context.Background(), tc.key, strings.NewReader(tc.content), int64(len(tc.content)), tc.contentType)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStorage_GetObject(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedBody string
		contentType string
		contentLen  int64
		mockErr     error
		wantErr     bool
	}{
		{
			name:         "successful get",
			key:          "workspaces/1/tasks/2/3_image.png",
			expectedBody: "binary_image_data",
			contentType:  "image/png",
			contentLen:   17,
			mockErr:      nil,
			wantErr:      false,
		},
		{
			name:    "s3 not found error",
			key:     "workspaces/1/tasks/2/missing.png",
			mockErr: errors.New("NoSuchKey"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockS3Client{
				getObjectFunc: func(ctx context.Context, params *s3Service.GetObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.GetObjectOutput, error) {
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return &s3Service.GetObjectOutput{
						Body:          io.NopCloser(bytes.NewReader([]byte(tc.expectedBody))),
						ContentType:   aws.String(tc.contentType),
						ContentLength: aws.Int64(tc.contentLen),
					}, nil
				},
			}

			storage := s3.NewStorageWithClient(mockClient, "test-bucket")
			reader, contentType, size, err := storage.GetObject(context.Background(), tc.key)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, reader)
			} else {
				require.NoError(t, err)
				require.NotNil(t, reader)
				defer reader.Close()

				data, readErr := io.ReadAll(reader)
				require.NoError(t, readErr)
				assert.Equal(t, tc.expectedBody, string(data))
				assert.Equal(t, tc.contentType, contentType)
				assert.Equal(t, tc.contentLen, size)
			}
		})
	}
}

func TestStorage_DeleteObject(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			key:     "workspaces/1/tasks/2/3_file.txt",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "s3 delete error",
			key:     "workspaces/1/tasks/2/3_file.txt",
			mockErr: errors.New("access denied"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockS3Client{
				deleteObjectFunc: func(ctx context.Context, params *s3Service.DeleteObjectInput, optFns ...func(*s3Service.Options)) (*s3Service.DeleteObjectOutput, error) {
					assert.Equal(t, "test-bucket", *params.Bucket)
					assert.Equal(t, tc.key, *params.Key)
					return &s3Service.DeleteObjectOutput{}, tc.mockErr
				},
			}

			storage := s3.NewStorageWithClient(mockClient, "test-bucket")
			err := storage.DeleteObject(context.Background(), tc.key)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

