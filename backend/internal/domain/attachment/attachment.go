package attachment

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultMaxAttachmentSizeBytes is 50 MB.
	DefaultMaxAttachmentSizeBytes int64 = 52428800
	// MaxFileNameLen is the maximum allowed length of sanitized file name.
	MaxFileNameLen = 255
)

// Attachment represents a file uploaded to S3 and linked to a task.
type Attachment struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	TaskID      uuid.UUID
	FileName    string
	FileSize    int64
	ContentType string
	S3Key       string
	CreatedByID uuid.UUID
	CreatedAt   time.Time
}

// SanitizeFileName cleans the file name from directory traversal patterns,
// quotes, control characters, and truncates it to MaxFileNameLen.
func SanitizeFileName(fileName string) string {
	// Normalize path separators
	cleaned := strings.ReplaceAll(fileName, "\\", "/")

	// Extract base name
	parts := strings.Split(cleaned, "/")
	base := ""
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p != "" && p != "." && p != ".." {
			base = p
			break
		}
	}
	if base == "" {
		return ""
	}

	// Filter out dangerous or control characters
	var sb strings.Builder
	for _, r := range base {
		if r < 32 || r == 127 || r == '"' || r == ';' || r == '/' || r == '\\' {
			sb.WriteRune('_')
		} else {
			sb.WriteRune(r)
		}
	}

	sanitized := strings.TrimSpace(sb.String())
	if sanitized == "" {
		return ""
	}

	// Truncate to MaxFileNameLen preserving extension if possible
	runes := []rune(sanitized)
	if len(runes) > MaxFileNameLen {
		ext := path.Ext(sanitized)
		extRunes := []rune(ext)
		if len(extRunes) < MaxFileNameLen {
			maxBase := MaxFileNameLen - len(extRunes)
			sanitized = string(runes[:maxBase]) + ext
		} else {
			sanitized = string(runes[:MaxFileNameLen])
		}
	}

	return sanitized
}

// GenerateS3Key constructs a canonical storage key in the S3 bucket.
func GenerateS3Key(workspaceID, taskID, attachmentID uuid.UUID, fileName string) string {
	return fmt.Sprintf("workspaces/%s/tasks/%s/%s_%s", workspaceID, taskID, attachmentID, fileName)
}

// NewAttachment validates domain invariants and creates a new Attachment entity.
func NewAttachment(
	workspaceID uuid.UUID,
	taskID uuid.UUID,
	createdByID uuid.UUID,
	fileName string,
	fileSize int64,
	contentType string,
	maxSizeBytes int64,
) (*Attachment, error) {
	if workspaceID == uuid.Nil {
		return nil, ErrInvalidWorkspaceID
	}
	if taskID == uuid.Nil {
		return nil, ErrInvalidTaskID
	}
	if createdByID == uuid.Nil {
		return nil, ErrInvalidCreatedByID
	}

	sanitizedName := SanitizeFileName(fileName)
	if sanitizedName == "" {
		return nil, ErrInvalidFileName
	}

	if fileSize <= 0 {
		return nil, ErrInvalidFileSize
	}

	if maxSizeBytes > 0 && fileSize > maxSizeBytes {
		return nil, ErrFileTooLarge
	}

	trimmedContentType := strings.TrimSpace(contentType)
	if trimmedContentType == "" {
		trimmedContentType = "application/octet-stream"
	}

	attachmentID := uuid.New()
	s3Key := GenerateS3Key(workspaceID, taskID, attachmentID, sanitizedName)

	return &Attachment{
		ID:          attachmentID,
		WorkspaceID: workspaceID,
		TaskID:      taskID,
		FileName:    sanitizedName,
		FileSize:    fileSize,
		ContentType: trimmedContentType,
		S3Key:       s3Key,
		CreatedByID: createdByID,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

