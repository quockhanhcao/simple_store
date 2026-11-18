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
	catalogSvc catalogService,
) {
	if cfg.IsDevelopment() {
		mux.Handle("GET /alive", handleLivenessCheck())
	}
	mux.Handle("GET /api/v1/categories", handleListCategories(logger, catalogSvc))
	mux.Handle("GET /api/v1/products", handleListProducts(logger, catalogSvc))
	mux.Handle("GET /api/v1/products/{id}", handleGetProduct(logger, catalogSvc))
}
