package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appAttachment "takt/internal/application/attachment"
	domainAttachment "takt/internal/domain/attachment"
	domainWorkspace "takt/internal/domain/workspace"
	"takt/internal/infrastructure/http/generated"
)

type mockAttachmentService struct {
	uploadFunc  func(ctx context.Context, input appAttachment.UploadAttachmentInput) (*domainAttachment.Attachment, error)
	listFunc    func(ctx context.Context, wsID, tID, uID uuid.UUID) ([]*domainAttachment.Attachment, error)
	contentFunc func(ctx context.Context, wsID, aID, uID uuid.UUID) (*appAttachment.AttachmentContentOutput, error)
	deleteFunc  func(ctx context.Context, wsID, aID, uID uuid.UUID) error
}

func (m *mockAttachmentService) UploadAttachment(ctx context.Context, input appAttachment.UploadAttachmentInput) (*domainAttachment.Attachment, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockAttachmentService) ListTaskAttachments(ctx context.Context, wsID, tID, uID uuid.UUID) ([]*domainAttachment.Attachment, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, wsID, tID, uID)
	}
	return nil, nil
}

func (m *mockAttachmentService) GetAttachmentContent(ctx context.Context, wsID, aID, uID uuid.UUID) (*appAttachment.AttachmentContentOutput, error) {
	if m.contentFunc != nil {
		return m.contentFunc(ctx, wsID, aID, uID)
	}
	return nil, nil
}

func (m *mockAttachmentService) DeleteAttachment(ctx context.Context, wsID, aID, uID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, wsID, aID, uID)
	}
	return nil
}

func newAttachmentTestContext(method, url string, body io.Reader, contentType string, userID, wsID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, url, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(UserIDContextKey), userID)
	c.Set(string(WorkspaceIDContextKey), wsID)
	return c, rec
}

func createMultipartBody(fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile(fieldName, fileName)
	_, _ = io.Copy(part, bytes.NewReader(content))
	_ = writer.Close()
	return buf, writer.FormDataContentType()
}

func TestHandler_UploadTaskAttachment(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()
	attID := uuid.New()

	tests := []struct {
		name       string
		noService  bool
		missingFile bool
		mockErr    error
		wantStatus int
	}{
		{
			name:       "successful upload",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing multipart file",
			missingFile: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			mockErr:    appAttachment.ErrTaskNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "access denied",
			mockErr:    domainWorkspace.ErrAccessDenied,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "file too large",
			mockErr:    domainAttachment.ErrFileTooLarge,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service nil",
			noService:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockAttachmentService{
				uploadFunc: func(ctx context.Context, input appAttachment.UploadAttachmentInput) (*domainAttachment.Attachment, error) {
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return &domainAttachment.Attachment{
						ID:          attID,
						WorkspaceID: wsID,
						TaskID:      taskID,
						FileName:    "image.png",
						FileSize:    1024,
						ContentType: "image/png",
						S3Key:       "workspaces/1/tasks/2/image.png",
						CreatedByID: userID,
						CreatedAt:   time.Now().UTC(),
					}, nil
				},
			}

			var body io.Reader
			contentType := ""
			if !tc.missingFile {
				buf, ct := createMultipartBody("file", "image.png", []byte("sample_content"))
				body = buf
				contentType = ct
			} else {
				body = strings.NewReader("not-a-multipart-body")
				contentType = "text/plain"
			}

			c, rec := newAttachmentTestContext(http.MethodPost, "/attachments", body, contentType, userID, wsID)
			h := NewHandler(nil, "0.4.2", false, nil, nil, nil, nil)
			if !tc.noService {
				h.SetAttachmentService(mockSvc, 52428800)
			}

			err := h.UploadTaskAttachment(c, wsID, taskID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusCreated {
				var resp generated.Attachment
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, attID, resp.Id)
				assert.Equal(t, "image.png", resp.FileName)
				assert.Contains(t, resp.ContentUrl, attID.String())
			}
		})
	}
}

