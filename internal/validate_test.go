package flagapi

import "testing"

func TestNormalizeUserInputRejectsInvalidEmail(t *testing.T) {
	_, err := normalizeUserInput(createUserRequest{
		Name:    "Test User",
		Email:   "invalid-email",
		Country: "IN",
	})
	if err == nil || err.Error() != "email is invalid" {
		t.Fatalf("expected invalid email error, got %v", err)
	}
}

func TestNormalizeUserInputAcceptsValidEmail(t *testing.T) {
	user, err := normalizeUserInput(createUserRequest{
		Name:    "Test User",
		Email:   "test.user+1@example.co.in",
		Country: "IN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test.user+1@example.co.in" {
		t.Fatalf("unexpected normalized email: %s", user.Email)
	}
}
