package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/httpx"
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
	return httpx.Chain(handler, httpx.RequestID, httpx.AccessLog(logger), httpx.Recover(logger))
}
