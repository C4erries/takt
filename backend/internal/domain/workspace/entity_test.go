package workspace

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := NewUser("Test@Example.com", "hashed_secret", "Pavel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Email != "test@example.com" {
		t.Errorf("expected lowercase email test@example.com, got %s", u.Email)
	}
	if u.DisplayName != "Pavel" {
		t.Errorf("expected displayName Pavel, got %s", u.DisplayName)
	}
	if u.ID == uuid.Nil {
		t.Errorf("expected non-nil UUID")
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := NewUser("invalid-email", "hashed_secret", "Pavel")
	if err != ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}

	_, err = NewUser("", "hashed_secret", "Pavel")
	if err != ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestNewUser_InvalidDisplayName(t *testing.T) {
	_, err := NewUser("user@example.com", "hashed_secret", "   ")
	if err != ErrInvalidDisplayName {
		t.Errorf("expected ErrInvalidDisplayName, got %v", err)
	}
}

func TestNewPersonalWorkspace(t *testing.T) {
	ownerID := uuid.New()
	ws, err := NewPersonalWorkspace(ownerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ws.Name != "Личное" {
		t.Errorf("expected default name 'Личное', got %s", ws.Name)
	}
	if ws.Type != TypePersonal {
		t.Errorf("expected type personal, got %s", ws.Type)
	}
	if ws.OwnerID != ownerID {
		t.Errorf("expected ownerID %v, got %v", ownerID, ws.OwnerID)
	}
}

func TestNewWorkspaceMember(t *testing.T) {
	wsID := uuid.New()
	userID := uuid.New()
	member, err := NewWorkspaceMember(wsID, userID, RoleOwner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if member.WorkspaceID != wsID || member.UserID != userID || member.Role != RoleOwner {
		t.Errorf("member fields do not match expected values")
	}
}
