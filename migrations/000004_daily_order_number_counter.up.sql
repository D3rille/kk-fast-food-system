-- Per-store daily counter backing sequential, day-scoped order numbers.
CREATE TABLE IF NOT EXISTS order_daily_counters (
    store_id    UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_date  DATE NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    PRIMARY KEY (store_id, order_date)
);

-- order_number is now assigned by the application via order_daily_counters
-- rather than a global, ever-increasing SERIAL sequence.
ALTER TABLE orders ALTER COLUMN order_number DROP DEFAULT;
ALTER TABLE orders ALTER COLUMN order_number SET NOT NULL;
DROP SEQUENCE IF EXISTS orders_order_number_seq;
