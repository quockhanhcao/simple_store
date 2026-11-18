package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/quockhanhcao/simple_store/internal/catalog"
	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/database"
	"github.com/shopspring/decimal"
)

// TestMain runs once before every test in this package. main() never runs under
// `go test`, so the global decimal setting from main() has to be repeated here.
func TestMain(m *testing.M) {
	decimal.MarshalJSONWithoutQuotes = true
	os.Exit(m.Run())
}

// fakeCatalog stands in for *catalog.Service. It returns canned values and
// records what the handler passed in, so tests can check the parsing.
type fakeCatalog struct {
	categories []catalog.Category
	products   []catalog.Product
	product    catalog.Product
	err        error

	gotListArgs  catalog.ListProductsParams
	gotProductID int32
}

func (f *fakeCatalog) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return f.categories, f.err
}

func (f *fakeCatalog) ListProducts(ctx context.Context, args catalog.ListProductsParams) ([]catalog.Product, error) {
	f.gotListArgs = args
	return f.products, f.err
}

func (f *fakeCatalog) GetProduct(ctx context.Context, productID int32) (catalog.Product, error) {
	f.gotProductID = productID
	return f.product, f.err
}

// serve sends one request through the real mux, so routing, path values and
// 405s are tested too.
func serve(t *testing.T, svc catalogService, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	srv := NewHandler(slog.New(slog.DiscardHandler), &config.Config{Env: "Development"}, svc)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

// assertErrorEnvelope checks the §3 error shape: {"error":{"code","message"}}.
func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not the error envelope: %v; body = %s", err, rec.Body)
	}
	if body.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Errorf("error.message is empty")
	}
}

func assertProductJSON(t *testing.T, p map[string]any, want catalog.Product) {
	t.Helper()

	wantKeys := []string{"category", "categoryId", "description", "id", "imageUrl", "name", "price", "stock"}
	var gotKeys []string
	for k := range p {
		gotKeys = append(gotKeys, k)
	}
	slices.Sort(gotKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Errorf("product keys = %v, want %v", gotKeys, wantKeys)
	}

	if cat, ok := p["category"].(map[string]any); !ok {
		t.Errorf("category = %v, want a nested object {id,name}", p["category"])
	} else if _, ok := cat["name"]; !ok {
		t.Errorf("category has no name: %v", cat)
	}

	if price, ok := p["price"].(float64); !ok {
		t.Errorf("price = %#v (%T), want a JSON number", p["price"], p["price"])
	} else if price != want.Price.InexactFloat64() {
		t.Errorf("price = %v, want %v", price, want.Price)
	}

	if p["id"] != float64(want.ID) {
		t.Errorf("id = %v, want %d", p["id"], want.ID)
	}
	if p["name"] != want.ProductName {
		t.Errorf("name = %v, want %q", p["name"], want.ProductName)
	}
	if p["imageUrl"] != want.ImageURL {
		t.Errorf("imageUrl = %v, want %q", p["imageUrl"], want.ImageURL)
	}
	// stock 0 must still be sent (no omitempty on numbers).
	if p["stock"] != float64(want.Stock) {
		t.Errorf("stock = %v, want %d", p["stock"], want.Stock)
	}
}

var phone = catalog.Product{
	ID:          1,
	ProductName: "Smartphone Pro",
	Description: "A smartphone",
	Price:       decimal.RequireFromString("1299.99"),
	ImageURL:    "/images/phone.jpg",
	Stock:       0,
}

func TestRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		env        string
		method     string
		path       string
		wantStatus int
		wantBody   string // checked only when non-empty
	}{
		{"alive in Development", "Development", http.MethodGet, "/alive", http.StatusOK, "Healthy"},
		{"HEAD alive is allowed", "Development", http.MethodHead, "/alive", http.StatusOK, ""},
		{"wrong method on alive", "Development", http.MethodPost, "/alive", http.StatusMethodNotAllowed, ""},
		{"alive is not mapped in Production", "Production", http.MethodGet, "/alive", http.StatusNotFound, ""},
		{"unknown path", "Development", http.MethodGet, "/nope", http.StatusNotFound, ""},
		{"POST products is 405", "Development", http.MethodPost, "/api/v1/products", http.StatusMethodNotAllowed, ""},
		{"POST product is 405", "Development", http.MethodPost, "/api/v1/products/1", http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// DiscardHandler: tests don't need log output.
			srv := NewHandler(slog.New(slog.DiscardHandler), &config.Config{Env: tt.env}, &fakeCatalog{})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			srv.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusOK {
				if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
					t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
				}
			}
		})
	}
}

func TestListCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fakeSvc    fakeCatalog
		wantStatus int
		wantBody   string
	}{
		{
			name: "the fake returns 2 categories",
			fakeSvc: fakeCatalog{categories: []catalog.Category{
				{ID: 1, Name: "category A", Description: "description A"},
				{ID: 2, Name: "category B", Description: "description B"},
			}},
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":1,"name":"category A","description":"description A"},{"id":2,"name":"category B","description":"description B"}]` + "\n",
		},
		{
			name:       "the fake returns nothing",
			fakeSvc:    fakeCatalog{},
			wantStatus: http.StatusOK,
			wantBody:   "[]\n",
		},
		{
			name:       "the fake returns an error",
			fakeSvc:    fakeCatalog{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal","message":"An unexpected error occurred."}}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, &tt.fakeSvc, http.MethodGet, "/api/v1/categories")

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	t.Parallel()

	ptr := func(v int32) *int32 { return &v }

	tests := []struct {
		name           string
		query          string
		fakeSvc        fakeCatalog
		wantStatus     int
		wantCode       string
		wantCategoryID *int32
		wantSearch     string
	}{
		{
			name:       "no filters returns the list",
			fakeSvc:    fakeCatalog{products: []catalog.Product{phone}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty list is [] not null",
			fakeSvc:    fakeCatalog{},
			wantStatus: http.StatusOK,
		},
		{
			name:           "categoryId and search reach the service",
			query:          "?categoryId=1&search=Smart",
			wantStatus:     http.StatusOK,
			wantCategoryID: ptr(1),
			wantSearch:     "Smart",
		},
		{
			name:       "whitespace search is passed through untrimmed",
			query:      "?search=%20%20",
			wantStatus: http.StatusOK,
			wantSearch: "  ",
		},
		{
			name:       "a literal % is passed through, not as a wildcard",
			query:      "?search=%25",
			wantStatus: http.StatusOK,
			wantSearch: "%",
		},
		{
			name:       "non-integer categoryId is 400",
			query:      "?categoryId=abc",
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name:       "categoryId overflowing int32 is 400",
			query:      "?categoryId=99999999999",
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name:       "service error is a generic 500",
			fakeSvc:    fakeCatalog{err: errors.New("boom: connection refused")},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, &tt.fakeSvc, http.MethodGet, "/api/v1/products"+tt.query)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "boom") {
				t.Errorf("body leaks the internal error: %s", rec.Body)
			}
			if tt.wantStatus != http.StatusOK {
				assertErrorEnvelope(t, rec, tt.wantCode)
				return
			}

			got := tt.fakeSvc.gotListArgs
			if !equalPtr(got.CategoryID, tt.wantCategoryID) {
				t.Errorf("service got CategoryID = %s, want %s", fmtPtr(got.CategoryID), fmtPtr(tt.wantCategoryID))
			}
			if got.Search != tt.wantSearch {
				t.Errorf("service got Search = %q, want %q", got.Search, tt.wantSearch)
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want JSON", ct)
			}
			var body []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not a JSON array: %v; body = %s", err, rec.Body)
			}
			if body == nil {
				t.Fatalf("body = null, want [] (got %q)", rec.Body)
			}
			if len(body) != len(tt.fakeSvc.products) {
				t.Fatalf("len(body) = %d, want %d", len(body), len(tt.fakeSvc.products))
			}
			for i, p := range body {
				assertProductJSON(t, p, tt.fakeSvc.products[i])
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		fakeSvc    fakeCatalog
		wantStatus int
		wantCode   string
		wantID     int32
	}{
		{
			name:       "existing product",
			path:       "/api/v1/products/1",
			fakeSvc:    fakeCatalog{product: phone},
			wantStatus: http.StatusOK,
			wantID:     1,
		},
		{
			name:       "missing product is 404",
			path:       "/api/v1/products/9999",
			fakeSvc:    fakeCatalog{err: fmt.Errorf("get product 9999: %w", database.ErrNotFound)},
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
		{
			name:       "non-integer id is 400",
			path:       "/api/v1/products/abc",
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name:       "id overflowing int32 is 400",
			path:       "/api/v1/products/99999999999",
			wantStatus: http.StatusBadRequest,
			wantCode:   "bad_request",
		},
		{
			name:       "service error is a generic 500",
			path:       "/api/v1/products/1",
			fakeSvc:    fakeCatalog{err: errors.New("boom: connection refused")},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, &tt.fakeSvc, http.MethodGet, tt.path)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "boom") {
				t.Errorf("body leaks the internal error: %s", rec.Body)
			}
			if tt.wantStatus != http.StatusOK {
				assertErrorEnvelope(t, rec, tt.wantCode)
				return
			}

			if tt.fakeSvc.gotProductID != tt.wantID {
				t.Errorf("service got id = %d, want %d", tt.fakeSvc.gotProductID, tt.wantID)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not a JSON object: %v; body = %s", err, rec.Body)
			}
			assertProductJSON(t, body, tt.fakeSvc.product)
		})
	}
}

func equalPtr(a, b *int32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func fmtPtr(p *int32) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprint(*p)
}
