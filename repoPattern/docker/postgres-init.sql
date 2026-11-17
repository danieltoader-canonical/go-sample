CREATE TABLE customer (
    customer_id BIGINT PRIMARY KEY,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    customer_name TEXT NOT NULL DEFAULT ''
);

INSERT INTO customer (customer_id, customer_name)
VALUES (1, 'Sample Customer')
ON CONFLICT (customer_id) DO NOTHING;
