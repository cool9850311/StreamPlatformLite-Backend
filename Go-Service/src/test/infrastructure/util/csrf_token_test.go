package infrastructure

import (
	"Go-Service/src/main/infrastructure/util"
	"strings"
	"testing"
)

func TestGenerateCsrfToken_HasNonceDotSigFormat(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		t.Fatalf("expected token to contain '.', got %q", token)
	}
	if parts[0] == "" || parts[1] == "" {
		t.Fatalf("expected both parts to be non-empty, got %q", token)
	}
}

func TestGenerateCsrfToken_ProducesUniqueTokens(t *testing.T) {
	token1, err1 := util.GenerateCsrfToken("secret", "user-1")
	token2, err2 := util.GenerateCsrfToken("secret", "user-1")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if token1 == token2 {
		t.Fatalf("expected unique tokens, got identical: %q", token1)
	}
}

func TestValidateCsrfToken_Valid(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !util.ValidateCsrfToken(token, "secret", "user-1") {
		t.Fatal("expected valid token to pass validation")
	}
}

func TestValidateCsrfToken_WrongSecret(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret-A", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if util.ValidateCsrfToken(token, "secret-B", "user-1") {
		t.Fatal("expected wrong secret to fail validation")
	}
}

func TestValidateCsrfToken_WrongUserID(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if util.ValidateCsrfToken(token, "secret", "user-2") {
		t.Fatal("expected wrong userID to fail validation")
	}
}

func TestValidateCsrfToken_TamperedNonce(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.SplitN(token, ".", 2)
	tampered := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + "." + parts[1]
	if util.ValidateCsrfToken(tampered, "secret", "user-1") {
		t.Fatal("expected tampered nonce to fail validation")
	}
}

func TestValidateCsrfToken_TamperedSignature(t *testing.T) {
	token, err := util.GenerateCsrfToken("secret", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.SplitN(token, ".", 2)
	tampered := parts[0] + "." + strings.Repeat("b", 64)
	if util.ValidateCsrfToken(tampered, "secret", "user-1") {
		t.Fatal("expected tampered signature to fail validation")
	}
}

func TestValidateCsrfToken_EmptyToken(t *testing.T) {
	if util.ValidateCsrfToken("", "secret", "user-1") {
		t.Fatal("expected empty token to fail validation")
	}
}

func TestValidateCsrfToken_NoDot(t *testing.T) {
	if util.ValidateCsrfToken("nodottoken", "secret", "user-1") {
		t.Fatal("expected token without dot to fail validation")
	}
}

func TestValidateCsrfToken_OnlyDot(t *testing.T) {
	if util.ValidateCsrfToken(".", "secret", "user-1") {
		t.Fatal("expected token with only dot to fail validation")
	}
}
