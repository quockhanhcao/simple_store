package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/quockhanhcao/simple_store/internal/database"
	"github.com/quockhanhcao/simple_store/internal/database/sqlcdb"
	"github.com/shopspring/decimal"
)

type Category struct {
	ID          int32
	Name        string
	Description string
}

// CategoryRef is the short form of a category that is embedded in a product.
type CategoryRef struct {
	ID   int32
	Name string
}

type Product struct {
	ID          int32
	ProductName string
	Description string
	Price       decimal.Decimal
	ImageURL    string
	Stock       int32
	CategoryID  int32
	Category    CategoryRef
}

type ListProductsParams struct {
	CategoryID *int32
	Search     string
}

type Service struct {
	queries *sqlcdb.Queries
}

func NewService(queries *sqlcdb.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	categories := make([]Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, Category{ID: row.ID, Name: row.Name, Description: row.Description})
	}
	return categories, nil
}

func (s *Service) ListProducts(ctx context.Context, args ListProductsParams) ([]Product, error) {
	searchParams := sqlcdb.ListProductsParams{}
	if args.CategoryID != nil {
		searchParams.CategoryID = pgtype.Int4{Int32: *args.CategoryID, Valid: true}
	}
	if strings.TrimSpace(args.Search) != "" {
		searchParams.Search = pgtype.Text{String: "%" + escapeLike(args.Search) + "%", Valid: true}
	}
	rows, err := s.queries.ListProducts(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	products := make([]Product, 0, len(rows))
	for _, row := range rows {
		// ListProductsRow and GetProductRow have identical fields, so Go allows a direct conversion.
		products = append(products, rowToProduct(sqlcdb.GetProductRow(row)))
	}
	return products, nil
}

func (s *Service) GetProduct(ctx context.Context, productID int32) (Product, error) {
	row, err := s.queries.GetProduct(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("get product %d: %w", productID, database.MapError(err))
	}
	return rowToProduct(row), nil
}

func rowToProduct(row sqlcdb.GetProductRow) Product {
	return Product{
		ID:          row.ID,
		ProductName: row.ProductName,
		Description: row.Description,
		Price:       row.Price,
		ImageURL:    row.ImageUrl,
		Stock:       row.Stock,
		CategoryID:  row.CategoryID,
		Category:    CategoryRef{ID: row.CategoryID, Name: row.CategoryName},
	}
}

func escapeLike(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(s)
}
