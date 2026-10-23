package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/config"
)

func addRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	logger *slog.Logger,
) {
	if cfg.IsDevelopment() {
		mux.Handle("GET /alive", handleLivenessCheck())
	}
}
