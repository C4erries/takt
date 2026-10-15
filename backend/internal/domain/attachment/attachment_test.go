package attachment_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"takt/internal/domain/attachment"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean simple filename",
			input:    "report.pdf",
			expected: "report.pdf",
		},
		{
			name:     "trims whitespace",
			input:    "  annual_report.xlsx  ",
			expected: "annual_report.xlsx",
		},
		{
			name:     "unix path traversal stripped",
			input:    "../../../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "windows path traversal stripped",
			input:    `..\..\windows\system32\calc.exe`,
			expected: "calc.exe",
		},
		{
			name:     "replaces quotes and semicolons",
			input:    `image"; evil="payload.png`,
			expected: "image__ evil=_payload.png",
		},
		{
			name:     "replaces newlines and carriage returns",
			input:    "file\r\nSet-Cookie:malicious.txt",
			expected: "file__Set-Cookie:malicious.txt",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only returns empty",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "pure traversal dots return empty",
			input:    "../../..",
			expected: "",
		},
		{
			name:     "truncates long name while preserving extension",
			input:    strings.Repeat("a", 300) + ".png",
			expected: strings.Repeat("a", 251) + ".png",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := attachment.SanitizeFileName(tc.input)
			assert.Equal(t, tc.expected, actual)
			if actual != "" {
				assert.LessOrEqual(t, len([]rune(actual)), attachment.MaxFileNameLen)
			}
		})
	}
}

func TestGenerateS3Key(t *testing.T) {
	workspaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	taskID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	attachmentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	fileName := "screenshot.png"

	expected := "workspaces/11111111-1111-1111-1111-111111111111/tasks/22222222-2222-2222-2222-222222222222/33333333-3333-3333-3333-333333333333_screenshot.png"
	actual := attachment.GenerateS3Key(workspaceID, taskID, attachmentID, fileName)

	assert.Equal(t, expected, actual)
}

func TestNewAttachment(t *testing.T) {
	validWorkspaceID := uuid.New()
	validTaskID := uuid.New()
	validCreatedByID := uuid.New()

	tests := []struct {
		name          string
		workspaceID   uuid.UUID
		taskID        uuid.UUID
		createdByID   uuid.UUID
		fileName      string
		fileSize      int64
		contentType   string
		maxSizeBytes  int64
		expectedErr   error
		checkValidate func(t *testing.T, att *attachment.Attachment)
	}{
		{
			name:         "successful creation with explicit content type",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "photo.jpg",
			fileSize:     1024,
			contentType:  "image/jpeg",
			maxSizeBytes: 2048,
			expectedErr:  nil,
			checkValidate: func(t *testing.T, att *attachment.Attachment) {
				require.NotNil(t, att)
				assert.NotEqual(t, uuid.Nil, att.ID)
				assert.Equal(t, validWorkspaceID, att.WorkspaceID)
				assert.Equal(t, validTaskID, att.TaskID)
				assert.Equal(t, validCreatedByID, att.CreatedByID)
				assert.Equal(t, "photo.jpg", att.FileName)
				assert.Equal(t, int64(1024), att.FileSize)
				assert.Equal(t, "image/jpeg", att.ContentType)
				assert.NotEmpty(t, att.S3Key)
				assert.False(t, att.CreatedAt.IsZero())
			},
		},
		{
			name:         "empty content type defaults to application/octet-stream",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "archive.bin",
			fileSize:     512,
			contentType:  "   ",
			maxSizeBytes: 1024,
			expectedErr:  nil,
			checkValidate: func(t *testing.T, att *attachment.Attachment) {
				require.NotNil(t, att)
				assert.Equal(t, "application/octet-stream", att.ContentType)
			},
		},
		{
			name:         "file size exactly at max size limit",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "data.csv",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 1024,
			expectedErr:  nil,
			checkValidate: func(t *testing.T, att *attachment.Attachment) {
				require.NotNil(t, att)
				assert.Equal(t, int64(1024), att.FileSize)
			},
		},
		{
			name:         "error nil workspace ID",
			workspaceID:  uuid.Nil,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "data.csv",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidWorkspaceID,
		},
		{
			name:         "error nil task ID",
			workspaceID:  validWorkspaceID,
			taskID:       uuid.Nil,
			createdByID:  validCreatedByID,
			fileName:     "data.csv",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidTaskID,
		},
		{
			name:         "error nil created by ID",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  uuid.Nil,
			fileName:     "data.csv",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidCreatedByID,
		},
		{
			name:         "error empty file name",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "   ",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidFileName,
		},
		{
			name:         "error file name with only traversal characters",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "../../",
			fileSize:     1024,
			contentType:  "text/csv",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidFileName,
		},
		{
			name:         "error zero file size",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "empty.txt",
			fileSize:     0,
			contentType:  "text/plain",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidFileSize,
		},
		{
			name:         "error negative file size",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "negative.txt",
			fileSize:     -100,
			contentType:  "text/plain",
			maxSizeBytes: 2048,
			expectedErr:  attachment.ErrInvalidFileSize,
		},
		{
			name:         "error file size exceeding maxSizeBytes",
			workspaceID:  validWorkspaceID,
			taskID:       validTaskID,
			createdByID:  validCreatedByID,
			fileName:     "huge.iso",
			fileSize:     52428801,
			contentType:  "application/octet-stream",
			maxSizeBytes: 52428800,
			expectedErr:  attachment.ErrFileTooLarge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			att, err := attachment.NewAttachment(
				tc.workspaceID,
				tc.taskID,
				tc.createdByID,
				tc.fileName,
				tc.fileSize,
				tc.contentType,
				tc.maxSizeBytes,
			)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, att)
			} else {
				require.NoError(t, err)
				require.NotNil(t, att)
				if tc.checkValidate != nil {
					tc.checkValidate(t, att)
				}
			}
		})
	}
}
