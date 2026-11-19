package model

import (
	"mime/multipart"
	"testing"
)

func TestValidateImageFileRejectsSVG(t *testing.T) {
	if err := ValidateImageFile(&multipart.FileHeader{Filename: "evil.svg", Size: 12}); err == nil {
		t.Fatal("expected svg to be rejected")
	}
	if err := ValidateImageFile(&multipart.FileHeader{Filename: "ok.png", Size: 12}); err != nil {
		t.Fatalf("png should be allowed: %v", err)
	}
}
