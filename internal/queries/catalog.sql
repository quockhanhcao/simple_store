
-- name: ListCategories :many
SELECT * FROM categories;

-- name: ListProducts :many
SELECT p.id, p.name AS product_name, p.description, p.price, p.stock, p.image_url, c.id AS category_id, c.name AS category_name FROM products p
JOIN categories c ON p.category_id = c.id
WHERE (sqlc.narg('category_id')::int IS NULL OR p.category_id = sqlc.narg('category_id'))
AND (sqlc.narg('search')::text IS NULL
        OR p.name LIKE '%' || sqlc.narg('search') || '%'
        OR p.description LIKE '%' || sqlc.narg('search') || '%');

-- name: GetProduct :one
SELECT p.id, p.name AS product_name, p.description, p.price, p.stock, p.image_url, c.id AS category_id, c.name AS category_name FROM products p
JOIN categories c ON p.category_id = c.id
WHERE p.id = $1;
