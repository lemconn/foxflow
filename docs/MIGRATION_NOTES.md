# 数据库迁移说明

## 修改概述

将数据库初始化方式从手动 SQL 文件改为使用 GORM AutoMigrate，解决了 DDL 格式不匹配导致的迁移失败问题。

## 主要变更

### 1. 数据库初始化逻辑 (`internal/database/database.go`)

**之前**:
- 使用 SQL 文件创建表结构
- 手动检查表存在性
- 复杂的迁移逻辑

**现在**:
- 使用 GORM AutoMigrate 创建和迁移表
- 简化的初始化流程
- 自动处理表结构变更

### 2. 表结构创建

**之前**: 手动 SQL 文件
```sql
CREATE TABLE fox_users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL DEFAULT '',
    -- ...
);
```

**现在**: GORM 自动生成
```sql
CREATE TABLE `fox_users` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `username` text NOT NULL DEFAULT "",
    -- ...
    CONSTRAINT `chk_fox_users_status` CHECK (status IN ('active', 'inactive'))
);
```

### 3. 默认数据插入

**之前**: SQL 文件中的 INSERT 语句
**现在**: Go 代码中的 `FirstOrCreate` 方法

## 优势

1. **格式统一**: 所有表结构都由 GORM 生成，格式一致
2. **自动迁移**: 支持表结构的自动升级和字段添加
3. **类型安全**: 使用 Go 结构体定义，编译时检查
4. **维护简单**: 不需要手动维护 SQL 文件
5. **版本兼容**: 自动处理数据库版本升级

## 文件变更

### 修改的文件
- `internal/database/database.go` - 重写初始化逻辑
- `.vscode/README.md` - 更新文档

### 新增的文件
- `scripts/foxflow_data.sql` - 纯数据插入脚本（备用）

### 保留的文件
- `scripts/foxflow.sql` - 保留作为参考

## 测试验证

✅ 数据库创建成功
✅ 表结构正确生成
✅ 默认数据正确插入
✅ AutoMigrate 功能正常
✅ 调试配置工作正常

## 使用说明

1. **首次运行**: 程序会自动创建数据库和表结构
2. **数据迁移**: 自动插入默认的交易所和策略配置
3. **结构升级**: 修改模型后，AutoMigrate 会自动处理表结构变更
4. **数据安全**: 使用 `FirstOrCreate` 避免重复插入数据

## 注意事项

- 删除 `.foxflow.db` 文件会重新创建数据库
- 修改模型结构后，AutoMigrate 会自动处理迁移
- 生产环境建议使用专门的数据库迁移工具
