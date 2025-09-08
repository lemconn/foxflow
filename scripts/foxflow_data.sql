-- Foxflow 默认数据插入脚本
-- 注意：表结构由 GORM AutoMigrate 自动创建，此文件仅包含默认数据

-- 插入默认交易所配置
INSERT OR IGNORE INTO fox_exchanges (name, api_url, proxy_url, status, is_active, created_at, updated_at) VALUES 
('okx', 'https://www.okx.com', 'http://127.0.0.1:7890', 'inactive', 0, datetime('now', 'localtime'), datetime('now', 'localtime')),
('binance', 'https://api.binance.com', 'http://127.0.0.1:7890', 'inactive', 0, datetime('now', 'localtime'), datetime('now', 'localtime')),
('gate', 'https://api.gateio.ws', 'http://127.0.0.1:7890', 'inactive', 0, datetime('now', 'localtime'), datetime('now', 'localtime'));

-- 插入示例策略
INSERT OR IGNORE INTO fox_strategies (name, description, parameters, status, created_at, updated_at) VALUES 
('volume', '成交量策略', '{"threshold": 100}', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('macd', 'MACD策略', '{"threshold": 50}', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('rsi', 'RSI策略', '{"threshold": 10}', 'active', datetime('now', 'localtime'), datetime('now', 'localtime'));

-- 插入测试用户数据
INSERT OR IGNORE INTO fox_users (username, exchange, access_key, secret_key, status, trade_type, is_active, created_at, updated_at) VALUES 
('test_user_1', 'binance', 'test_binance_access_key_1', 'test_binance_secret_key_1', 'active', 'mock', 1, datetime('now', 'localtime'), datetime('now', 'localtime')),
('test_user_2', 'okx', 'test_okx_access_key_2', 'test_okx_secret_key_2', 'active', 'real', 1, datetime('now', 'localtime'), datetime('now', 'localtime')),
('test_user_3', 'gate', 'test_gate_access_key_3', 'test_gate_secret_key_3', 'inactive', 'mock', 0, datetime('now', 'localtime'), datetime('now', 'localtime')),
('demo_trader', 'binance', 'demo_binance_key', 'demo_binance_secret', 'active', 'mock', 1, datetime('now', 'localtime'), datetime('now', 'localtime'));

-- 插入测试标的数据
INSERT OR IGNORE INTO fox_symbols (name, user_id, exchange, leverage, margin_type, status, created_at, updated_at) VALUES 
('BTCUSDT', 1, 'binance', 10, 'isolated', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('ETHUSDT', 1, 'binance', 5, 'cross', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('BTC-USDT-SWAP', 2, 'okx', 20, 'isolated', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('ETH-USDT-SWAP', 2, 'okx', 15, 'cross', 'active', datetime('now', 'localtime'), datetime('now', 'localtime')),
('BTC_USDT', 3, 'gate', 8, 'isolated', 'inactive', datetime('now', 'localtime'), datetime('now', 'localtime')),
('ADAUSDT', 4, 'binance', 3, 'isolated', 'active', datetime('now', 'localtime'), datetime('now', 'localtime'));

-- 插入测试策略订单数据
INSERT OR IGNORE INTO fox_ss (user_id, symbol, side, pos_side, px, sz, order_type, strategy, order_id, type, status, created_at, updated_at) VALUES 
-- 用户1的订单
(1, 'BTCUSDT', 'buy', 'long', 45000.50, 0.01, 'limit', 'macd', 'binance_order_001', 'open', 'waiting', datetime('now', 'localtime'), datetime('now', 'localtime')),
(1, 'BTCUSDT', 'sell', 'long', 46000.00, 0.01, 'limit', 'macd', 'binance_order_002', 'close', 'pending', datetime('now', 'localtime'), datetime('now', 'localtime')),
(1, 'ETHUSDT', 'buy', 'long', 3200.25, 0.1, 'market', 'volume', 'binance_order_003', 'open', 'filled', datetime('now', 'localtime'), datetime('now', 'localtime')),

-- 用户2的订单
(2, 'BTC-USDT-SWAP', 'buy', 'long', 45100.00, 0.02, 'limit', 'rsi', 'okx_order_001', 'open', 'waiting', datetime('now', 'localtime'), datetime('now', 'localtime')),
(2, 'ETH-USDT-SWAP', 'sell', 'short', 3150.00, 0.05, 'limit', 'volume', 'okx_order_002', 'open', 'pending', datetime('now', 'localtime'), datetime('now', 'localtime')),
(2, 'BTC-USDT-SWAP', 'sell', 'long', 46000.00, 0.02, 'limit', 'rsi', 'okx_order_003', 'close', 'cancelled', datetime('now', 'localtime'), datetime('now', 'localtime')),

-- 用户4的订单
(4, 'ADAUSDT', 'buy', 'long', 0.45, 1000, 'limit', 'macd', 'binance_order_004', 'open', 'waiting', datetime('now', 'localtime'), datetime('now', 'localtime')),
(4, 'ADAUSDT', 'sell', 'long', 0.48, 1000, 'limit', 'macd', 'binance_order_005', 'close', 'waiting', datetime('now', 'localtime'), datetime('now', 'localtime')),
(4, 'BTCUSDT', 'buy', 'long', 44800.00, 0.005, 'market', 'volume', 'binance_order_006', 'open', 'filled', datetime('now', 'localtime'), datetime('now', 'localtime'));
