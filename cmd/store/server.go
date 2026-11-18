package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/config"
)

func NewHandler(
	logger *slog.Logger,
	cfg *config.Config,
	catalogSvc catalogService,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		cfg,
		logger,
		catalogSvc,
	)
	var handler http.Handler = mux
	return handler
}