func TestHandler_ListTaskAttachments(t *testing.T) {
	wsID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		mockErr    error
		items      []*domainAttachment.Attachment
		wantStatus int
	}{
		{
			name: "successful list",
			items: []*domainAttachment.Attachment{
				{ID: uuid.New(), WorkspaceID: wsID, TaskID: taskID, FileName: "f1.pdf"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "task not found",
			mockErr:    appAttachment.ErrTaskNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "access denied",
			mockErr:    domainWorkspace.ErrAccessDenied,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockAttachmentService{
				listFunc: func(ctx context.Context, wID, tID, uID uuid.UUID) ([]*domainAttachment.Attachment, error) {
					return tc.items, tc.mockErr
				},
			}

			c, rec := newAttachmentTestContext(http.MethodGet, "/attachments", nil, "", userID, wsID)
			h := NewHandler(nil, "0.4.2", false, nil, nil, nil, nil)
			h.SetAttachmentService(mockSvc, 52428800)

			err := h.ListTaskAttachments(c, wsID, taskID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var list []generated.Attachment
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
				assert.Len(t, list, len(tc.items))
			}
		})
	}
}

func TestHandler_GetAttachmentContent(t *testing.T) {
	wsID := uuid.New()
	attID := uuid.New()
	userID := uuid.New()

	downloadTrue := true
	downloadFalse := false

	tests := []struct {
		name            string
		downloadParam   *bool
		mockErr         error
		wantStatus      int
		wantDisposition string
	}{
		{
			name:            "inline streaming",
			downloadParam:   &downloadFalse,
			wantStatus:      http.StatusOK,
			wantDisposition: `inline; filename="photo.jpg"`,
		},
		{
			name:            "download attachment streaming",
			downloadParam:   &downloadTrue,
			wantStatus:      http.StatusOK,
			wantDisposition: `attachment; filename="photo.jpg"`,
		},
		{
			name:       "not found",
			mockErr:    domainAttachment.ErrAttachmentNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "access denied",
			mockErr:    domainWorkspace.ErrAccessDenied,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockAttachmentService{
				contentFunc: func(ctx context.Context, wID, aID, uID uuid.UUID) (*appAttachment.AttachmentContentOutput, error) {
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return &appAttachment.AttachmentContentOutput{
						Body:        io.NopCloser(bytes.NewReader([]byte("jpeg_bytes"))),
						FileName:    "photo.jpg",
						ContentType: "image/jpeg",
						FileSize:    10,
					}, nil
				},
			}

			c, rec := newAttachmentTestContext(http.MethodGet, "/attachments/content", nil, "", userID, wsID)
			h := NewHandler(nil, "0.4.2", false, nil, nil, nil, nil)
			h.SetAttachmentService(mockSvc, 52428800)

			params := generated.GetAttachmentContentParams{
				Download: tc.downloadParam,
			}

			err := h.GetAttachmentContent(c, wsID, attID, params)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				assert.Equal(t, tc.wantDisposition, rec.Header().Get("Content-Disposition"))
				assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
				assert.Contains(t, rec.Header().Get("Cache-Control"), "public")
				assert.Equal(t, "jpeg_bytes", rec.Body.String())
			}
		})
	}
}

func TestHandler_DeleteAttachment(t *testing.T) {
	wsID := uuid.New()
	attID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "successful delete",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found",
			mockErr:    domainAttachment.ErrAttachmentNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "access denied",
			mockErr:    domainWorkspace.ErrAccessDenied,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "internal server error",
			mockErr:    errors.New("db failure"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockAttachmentService{
				deleteFunc: func(ctx context.Context, wID, aID, uID uuid.UUID) error {
					return tc.mockErr
				},
			}

			c, rec := newAttachmentTestContext(http.MethodDelete, "/attachments", nil, "", userID, wsID)
			h := NewHandler(nil, "0.4.2", false, nil, nil, nil, nil)
			h.SetAttachmentService(mockSvc, 52428800)

			err := h.DeleteAttachment(c, wsID, attID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

