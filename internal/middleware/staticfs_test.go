package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNoListDirBlocksListingButServesFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "report.txt")
	if err := os.WriteFile(filePath, []byte("secret-report"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := http.FileServer(NoListDir(dir))

	listResp := httptest.NewRecorder()
	h.ServeHTTP(listResp, httptest.NewRequest(http.MethodGet, "/", nil))
	if listResp.Code != http.StatusNotFound {
		t.Fatalf("directory listing status = %d, want 404", listResp.Code)
	}

	fileResp := httptest.NewRecorder()
	h.ServeHTTP(fileResp, httptest.NewRequest(http.MethodGet, "/report.txt", nil))
	if fileResp.Code != http.StatusOK {
		t.Fatalf("file status = %d, want 200", fileResp.Code)
	}
	body, _ := io.ReadAll(fileResp.Body)
	if string(body) != "secret-report" {
		t.Fatalf("file body = %q", body)
	}
}
