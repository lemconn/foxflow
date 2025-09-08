CREATE TABLE fox_users
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL DEFAULT '',
    exchange   TEXT NOT NULL DEFAULT 'binance' CHECK (exchange IN ('binance', 'okx', 'gate')),
    access_key TEXT NOT NULL DEFAULT '',
    secret_key TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive')),
    trade_type TEXT NOT NULL DEFAULT '' CHECK (trade_type IN ('mock', 'real')),
    is_active  BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (username, access_key, secret_key)
);

CREATE TABLE fox_symbols
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL DEFAULT '',
    user_id     INTEGER NOT NULL DEFAULT 0,
    exchange    TEXT    NOT NULL DEFAULT 'binance' CHECK (exchange IN ('binance', 'okx', 'gate')),
    leverage    INTEGER NOT NULL DEFAULT 1,
    margin_type TEXT    NOT NULL DEFAULT 'isolated' CHECK (margin_type IN ('isolated', 'cross')),
    status      TEXT    NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at  TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (user_id, exchange, name)
);
CREATE INDEX idx_fox_symbols_exchange ON fox_symbols (exchange);
CREATE INDEX idx_fox_symbols_user_id ON fox_symbols (user_id);

CREATE TABLE fox_ss
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL DEFAULT 0,
    symbol     TEXT    NOT NULL DEFAULT '',
    side       TEXT    NOT NULL DEFAULT '' CHECK (side IN ('buy', 'sell')),
    pos_side   TEXT    NOT NULL DEFAULT '' CHECK (pos_side IN ('long', 'short', 'net')),
    px         REAL    NOT NULL DEFAULT 0,
    sz         REAL    NOT NULL DEFAULT 0,
    order_type TEXT    NOT NULL DEFAULT 'limit' CHECK (order_type IN ('limit', 'market')),
    strategy   TEXT    NOT NULL DEFAULT '',
    order_id   TEXT    NOT NULL DEFAULT '',
    type       TEXT    NOT NULL DEFAULT 'open' CHECK (type IN ('open', 'close')),
    status     TEXT    NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'pending', 'filled', 'cancelled')),
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);
CREATE INDEX idx_fox_ss_user_id ON fox_ss (user_id);
CREATE INDEX idx_fox_ss_status ON fox_ss (status);
CREATE INDEX idx_fox_ss_order_id ON fox_ss (order_id);
CREATE INDEX idx_fox_ss_symbol ON fox_ss (symbol);

-- 交易所配置表
CREATE TABLE fox_exchanges
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL DEFAULT '',
    api_url    TEXT NOT NULL DEFAULT '',
    proxy_url  TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive')),
    is_active  BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (name)
);

-- 策略配置表
CREATE TABLE fox_strategies
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    parameters  TEXT NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at  TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now', 'localtime')),
    UNIQUE (name)
);

-- 插入默认交易所配置
INSERT INTO fox_exchanges (name, api_url, proxy_url, status, is_active) VALUES 
('okx', 'https://www.okx.com', 'http://127.0.0.1:7890', 'inactive', 0),
('binance', 'https://api.binance.com', 'http://127.0.0.1:7890', 'inactive', 0),
('gate', 'https://api.gateio.ws', 'http://127.0.0.1:7890', 'inactive', 0);

-- 插入示例策略
INSERT INTO fox_strategies (name, description, parameters) VALUES 
('volume', '成交量策略', '{"threshold": 100}'),
('macd', 'MACD策略', '{"threshold": 50}'),
('rsi', 'RSI策略', '{"threshold": 10}');