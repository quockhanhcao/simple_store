package main

import (
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/httpx"
	"github.com/quockhanhcao/simple_store/internal/database/sqlcdb"
)

type categoryResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func handleLivenessCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Healthy"))
	})
}

func handleGetCategories(logger *slog.Logger, queries *sqlcdb.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := queries.ListCategories(r.Context())
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}

		resp := make([]categoryResponse, 0, len(rows))
		for _, row := range rows {
			resp = append(resp, categoryResponse{ID: row.ID, Name: row.Name, Description: row.Description})
		}
		if err := httpx.WriteJSON(w, http.StatusOK, resp); err != nil {
			httpx.Error(w, r, logger, err)
		}
	})
}
