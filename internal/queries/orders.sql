
-- name: InsertOrder :one
INSERT INTO orders (
    user_id,
    order_date,
    total_amount,
    shipping_address
) VALUES (
    $1, $2, $3, $4
)
RETURNING id;

-- name: InsertOrderItem :exec
INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    unit_price
) VALUES (
    $1, $2, $3, $4
);

-- name: ListOrdersByUser :many
SELECT * FROM orders o
WHERE o.user_id = $1
ORDER BY order_date DESC;

-- name: GetOrderForUser :one
SELECT * FROM orders o
WHERE o.user_id = $1
AND o.id = $2;

-- name: ListOrderItemsByOrderIDs :many
SELECT oi.order_id, oi.id, oi.product_id, oi.quantity, oi.unit_price,
        p.name AS product_name, p.image_url AS product_image_url
FROM order_items oi
JOIN products p ON p.id = oi.product_id
WHERE oi.order_id = ANY(sqlc.arg('order_ids')::int[]);
