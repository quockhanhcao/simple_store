package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quockhanhcao/simple_store/internal/database/sqlcdb"
	"github.com/quockhanhcao/simple_store/internal/identity"
	"github.com/shopspring/decimal"
)

func Seed(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) error {
	err := seedCategory(ctx, pool)
	if err != nil {
		return fmt.Errorf("categories: %w", err)
	}
	err = seedProducts(ctx, pool)
	if err != nil {
		return fmt.Errorf("products: %w", err)
	}
	err = seedUser(ctx, pool)
	if err != nil {
		logger.Error("seed demo user", "err", err)
	}
	return nil
}

func seedCategory(ctx context.Context, pool *pgxpool.Pool) error {
	var seedCategories = []sqlcdb.InsertCategoryParams{
		{Name: "Electronics", Description: "Electronic devices and accessories"},
		{Name: "Clothing", Description: "Fashion and apparel"},
		{Name: "Books", Description: "Books and educational materials"},
		{Name: "Home & Garden", Description: "Home and garden products"},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	db := sqlcdb.New(tx)
	if err := db.AcquireSeedLock(ctx); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}

	exist, err := db.CategoriesExist(ctx)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}
	if exist {
		return nil
	}

	for _, category := range seedCategories {
		err = db.InsertCategory(ctx, category)
		if err != nil {
			return fmt.Errorf("insert %q: %w", category.Name, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func seedUser(ctx context.Context, pool *pgxpool.Pool) error {
	var demoUser = pgtype.Text{String: "demo@simplestore.local", Valid: true}
	var password = "Demo123!"
	var normalized = pgtype.Text{String: strings.ToUpper("demo@simplestore.local"), Valid: true}

	var user = sqlcdb.InsertUserParams{
		ID:                 uuid.NewString(),
		UserName:           demoUser,
		NormalizedUserName: normalized,
		Email:              demoUser,
		NormalizedEmail:    normalized,
		EmailConfirmed:     true,
		FullName:           "Demo User",
		ConcurrencyStamp:   pgtype.Text{String: uuid.NewString(), Valid: true},
		SecurityStamp:      pgtype.Text{String: identity.NewSecurityStamp(), Valid: true},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	db := sqlcdb.New(tx)
	if err := db.AcquireSeedLock(ctx); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}

	_, err = db.GetUserByNormalizedEmail(ctx, normalized)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check existence: %w", err)
	}

	hashedPassword, err := identity.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = pgtype.Text{String: hashedPassword, Valid: true}
	err = db.InsertUser(ctx, user)
	if err != nil {
		return fmt.Errorf("insert %q: %w", user.Email.String, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func seedProducts(ctx context.Context, pool *pgxpool.Pool) error {
	var categories = []string{
		"Electronics",
		"Clothing",
		"Books",
		"Home & Garden",
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	db := sqlcdb.New(tx)
	if err := db.AcquireSeedLock(ctx); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}

	// checking existence
	exist, err := db.ProductsExist(ctx)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}
	if exist {
		return nil
	}

	// get categories' id
	var categoriesIDMap = make(map[string]int32)
	for _, category := range categories {
		id, err := db.GetCategoryIDByName(ctx, category)
		if err != nil {
			return fmt.Errorf("get category id %q: %w", category, err)
		}
		categoriesIDMap[category] = id
	}
	seedProducts := []sqlcdb.InsertProductParams{
		{
			Name: "Wireless Headphones", Description: "Premium wireless headphones with noise cancellation", ImageUrl: "/images/products/headphones.jpg", Stock: 50, Price: decimal.RequireFromString("199.99"), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Smartphone Pro", Description: "Latest flagship smartphone with advanced camera", ImageUrl: "/images/products/smartphone.jpg", Stock: 30, Price: decimal.RequireFromString("999.99"), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Laptop Ultra", Description: "High-performance ultrabook for professionals", ImageUrl: "/images/products/laptop.jpg", Stock: 20, Price: decimal.RequireFromString("1499.99"), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Smart Watch", Description: "Feature-rich smartwatch with health monitoring", ImageUrl: "/images/products/smartwatch.jpg", Stock: 40, Price: decimal.RequireFromString("299.99"), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Classic T-Shirt", Description: "Comfortable cotton t-shirt in various colors", ImageUrl: "/images/products/tshirt.jpg", Stock: 100, Price: decimal.RequireFromString("29.99"), CategoryID: categoriesIDMap["Clothing"],
		},
		{
			Name: "Denim Jeans", Description: "Premium quality denim jeans", ImageUrl: "/images/products/jeans.jpg", Stock: 75, Price: decimal.RequireFromString("79.99"), CategoryID: categoriesIDMap["Clothing"],
		},
		{
			Name: "Golang in Action", Description: "Complete guide to building web apps with Golang", ImageUrl: "/images/products/golang.jpg", Stock: 60, Price: decimal.RequireFromString("49.99"), CategoryID: categoriesIDMap["Books"],
		},
		{
			Name: "Clean Code", Description: "A Handbook of Agile Software Craftsmanship", ImageUrl: "/images/products/clean-code.jpg", Stock: 55, Price: decimal.RequireFromString("39.99"), CategoryID: categoriesIDMap["Books"],
		},
		{
			Name: "Garden Tool Set", Description: "Complete set of essential garden tools", ImageUrl: "/images/products/garden-tools.jpg", Stock: 35, Price: decimal.RequireFromString("89.99"), CategoryID: categoriesIDMap["Home & Garden"],
		},
		{
			Name: "Indoor Plant Kit", Description: "Starter kit for indoor plants with pots and soil", ImageUrl: "/images/products/plant-kit.jpg", Stock: 45, Price: decimal.RequireFromString("34.99"), CategoryID: categoriesIDMap["Home & Garden"],
		},
	}

	// inserting
	for _, product := range seedProducts {
		err = db.InsertProduct(ctx, product)
		if err != nil {
			return fmt.Errorf("insert %q: %w", product.Name, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
