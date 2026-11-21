package auth

import (
	"testing"
	"time"

	"zbxtable/internal/model"

	"github.com/golang-jwt/jwt/v4"
)

func TestIssueAndValidateToken(t *testing.T) {
	cleanup := model.SetEncryptionKeyForTest("test-install-jwt-secret-key-32bytes!!")
	defer cleanup()

	exp := time.Now().Add(time.Hour).Unix()
	tok, err := IssueToken("admin", exp)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	ok, user, err := ValidateToken(tok)
	if err != nil || !ok || user != "admin" {
		t.Fatalf("ValidateToken got ok=%v user=%q err=%v", ok, user, err)
	}
}

func TestValidateTokenRejectsNonHMAC(t *testing.T) {
	cleanup := model.SetEncryptionKeyForTest("test-install-jwt-secret-key-32bytes!!")
	defer cleanup()

	// alg=none 伪令牌
	noneTok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJpc3MiOiJhZG1pbiIsImV4cCI6OTk5OTk5OTk5OX0."
	ok, _, err := ValidateToken(noneTok)
	if ok || err == nil {
		t.Fatalf("expected reject none alg, got ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	cleanup := model.SetEncryptionKeyForTest("test-install-jwt-secret-key-32bytes!!")
	defer cleanup()
	tok, err := IssueToken("admin", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	cleanup2 := model.SetEncryptionKeyForTest("another-install-jwt-secret-key-32b!!")
	defer cleanup2()
	ok, _, err := ValidateToken(tok)
	if ok || err == nil {
		t.Fatalf("expected reject wrong secret, got ok=%v err=%v", ok, err)
	}
}

func TestValidateTokenRejectsDefaultEncryptionKey(t *testing.T) {
	cleanup := model.SetEncryptionKeyForTest(defaultEncryptionKey)
	defer cleanup()
	_, err := IssueToken("admin", time.Now().Add(time.Hour).Unix())
	if err == nil {
		t.Fatal("expected IssueToken to reject default encryption key")
	}
}

func TestValidateTokenRejectsRSAAlgHeader(t *testing.T) {
	cleanup := model.SetEncryptionKeyForTest("test-install-jwt-secret-key-32bytes!!")
	defer cleanup()
	claims := &jwt.StandardClaims{Issuer: "admin", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, _ := token.SigningString()
	ok, _, err := ValidateToken(s + ".x")
	if ok || err == nil {
		t.Fatalf("expected reject RS256-ish token, got ok=%v err=%v", ok, err)
	}
}
