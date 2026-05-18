CREATE TABLE IF NOT EXISTS purchase_order_states (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

PRAGMA foreign_keys=off;

-- Recreate purchase_orders to add FK to purchase_order_states
CREATE TABLE new_purchase_orders (
    id TEXT PRIMARY KEY,
    vendor_id TEXT NOT NULL REFERENCES suppliers(id),
    target_build TEXT,
    status TEXT NOT NULL DEFAULT 'Draft' REFERENCES purchase_order_states(name),
    total_value REAL NOT NULL DEFAULT 0.0,
    payment_terms TEXT,
    expected_delivery DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

INSERT INTO new_purchase_orders (id, vendor_id, target_build, status, total_value, payment_terms, expected_delivery, created_at, updated_at, deleted_at)
SELECT id, vendor_id, target_build, status, total_value, payment_terms, expected_delivery, created_at, updated_at, deleted_at FROM purchase_orders;

DROP TABLE purchase_orders;
ALTER TABLE new_purchase_orders RENAME TO purchase_orders;

PRAGMA foreign_keys=on;
