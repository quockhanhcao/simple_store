package main

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/constant"
	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/httpx"
	"github.com/quockhanhcao/simple_store/web"
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

	// Unknown API paths get a JSON 404, not the file server's text 404.
	// The mux picks the most specific pattern, so the routes above still win.
	// It must be "GET /api/", not "/api/": a method-less "/api/" next to "GET /" panics
	// as a conflict (neither is more specific: one has the longer path, the other the method).
	mux.Handle("GET /api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, http.StatusNotFound, "not_found", constant.NotFoundMsg)
	}))

	// Everything else is the embedded storefront.
	storeFiles, err := fs.Sub(web.StoreFS, "store") // strip the "store/" prefix: /index.html, not /store/index.html
	if err != nil {
		panic(err) // only fails if the path is invalid: a programming error, caught on first start
	}
	mux.Handle("GET /", http.FileServerFS(storeFiles))
}
