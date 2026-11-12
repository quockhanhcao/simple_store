package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/database/sqlcdb"
)

func NewHandler(
	logger *slog.Logger,
	cfg *config.Config,
	repository *sqlcdb.Queries,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		cfg,
		logger,
		repository,
	)
	var handler http.Handler = mux
	return handler
}
