package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mail-api/internal/config"
	"mail-api/internal/handler"
	"mail-api/internal/mailer"
)

func main() {
	// Initialize JSON structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Configuration checks
	if cfg.APIKey == "" {
		slog.Warn("PERINGATAN: API_KEY kosong! Semua request ke /api/send-email akan ditolak (401). Harap tentukan API_KEY di .env atau environment variables.")
	}
	if cfg.SMTPPass == "" {
		slog.Warn("PERINGATAN: SMTP_PASS kosong! Pengiriman email kemungkinan besar akan gagal saat autentikasi SMTP.")
	}

	// Initialize Mailer and Handler
	mailService := mailer.New(cfg)
	h := handler.New(cfg, mailService)

	// Setup HTTP router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/send-email", h.SendEmail)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in background goroutine
	go func() {
		slog.Info("Server email microservice berjalan",
			"port", cfg.Port,
			"smtp_host", cfg.SMTPHost,
			"smtp_port", cfg.SMTPPort,
			"from_email", cfg.FromEmail,
			"from_name", cfg.FromName,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Gagal menjalankan server HTTP", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	slog.Info("Sinyal terminasi diterima, mematikan server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Terjadi error saat shutdown server", "error", err.Error())
	}

	slog.Info("Server berhasil dimatikan secara graceful")
}
