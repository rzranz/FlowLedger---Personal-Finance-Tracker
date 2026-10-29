CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    created_at TEXT NOT NULL,
    UNIQUE(name COLLATE NOCASE, type)
);
CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    amount_idr INTEGER NOT NULL CHECK(amount_idr > 0),
    occurred_on TEXT NOT NULL, -- YYYY-MM-DD
    note TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tx_occurred_id ON transactions(occurred_on DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_tx_cat_occurred ON transactions(category_id, occurred_on);
CREATE INDEX IF NOT EXISTS idx_tx_month_type ON transactions(substr(occurred_on, 1, 7), type);

CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    month TEXT NOT NULL, -- YYYY-MM
    limit_idr INTEGER NOT NULL CHECK(limit_idr > 0),
    updated_at TEXT NOT NULL,
    UNIQUE(category_id, month)
);
CREATE INDEX IF NOT EXISTS idx_budgets_month ON budgets(month);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_body BLOB NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_idem_created ON idempotency_keys(created_at);

