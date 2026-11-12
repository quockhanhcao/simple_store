package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/database/sqlcdb"
)

func addRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	logger *slog.Logger,
	repository *sqlcdb.Queries,
) {
	if cfg.IsDevelopment() {
		mux.Handle("GET /alive", handleLivenessCheck())
	}
	mux.Handle("GET /api/v1/categories", handleGetCategories(logger, repository))
}
