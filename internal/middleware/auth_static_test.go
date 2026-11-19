package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJWTAuthUnlessAllowsPublicUploadPrefixes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		path       string
		wantStatus int
	}{
		{"/upload/logo/a.png", http.StatusOK},
		{"/upload/background/b.png", http.StatusOK},
		{"/upload/other/c.png", http.StatusUnauthorized},
		{"/download/report.xlsx", http.StatusUnauthorized},
	}

	for _, tc := range cases {
		r := gin.New()
		if tc.path == "/download/report.xlsx" {
			r.GET("/download/*filepath", JWTAuthMiddleware(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
		} else {
			r.GET("/upload/*filepath", JWTAuthUnless("/upload/logo", "/upload/background"), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
		}

		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.wantStatus {
			t.Fatalf("path %s: status = %d, want %d", tc.path, w.Code, tc.wantStatus)
		}
	}
}
