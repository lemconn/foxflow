CREATE TABLE fox_users
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL DEFAULT '',
    exchange   TEXT NOT NULL DEFAULT 'binance' CHECK (exchange IN ('binance', 'okx')),
    access_key TEXT NOT NULL DEFAULT '',
    secret_key TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    trade_type TEXT NOT NULL DEFAULT '' CHECK (trade_type IN ('mock', 'real')),
    created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (username, access_key, secret_key)
);

CREATE TABLE fox_symbols
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL DEFAULT '',
    user_id    INTEGER NOT NULL DEFAULT 0,
    exchange   TEXT    NOT NULL DEFAULT 'binance' CHECK (exchange IN ('binance', 'okx')),
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (user_id, exchange)
);
CREATE INDEX idx_fox_symbols_exchange ON fox_symbols (exchange);
CREATE INDEX idx_fox_symbols_user_id ON fox_symbols (user_id);

CREATE TABLE fox_ss
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL DEFAULT 0,
    order_id   TEXT    NOT NULL DEFAULT '',
    schema     TEXT,
    status     TEXT    NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'pending', 'filled', 'cancelled')),
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);
CREATE INDEX idx_fox_ss_user_id ON fox_ss (user_id);
CREATE INDEX idx_fox_ss_status ON fox_ss (status);
CREATE INDEX idx_fox_ss_order_id ON fox_ss (order_id);