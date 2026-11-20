package handler

import (
	stdjson "encoding/json"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuleConditionsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/rule/test", TestRuleConditions)
	for _, tt := range []struct {
		body    string
		code    int
		matched bool
	}{
		{`{"conditions":"[{\"r_type\":\"host\",\"r_func\":\"==\",\"r_value\":\"prod\"}]","data":{"host":"prod"}}`, 200, true},
		{`{"conditions":"[{\"r_type\":\"host\",\"r_func\":\"==\",\"r_value\":\"prod\"}]","data":{"host":"other"}}`, 200, false},
		{`{"conditions":"[]","data":{}}`, 422, false},
		{`{"conditions":"broken","data":{}}`, 422, false},
		{`{`, 400, false},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/rule/test", strings.NewReader(tt.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		var got struct {
			Code int `json:"code"`
			Data struct {
				Matched  bool  `json:"matched"`
				Children []any `json:"children"`
			} `json:"data"`
		}
		if err := stdjson.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.Code != tt.code || got.Data.Matched != tt.matched {
			t.Fatalf("%s: %s", tt.body, w.Body.String())
		}
		if got.Code == 200 && len(got.Data.Children) != 1 {
			t.Fatal("missing trace")
		}
	}
}

func TestRuleWritesRejectInvalidConditionsBeforeDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/rule", CreateRule)
	router.PUT("/rule/:id", UpdateRule)
	for _, method := range []string{"POST", "PUT"} {
		path := "/rule"
		if method == "PUT" {
			path += "/1"
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(`{"name":"invalid","m_type":"1","conditions":"[]"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		var got struct {
			Code int `json:"code"`
		}
		if err := stdjson.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.Code != 422 {
			t.Fatalf("%s accepted invalid conditions: %s", method, w.Body.String())
		}
	}
}
