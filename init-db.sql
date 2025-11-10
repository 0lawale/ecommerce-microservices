-- This script creates separate databases for each microservice
-- Runs automatically when postgres container starts for the first time

-- Create databases for each service
CREATE DATABASE user_service;
CREATE DATABASE product_service;
CREATE DATABASE order_service;
CREATE DATABASE notification_service;

-- Connect to user_service and create schema
\c user_service;

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Insert test admin user (password: admin123)
-- Password hash generated with bcrypt cost 10
INSERT INTO users (id, email, password_hash, full_name, role, created_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@ecommerce.com',
    '$2a$10$YourHashedPasswordHere',
    'Admin User',
    'admin',
    CURRENT_TIMESTAMP
) ON CONFLICT DO NOTHING;

-- Connect to product_service and create schema
\c product_service;

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    category VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_price ON products(price);

-- Insert sample products
INSERT INTO products (id, name, description, price, stock, category, created_at)
VALUES
    ('prod-001', 'Laptop Pro', 'High-performance laptop', 1299.99, 50, 'Electronics', CURRENT_TIMESTAMP),
    ('prod-002', 'Wireless Mouse', 'Ergonomic wireless mouse', 29.99, 200, 'Accessories', CURRENT_TIMESTAMP),
    ('prod-003', 'USB-C Cable', 'Fast charging USB-C cable', 19.99, 500, 'Accessories', CURRENT_TIMESTAMP),
    ('prod-004', 'Mechanical Keyboard', 'RGB mechanical keyboard', 149.99, 75, 'Accessories', CURRENT_TIMESTAMP),
    ('prod-005', 'Monitor 27"', '4K Ultra HD monitor', 449.99, 30, 'Electronics', CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- Connect to order_service and create schema
\c order_service;

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- Connect to notification_service and create schema
\c notification_service;

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(255),
    message TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);