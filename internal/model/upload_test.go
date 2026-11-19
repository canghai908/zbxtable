package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteBackgroundImageRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(prev) }()

	if err := os.MkdirAll("upload/background", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("secret.txt", []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}

	err = DeleteBackgroundImage("/upload/background/../../secret.txt")
	if err == nil {
		t.Fatal("expected error for traversal/missing basename under upload dir")
	}
	if _, err := os.Stat("secret.txt"); err != nil {
		t.Fatalf("outside secret file must remain: %v", err)
	}
}

func TestDeleteBackgroundImageDeletesBasenameOnly(t *testing.T) {
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(prev) }()

	if err := os.MkdirAll("upload/background", 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join("upload", "background", "ok.png")
	if err := os.WriteFile(target, []byte("img"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := DeleteBackgroundImage("/upload/background/ok.png"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err=%v", err)
	}
}

func TestSafePathUnder(t *testing.T) {
	tmp := t.TempDir()
	ok, err := safePathUnder(tmp, "a.png")
	if err != nil {
		t.Fatal(err)
	}
	wantAbs, _ := filepath.Abs(filepath.Join(tmp, "a.png"))
	if ok != wantAbs {
		t.Fatalf("got %q want %q", ok, wantAbs)
	}

	if _, err := safePathUnder(tmp, filepath.Join("..", "outside.txt")); err == nil {
		t.Fatal("expected escape to fail")
	}
}
