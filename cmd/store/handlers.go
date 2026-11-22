package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/quockhanhcao/simple_store/internal/catalog"
	"github.com/quockhanhcao/simple_store/internal/httpx"
	"github.com/shopspring/decimal"
)

type categoryResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// productResponse is the §3 product shape:
// {id,name,description,price,stock,imageUrl,categoryId,category:{id,name}}.
type productResponse struct {
	ID          int32               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       decimal.Decimal     `json:"price"`
	Stock       int32               `json:"stock"`
	ImageURL    string              `json:"imageUrl"`
	CategoryID  int32               `json:"categoryId"`
	Category    categoryRefResponse `json:"category"`
}

type categoryRefResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// toProductResponse is shared by the list and details handlers.
func toProductResponse(p catalog.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		Name:        p.ProductName,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
		CategoryID:  p.CategoryID,
		Category:    categoryRefResponse{ID: p.Category.ID, Name: p.Category.Name},
	}
}

func handleLivenessCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Healthy"))
	})
}

type catalogService interface {
	ListCategories(ctx context.Context) ([]catalog.Category, error)
	ListProducts(ctx context.Context, args catalog.ListProductsParams) ([]catalog.Product, error)
	GetProduct(ctx context.Context, productID int32) (catalog.Product, error)
}

func handleListCategories(logger *slog.Logger, svc catalogService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categories, err := svc.ListCategories(r.Context())
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}

		resp := make([]categoryResponse, 0, len(categories))
		for _, row := range categories {
			resp = append(resp, categoryResponse{ID: row.ID, Name: row.Name, Description: row.Description})
		}
		if err := httpx.WriteJSON(w, http.StatusOK, resp); err != nil {
			httpx.Error(w, r, logger, err)
		}
	})
}

func handleListProducts(logger *slog.Logger, svc catalogService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categoryID := r.URL.Query().Get("categoryId")
		search := r.URL.Query().Get("search")
		searchParams := catalog.ListProductsParams{Search: search}
		parsedCategoryID, err := parseOptionalInt32("categoryId", categoryID)
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}
		searchParams.CategoryID = parsedCategoryID
		products, err := svc.ListProducts(r.Context(), searchParams)
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}

		resp := make([]productResponse, 0, len(products))
		for _, row := range products {
			resp = append(resp, toProductResponse(row))
		}
		if err := httpx.WriteJSON(w, http.StatusOK, resp); err != nil {
			httpx.Error(w, r, logger, err)
		}
	})
}

func handleGetProduct(logger *slog.Logger, svc catalogService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		productID, err := parseInt32("productId", r.PathValue("id"))
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}
		row, err := svc.GetProduct(r.Context(), productID)
		if err != nil {
			httpx.Error(w, r, logger, err)
			return
		}
		if err := httpx.WriteJSON(w, http.StatusOK, toProductResponse(row)); err != nil {
			httpx.Error(w, r, logger, err)
		}
	})
}

func parseOptionalInt32(name, value string) (*int32, error) {
	if value == "" {
		return nil, nil
	}
	v, err := parseInt32(name, value)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func parseInt32(name, value string) (int32, error) {
	if value == "" {
		return 0, httpx.BadRequest(name + " must be an integer")
	}
	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, httpx.BadRequest(name + " must be an integer")
	}
	v := int32(i)
	return v, nil
}
