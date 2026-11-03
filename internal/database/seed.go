package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quockhanhcao/simple_store/internal/queries"
	"github.com/shopspring/decimal"
)

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	err := seedCategory(ctx, pool)
	if err != nil {
		return err
	}
	err = seedProducts(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

func seedCategory(ctx context.Context, pool *pgxpool.Pool) error {
	var seedCategories = []queries.InsertCategoryParams{
		{Name: "Electronics", Description: "Electronic devices and accessories"},
		{Name: "Clothing", Description: "Fashion and apparel"},
		{Name: "Books", Description: "Books and educational materials"},
		{Name: "Home & Garden", Description: "Home and garden products"},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := queries.New(pool).WithTx(tx)
	exist, err := query.CategoriesExist(ctx)
	if err != nil {
		return fmt.Errorf("error checking categories existence: %w", err)
	}
	if exist {
		return nil
	}

	for _, category := range seedCategories {
		err = query.InsertCategory(ctx, category)
		if err != nil {
			return fmt.Errorf("error seeding category: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error committing seeding categories: %w", err)
	}

	return nil
}

// func seedUser(ctx context.Context) error {
// TODO: seed 1 user:
// username: DemoUserEmail
// email: DemoUserEmail
// email confirm: true
// full name: Demo User

// check existence
// insert
// }

func seedProducts(ctx context.Context, pool *pgxpool.Pool) error {
	var categories = []string{
		"Electronics",
		"Clothing",
		"Books",
		"Home & Garden",
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := queries.New(pool).WithTx(tx)

	// checking existence
	exist, err := query.ProductsExist(ctx)
	if err != nil {
		return fmt.Errorf("error checking products existence: %w", err)
	}
	if exist {
		return nil
	}

	// get categories' id
	var categoriesIDMap = make(map[string]int32)
	for _, category := range categories {
		id, err := query.GetCategoryIDByName(ctx, category)
		if err != nil {
			return fmt.Errorf("error getting category id: %w", err)
		}
		categoriesIDMap[category] = id
	}
	seedProducts := []queries.InsertProductParams{
		{
			Name: "Wireless Headphones", Description: "Premium wireless headphones with noise cancellation", ImageUrl: "/images/products/headphones.jpg", Stock: 50, Price: decimal.NewFromFloat(199.99), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Smartphone Pro", Description: "Latest flagship smartphone with advanced camera", ImageUrl: "/images/products/smartphone.jpg", Stock: 30, Price: decimal.NewFromFloat(999.99), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Laptop Ultra", Description: "High-performance ultrabook for professionals", ImageUrl: "/images/products/laptop.jpg", Stock: 20, Price: decimal.NewFromFloat(1499.99), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Smart Watch", Description: "Feature-rich smartwatch with health monitoring", ImageUrl: "/images/products/smartwatch.jpg", Stock: 40, Price: decimal.NewFromFloat(299.99), CategoryID: categoriesIDMap["Electronics"],
		},
		{
			Name: "Classic T-Shirt", Description: "Comfortable cotton t-shirt in various colors", ImageUrl: "/images/products/tshirt.jpg", Stock: 100, Price: decimal.NewFromFloat(29.99), CategoryID: categoriesIDMap["Clothing"],
		},
		{
			Name: "Denim Jeans", Description: "Premium quality denim jeans", ImageUrl: "/images/products/jeans.jpg", Stock: 75, Price: decimal.NewFromFloat(79.99), CategoryID: categoriesIDMap["Clothing"],
		},
		{
			Name: "Golang in Action", Description: "Complete guide to building web apps with Golang", ImageUrl: "/images/products/golang.jpg", Stock: 60, Price: decimal.NewFromFloat(49.99), CategoryID: categoriesIDMap["Books"],
		},
		{
			Name: "Clean Code", Description: "A Handbook of Agile Software Craftsmanship", ImageUrl: "/images/products/clean-code.jpg", Stock: 55, Price: decimal.NewFromFloat(39.99), CategoryID: categoriesIDMap["Books"],
		},
		{
			Name: "Garden Tool Set", Description: "Complete set of essential garden tools", ImageUrl: "/images/products/garden-tools.jpg", Stock: 35, Price: decimal.NewFromFloat(89.99), CategoryID: categoriesIDMap["Home & Garden"],
		},
		{
			Name: "Indoor Plant Kit", Description: "Starter kit for indoor plants with pots and soil", ImageUrl: "/images/products/plant-kit.jpg", Stock: 45, Price: decimal.NewFromFloat(34.99), CategoryID: categoriesIDMap["Home & Garden"],
		},
	}

	// inserting
	for _, product := range seedProducts {
		err = query.InsertProduct(ctx, product)
		if err != nil {
			return fmt.Errorf("error seeding product: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error committing seeding products: %w", err)
	}

	return nil
}
