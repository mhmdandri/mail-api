package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mail-api/internal/config"
	"mail-api/internal/mailer"
)

func setupTestHandler() *Handler {
	cfg := &config.Config{
		Port:      "8082",
		APIKey:    "test-secret-key",
		SMTPHost:  "smtp.mail.me.com",
		SMTPPort:  587,
		SMTPUser:  "mhmdandri03@icloud.com",
		SMTPPass:  "dummy-pass",
		FromEmail: "no-reply@weddingofandricica.me",
		FromName:  "Andri & Cica",
	}
	m := mailer.New(cfg)
	return New(cfg, m)
}

func TestHealthCheck(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var res Response
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", res.Status)
	}
}

func TestAuthentication(t *testing.T) {
	h := setupTestHandler()

	// 1. Missing X-API-Key
	req := httptest.NewRequest(http.MethodPost, "/api/send-email", bytes.NewBufferString("{}"))
	rr := httptest.NewRecorder()
	h.SendEmail(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 when API key is missing, got %d", rr.Code)
	}

	// 2. Invalid X-API-Key
	req = httptest.NewRequest(http.MethodPost, "/api/send-email", bytes.NewBufferString("{}"))
	req.Header.Set("X-API-Key", "wrong-key")
	rr = httptest.NewRecorder()
	h.SendEmail(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 when API key is wrong, got %d", rr.Code)
	}
}

func TestPayloadValidation(t *testing.T) {
	h := setupTestHandler()

	tests := []struct {
		name       string
		payload    string
		expectCode int
	}{
		{
			name:       "Empty Body",
			payload:    "",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Missing 'to'",
			payload:    `{"subject": "Hello", "text": "World"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid 'to' format",
			payload:    `{"to": "not-an-email", "subject": "Hello", "text": "World"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Missing 'subject'",
			payload:    `{"to": "recipient@example.com", "text": "World"}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Neither text nor html provided",
			payload:    `{"to": "recipient@example.com", "subject": "Hello"}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/send-email", bytes.NewBufferString(tt.payload))
			req.Header.Set("X-API-Key", "test-secret-key")
			rr := httptest.NewRecorder()

			h.SendEmail(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("%s: expected status %d, got %d. Body: %s", tt.name, tt.expectCode, rr.Code, rr.Body.String())
			}
		})
	}
}
