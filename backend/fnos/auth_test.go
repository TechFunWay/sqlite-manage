package fnos

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"sqlite-manager/auth"
)

func testRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	auth.SetDB(db)
	if err := auth.InitTables(); err != nil {
		t.Fatal(err)
	}
	if err := InitTables(db); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(db).RegisterRoutes(r.Group("/api/auth/fnos"))
	return r, db
}

func gatewayRequest(method, path string, body any) *http.Request {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trim-Userid", "1001")
	req.Header.Set("X-Trim-Username", "nas-user")
	return req.WithContext(context.WithValue(req.Context(), gatewayContextKey{}, true))
}

func TestRejectsSpoofedGatewayHeaders(t *testing.T) {
	r, _ := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/fnos/identity", nil)
	req.Header.Set("X-Trim-Userid", "1001")
	req.Header.Set("X-Trim-Username", "spoofed")
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestRegisterBindAndLogin(t *testing.T) {
	r, _ := testRouter(t)
	payload := map[string]string{"username": "admin", "password": auth.MD5Hash("secret")}

	registerRecorder := httptest.NewRecorder()
	r.ServeHTTP(registerRecorder, gatewayRequest(http.MethodPost, "/api/auth/fnos/register", payload))
	if registerRecorder.Code != http.StatusOK {
		t.Fatalf("register expected 200, got %d: %s", registerRecorder.Code, registerRecorder.Body.String())
	}

	loginRecorder := httptest.NewRecorder()
	r.ServeHTTP(loginRecorder, gatewayRequest(http.MethodPost, "/api/auth/fnos/login", nil))
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d: %s", loginRecorder.Code, loginRecorder.Body.String())
	}
}
