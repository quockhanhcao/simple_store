
-- name: AcquireSeedLock :exec
SELECT pg_advisory_xact_lock(7243001);

-- name: CategoriesExist :one
SELECT EXISTS (SELECT 1 FROM categories);

-- name: ProductsExist :one
SELECT EXISTS (SELECT 1 FROM products);

-- name: InsertCategory :exec
INSERT INTO categories (
    name,
    description
) VALUES (
    $1, $2
);

-- name: GetCategoryIDByName :one
SELECT id FROM categories
WHERE name = $1;

-- name: InsertProduct :exec
INSERT INTO products (
    name,
    description,
    price,
    image_url,
    stock,
    category_id
) VALUES (
    $1, $2, $3, $4, $5, $6
);
