CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    items JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    source VARCHAR(32) NOT NULL DEFAULT 'web',
    payment_method VARCHAR(32) NOT NULL DEFAULT 'cod',
    payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
    shipping_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    shipping_address TEXT NOT NULL,
    vendor_id VARCHAR(64) NOT NULL,
    order_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);