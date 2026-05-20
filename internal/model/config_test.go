package model

import "testing"

func TestSensitiveConfigKeyIncludesCustomAPIKey(t *testing.T) {
	if !isSensitiveConfigKey("custom_api_key") {
		t.Fatal("expected custom_api_key to be treated as sensitive")
	}
	if isSensitiveConfigKey("custom_model") {
		t.Fatal("expected custom_model to be treated as non-sensitive")
	}
}

func TestEncryptDecryptConfigValueIfSensitive(t *testing.T) {
	encryptionKey := "12345678901234567890123456789012"
	plainText := "custom-secret-token"

	encryptedValue, err := encryptConfigValueIfSensitive("custom_api_key", plainText, encryptionKey)
	if err != nil {
		t.Fatalf("encryptConfigValueIfSensitive returned error: %v", err)
	}
	if encryptedValue == plainText {
		t.Fatal("expected encrypted value to differ from plain text")
	}

	decryptedValue, err := decryptConfigValueIfSensitive("custom_api_key", encryptedValue, encryptionKey)
	if err != nil {
		t.Fatalf("decryptConfigValueIfSensitive returned error: %v", err)
	}
	if decryptedValue != plainText {
		t.Fatalf("decrypted value = %q, want %q", decryptedValue, plainText)
	}
}

func TestEncryptConfigValueIfSensitiveSkipsNonSensitiveKeys(t *testing.T) {
	plainText := "gpt-4.1-mini"
	value, err := encryptConfigValueIfSensitive("custom_model", plainText, "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("encryptConfigValueIfSensitive returned error: %v", err)
	}
	if value != plainText {
		t.Fatalf("non-sensitive value = %q, want %q", value, plainText)
	}
}
