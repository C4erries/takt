package workspace

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := NewUser("Test@Example.com", "hashed_secret", "Pavel")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, "Pavel", u.DisplayName)
	assert.Nil(t, u.AvatarURL)
	assert.NotEqual(t, uuid.Nil, u.ID)
	assert.False(t, u.CreatedAt.IsZero())
}

func TestUser_UpdateProfile(t *testing.T) {
	u, err := NewUser("user@example.com", "hash", "Initial")
	require.NoError(t, err)

	t.Run("valid display name update", func(t *testing.T) {
		newName := "Updated Name"
		err := u.UpdateProfile(&newName)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", u.DisplayName)
	})

	t.Run("nil display name does nothing", func(t *testing.T) {
		err := u.UpdateProfile(nil)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", u.DisplayName)
	})

	t.Run("empty display name returns error", func(t *testing.T) {
		empty := "   "
		err := u.UpdateProfile(&empty)
		require.ErrorIs(t, err, ErrInvalidDisplayName)
	})

	t.Run("display name exceeding 100 chars returns error", func(t *testing.T) {
		tooLong := strings.Repeat("a", 101)
		err := u.UpdateProfile(&tooLong)
		require.ErrorIs(t, err, ErrInvalidDisplayName)
	})
}

func TestUser_SetAvatar(t *testing.T) {
	u, err := NewUser("user@example.com", "hash", "Initial")
	require.NoError(t, err)

	t.Run("set valid avatar url", func(t *testing.T) {
		u.SetAvatar("https://cdn.example.com/avatar.png")
		require.NotNil(t, u.AvatarURL)
		assert.Equal(t, "https://cdn.example.com/avatar.png", *u.AvatarURL)
	})

	t.Run("empty avatar url clears avatar", func(t *testing.T) {
		u.SetAvatar("   ")
		assert.Nil(t, u.AvatarURL)
	})
}

func TestNewUser_InvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "empty email", email: ""},
		{name: "missing at sign", email: "invalid-email"},
		{name: "only spaces", email: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(tt.email, "hashed_secret", "Pavel")
			assert.ErrorIs(t, err, ErrInvalidEmail)
		})
	}
}

func TestNewUser_InvalidDisplayName(t *testing.T) {
	tests := []struct {
		name string
		nameVal string
	}{
		{name: "empty name", nameVal: ""},
		{name: "spaces only", nameVal: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser("user@example.com", "hashed_secret", tt.nameVal)
			assert.ErrorIs(t, err, ErrInvalidDisplayName)
		})
	}
}

func TestNewPersonalWorkspace(t *testing.T) {
	ownerID := uuid.New()
	ws, err := NewPersonalWorkspace(ownerID, "")
	require.NoError(t, err)

	assert.Equal(t, "Личное", ws.Name)
	assert.Equal(t, TypePersonal, ws.Type)
	assert.Equal(t, ownerID, ws.OwnerID)
	assert.NotEqual(t, uuid.Nil, ws.ID)
}

func TestNewWorkspace(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name        string
		wsName      string
		wantErr     error
		expectedName string
	}{
		{
			name:         "valid workspace name",
			wsName:       "Команда Продукта",
			wantErr:      nil,
			expectedName: "Команда Продукта",
		},
		{
			name:         "trimmed whitespace",
			wsName:       "   Takt Dev   ",
			wantErr:      nil,
			expectedName: "Takt Dev",
		},
		{
			name:    "empty workspace name",
			wsName:  "",
			wantErr: ErrInvalidWorkspaceName,
		},
		{
			name:    "whitespace only",
			wsName:  "    ",
			wantErr: ErrInvalidWorkspaceName,
		},
		{
			name:    "name exceeding 255 chars",
			wsName:  strings.Repeat("a", 256),
			wantErr: ErrInvalidWorkspaceName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := NewWorkspace(ownerID, tt.wsName)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, ws)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ws)
				assert.Equal(t, tt.expectedName, ws.Name)
				assert.Equal(t, ownerID, ws.OwnerID)
				assert.NotEqual(t, uuid.Nil, ws.ID)
				assert.False(t, ws.CreatedAt.IsZero())
			}
		})
	}
}

func TestNewWorkspaceMember(t *testing.T) {
	wsID := uuid.New()
	userID := uuid.New()
	member, err := NewWorkspaceMember(wsID, userID, RoleOwner)
	require.NoError(t, err)

	assert.Equal(t, wsID, member.WorkspaceID)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, RoleOwner, member.Role)
	assert.NotEqual(t, uuid.Nil, member.ID)
}

func TestNewWorkspaceInvite(t *testing.T) {
	wsID := uuid.New()
	invitedByID := uuid.New()
	future := time.Now().UTC().Add(7 * 24 * time.Hour)

	tests := []struct {
		name      string
		token     string
		expiresAt *time.Time
		wantErr   error
	}{
		{
			name:      "valid invite with expiration",
			token:     "valid-secret-token-12345678",
			expiresAt: &future,
			wantErr:   nil,
		},
		{
			name:      "valid invite without expiration",
			token:     "permanent-token-abcdef",
			expiresAt: nil,
			wantErr:   nil,
		},
		{
			name:      "empty token error",
			token:     "",
			expiresAt: nil,
			wantErr:   ErrInvalidInviteToken,
		},
		{
			name:      "whitespace only token error",
			token:     "   ",
			expiresAt: nil,
			wantErr:   ErrInvalidInviteToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invite, err := NewWorkspaceInvite(wsID, invitedByID, tt.token, tt.expiresAt)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, invite)
			} else {
				require.NoError(t, err)
				require.NotNil(t, invite)
				assert.Equal(t, wsID, invite.WorkspaceID)
				assert.Equal(t, invitedByID, invite.InvitedByID)
				assert.Equal(t, strings.TrimSpace(tt.token), invite.Token)
				assert.Equal(t, tt.expiresAt, invite.ExpiresAt)
				assert.NotEqual(t, uuid.Nil, invite.ID)
			}
		})
	}
}

func TestWorkspaceInvite_IsExpired(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{
			name:      "nil expiration never expires",
			expiresAt: nil,
			want:      false,
		},
		{
			name:      "future expiration is not expired",
			expiresAt: &future,
			want:      false,
		},
		{
			name:      "past expiration is expired",
			expiresAt: &past,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invite := &WorkspaceInvite{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, invite.IsExpired())
		})
	}
}
