CREATE SEQUENCE IF NOT EXISTS orders_order_number_seq;
ALTER TABLE orders ALTER COLUMN order_number SET DEFAULT nextval('orders_order_number_seq');
ALTER SEQUENCE orders_order_number_seq OWNED BY orders.order_number;
ALTER TABLE orders ALTER COLUMN order_number DROP NOT NULL;

DROP TABLE IF EXISTS order_daily_counters;
