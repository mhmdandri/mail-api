package handler

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"mail-api/internal/config"
	"mail-api/internal/mailer"
)

type SendEmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Text    string `json:"text,omitempty"`
	HTML    string `json:"html,omitempty"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Handler struct {
	cfg    *config.Config
	mailer *mailer.Mailer
}

func New(cfg *config.Config, mailer *mailer.Mailer) *Handler {
	return &Handler{
		cfg:    cfg,
		mailer: mailer,
	}
}

// Health handles health check requests.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, Response{
			Status:  "error",
			Message: "Method tidak diizinkan",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Service is healthy",
	})
}

// SendEmail handles POST /api/send-email.
func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, Response{
			Status:  "error",
			Message: "Method tidak diizinkan, gunakan POST",
		})
		return
	}

	// 1. Authenticate via X-API-Key header
	apiKey := r.Header.Get("X-API-Key")
	if !h.isAPIKeyValid(apiKey) {
		slog.Warn("Autentikasi gagal: X-API-Key tidak valid atau hilang",
			"remote_addr", r.RemoteAddr,
			"path", r.URL.Path,
		)
		h.writeJSON(w, http.StatusUnauthorized, Response{
			Status:  "error",
			Message: "Unauthorized: API Key tidak valid atau tidak disertakan",
		})
		return
	}

	// Limit request payload to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req SendEmailRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.writeJSON(w, http.StatusBadRequest, Response{
				Status:  "error",
				Message: "Payload melebihi batas ukuran maksimum (1MB)",
			})
			return
		}
		slog.Warn("Gagal mem-parsing request body", "error", err.Error(), "remote_addr", r.RemoteAddr)
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Payload JSON tidak valid: " + err.Error(),
		})
		return
	}

	// Ensure no trailing JSON content
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Payload JSON mengandung data tambahan setelah root object",
		})
		return
	}

	// 2. Validate input fields
	req.To = strings.TrimSpace(req.To)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Text = strings.TrimSpace(req.Text)
	req.HTML = strings.TrimSpace(req.HTML)

	if req.To == "" {
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Field 'to' wajib diisi",
		})
		return
	}

	if _, err := mail.ParseAddress(req.To); err != nil {
		slog.Warn("Alamat email penerima tidak valid", "to", req.To, "error", err.Error())
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Format email 'to' tidak valid",
		})
		return
	}

	if req.Subject == "" {
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Field 'subject' wajib diisi",
		})
		return
	}

	if req.Text == "" && req.HTML == "" {
		h.writeJSON(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Setidaknya salah satu dari 'text' atau 'html' harus disediakan",
		})
		return
	}

	// 3. Send email via SMTP
	err := h.mailer.Send(req.To, req.Subject, req.Text, req.HTML)
	duration := time.Since(start)

	if err != nil {
		slog.Error("Pengiriman email gagal",
			"to", req.To,
			"subject", req.Subject,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
		h.writeJSON(w, http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Gagal mengirim email: " + err.Error(),
		})
		return
	}

	slog.Info("Email berhasil dikirim",
		"to", req.To,
		"subject", req.Subject,
		"duration_ms", duration.Milliseconds(),
		"status", "success",
	)

	h.writeJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Email berhasil dikirim",
	})
}

func (h *Handler) isAPIKeyValid(providedKey string) bool {
	if h.cfg.APIKey == "" || providedKey == "" {
		return false
	}
	// Constant time comparison to avoid timing attacks
	return subtle.ConstantTimeCompare([]byte(providedKey), []byte(h.cfg.APIKey)) == 1
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
