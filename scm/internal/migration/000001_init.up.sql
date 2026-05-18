CREATE TABLE IF NOT EXISTS suppliers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    contact TEXT NOT NULL,
    tier TEXT NOT NULL DEFAULT 'Qualified',
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    quality_score REAL NOT NULL DEFAULT 0.0,
    on_time_rate REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS sku_mappings (
    id TEXT PRIMARY KEY,
    supplier_id TEXT NOT NULL REFERENCES suppliers(id),
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    unit_price REAL NOT NULL DEFAULT 0.0,
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    min_order_qty INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS purchase_orders (
    id TEXT PRIMARY KEY,
    vendor_id TEXT NOT NULL,
    target_build TEXT,
    status TEXT NOT NULL DEFAULT 'Draft',
    total_value REAL NOT NULL DEFAULT 0.0,
    payment_terms TEXT,
    expected_delivery DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS po_line_items (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    po_id TEXT NOT NULL REFERENCES purchase_orders(id),
    sku TEXT NOT NULL,
    description TEXT NOT NULL,
    ordered_qty INTEGER NOT NULL DEFAULT 0,
    received_qty INTEGER NOT NULL DEFAULT 0,
    unit_price REAL NOT NULL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS shipments (
    id TEXT PRIMARY KEY,
    po_ref TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Scheduled',
    carrier TEXT,
    tracking_no TEXT,
    origin TEXT,
    ship_date DATETIME,
    eta DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS shipment_items (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    shipment_id TEXT NOT NULL REFERENCES shipments(id),
    sku TEXT NOT NULL,
    description TEXT NOT NULL,
    qty INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS goods_receipts (
    id TEXT PRIMARY KEY,
    po_ref TEXT NOT NULL,
    vendor_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Pending',
    arrival_date DATETIME NOT NULL,
    operator_id TEXT,
    locked_by TEXT,
    lock_expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS gr_line_items (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    gr_id TEXT NOT NULL REFERENCES goods_receipts(id),
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    ordered_qty INTEGER NOT NULL DEFAULT 0,
    received_qty INTEGER,
    defective_qty INTEGER,
    aging_sensitive INTEGER NOT NULL DEFAULT 0,
    production_date DATETIME,
    aging_label TEXT
);

CREATE TABLE IF NOT EXISTS component_stocks (
    sku TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    stock_qty INTEGER NOT NULL DEFAULT 0,
    reorder_point INTEGER NOT NULL DEFAULT 0,
    unit_cost REAL NOT NULL DEFAULT 0.0,
    status TEXT NOT NULL DEFAULT 'In Stock',
    primary_supplier_id TEXT,
    lead_time_days INTEGER NOT NULL DEFAULT 0,
    location TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    product_model_code TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    product_name TEXT NOT NULL,
    serial_number TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS product_models (
    model_code TEXT PRIMARY KEY,
    model_name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS parts (
    id TEXT PRIMARY KEY,
    part_catalog_id TEXT NOT NULL,
    product_id TEXT,
    serial_number TEXT NOT NULL,
    part_condition_id INTEGER NOT NULL DEFAULT 1,
    manufactured_date DATETIME NOT NULL,
    installation_date DATETIME,
    removal_date DATETIME,
    scrapped_date DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS part_catalogs (
    id TEXT PRIMARY KEY,
    part_number TEXT NOT NULL,
    part_types_id INTEGER NOT NULL,
    mfg_number TEXT NOT NULL,
    description TEXT,
    part_mfg_status INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS part_conditions (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS part_mfg_statuses (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    account_status INTEGER NOT NULL DEFAULT 1,
    role_id INTEGER NOT NULL,
    email TEXT NOT NULL UNIQUE,
    full_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    province TEXT,
    fcm_token TEXT,
    installation_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS warranties (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    warranty_status TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);
