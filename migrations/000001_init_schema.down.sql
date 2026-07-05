-- Drop Indexes
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_order_item_modifiers_order_item_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_store_id;
DROP INDEX IF EXISTS idx_modifier_options_modifier_group_id;
DROP INDEX IF EXISTS idx_products_is_available;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_categories_store_id;
DROP INDEX IF EXISTS idx_users_store_id;

-- Drop Tables
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_item_modifiers;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS product_modifier_groups;
DROP TABLE IF EXISTS modifier_options;
DROP TABLE IF EXISTS modifier_groups;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS stores;

-- Drop Enums
DROP TYPE IF EXISTS payment_provider;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS order_source;
DROP TYPE IF EXISTS user_role;
