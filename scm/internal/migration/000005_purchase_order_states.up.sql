CREATE TABLE IF NOT EXISTS purchase_order_states (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS goods_receipt_states (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS component_stock_states (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS shipment_states (
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
INSERT INTO new_purchase_orders SELECT * FROM purchase_orders;
DROP TABLE purchase_orders;
ALTER TABLE new_purchase_orders RENAME TO purchase_orders;

-- Recreate goods_receipts
CREATE TABLE new_goods_receipts (
    id TEXT PRIMARY KEY,
    po_ref TEXT NOT NULL REFERENCES purchase_orders(id),
    vendor_id TEXT NOT NULL REFERENCES suppliers(id),
    status TEXT NOT NULL DEFAULT 'Pending' REFERENCES goods_receipt_states(name),
    arrival_date DATETIME NOT NULL,
    operator_id TEXT,
    locked_by TEXT,
    lock_expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);
INSERT INTO new_goods_receipts SELECT * FROM goods_receipts;
DROP TABLE goods_receipts;
ALTER TABLE new_goods_receipts RENAME TO goods_receipts;

-- Recreate component_stocks
CREATE TABLE new_component_stocks (
    sku TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    stock_qty INTEGER NOT NULL DEFAULT 0,
    reorder_point INTEGER NOT NULL DEFAULT 0,
    unit_cost REAL NOT NULL DEFAULT 0.0,
    status TEXT NOT NULL DEFAULT 'In Stock' REFERENCES component_stock_states(name),
    primary_supplier_id TEXT REFERENCES suppliers(id),
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);
INSERT INTO new_component_stocks SELECT * FROM component_stocks;
DROP TABLE component_stocks;
ALTER TABLE new_component_stocks RENAME TO component_stocks;

-- Recreate shipments
CREATE TABLE new_shipments (
    id TEXT PRIMARY KEY,
    po_ref TEXT NOT NULL REFERENCES purchase_orders(id),
    supplier_id TEXT NOT NULL REFERENCES suppliers(id),
    status TEXT NOT NULL DEFAULT 'Scheduled' REFERENCES shipment_states(name),
    carrier TEXT,
    tracking_no TEXT,
    origin TEXT,
    ship_date DATETIME,
    eta DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);
INSERT INTO new_shipments SELECT * FROM shipments;
DROP TABLE shipments;
ALTER TABLE new_shipments RENAME TO shipments;

PRAGMA foreign_keys=on;
