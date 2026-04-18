# MySQL Test Database Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将测试态应用从临时 SQLite 切换为自动创建并自动清理的 MySQL `admin_test` 数据库，避免污染日常运行使用的 `admin`。

**Architecture:** 保留现有 `NewTestCore()` 作为测试统一入口，但把数据库初始化改为 MySQL 专用测试链路：从默认 DSN 推导 `admin_test`，先连接 MySQL server 自动建库并清空旧数据，再用现有 bootstrap 回灌 schema 和种子。Redis 仍使用 `miniredis`，短信仍使用 mock sender，避免把这次改动扩大到非数据库依赖。

**Tech Stack:** Go, GoFrame gdb/gredis, MySQL 8.4, go-sql-driver/mysql, Testify

---

### Task 1: 锁定测试态数据库行为

**Files:**
- Modify: `internal/app/core_config_test.go`

- [ ] **Step 1: 写出失败测试，约束 `NewTestCore()` 必须使用 MySQL `admin_test`**

```go
func TestNewTestCore_UsesAdminTestMySQLDatabase(t *testing.T) {
    core, err := NewTestCore()
    require.NoError(t, err)
    t.Cleanup(func() { _ = core.Close() })

    require.Equal(t, "mysql", core.Config().Database.Driver)

    schema, err := core.DB().GetCore().GetValue(context.Background(), `SELECT DATABASE()`)
    require.NoError(t, err)
    require.Equal(t, "admin_test", schema.String())
}
```

- [ ] **Step 2: 运行单测，确认当前实现失败**

Run: `go test ./internal/app -run TestNewTestCore_UsesAdminTestMySQLDatabase -count=1 -timeout 60s`
Expected: FAIL，当前 `NewTestCore()` 仍把 driver 设为 `sqlite`，且无法满足 `admin_test` 断言。

- [ ] **Step 3: 再补一个失败测试，约束测试库会在每次启动前清理旧数据**

```go
func TestNewTestCore_ResetsAdminTestDataBetweenRuns(t *testing.T) {
    first, err := NewTestCore()
    require.NoError(t, err)

    ctx := context.Background()
    _, err = first.CreateTestUser(ctx, "tester01", "abc123", "13800000001")
    require.NoError(t, err)
    require.NoError(t, first.Close())

    second, err := NewTestCore()
    require.NoError(t, err)
    t.Cleanup(func() { _ = second.Close() })

    _, err = second.GetUserByUsername(ctx, "tester01")
    require.ErrorIs(t, err, sql.ErrNoRows)
}
```

- [ ] **Step 4: 运行单测，确认“清库”断言在当前实现下失败**

Run: `go test ./internal/app -run TestNewTestCore_ResetsAdminTestDataBetweenRuns -count=1 -timeout 60s`
Expected: FAIL，当前实现不会使用固定 MySQL 测试库，也不会在重建前清理旧数据。

### Task 2: 实现 MySQL `admin_test` 自动建库与清库

**Files:**
- Modify: `internal/app/core.go`
- Create: `internal/app/test_mysql.go`

- [ ] **Step 1: 抽出测试态 MySQL 配置与 DSN 推导 helper**

```go
const testMySQLDatabase = "admin_test"

func newTestMySQLConfig() modelconfig.Config {
    cfg := modelconfig.Default()
    cfg.AppEnv = "test"
    cfg.Database.Driver = "mysql"
    cfg.Database.DSN = withMySQLDatabase(cfg.Database.DSN, testMySQLDatabase)
    cfg.SMS.Provider = "mock"
    cfg.Audit.Async = false
    return cfg
}
```

- [ ] **Step 2: 在新 helper 文件里实现 DSN 解析、自动建库和清库逻辑**

```go
func ensureAndResetTestMySQLDatabase(dsn string) error {
    adminDB, err := openMySQLServer(dsn)
    if err != nil {
        return err
    }
    defer adminDB.Close()

    if _, err = adminDB.Exec("CREATE DATABASE IF NOT EXISTS `admin_test` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
        return err
    }

    testDB, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    defer testDB.Close()

    return dropAllTables(testDB, "admin_test")
}
```

- [ ] **Step 3: 改造 `NewTestCore()`，在 `newCore()` 前完成建库和清库**

```go
func NewTestCore() (*Core, error) {
    cfg := newTestMySQLConfig()
    if err := ensureAndResetTestMySQLDatabase(cfg.Database.DSN); err != nil {
        return nil, err
    }

    mr, err := miniredis.Run()
    if err != nil {
        return nil, err
    }

    core, err := newCore(cfg, g.Cfg(fmt.Sprintf("myjob-test-%d", time.Now().UnixNano())), "")
    if err != nil {
        mr.Close()
        return nil, err
    }
    core.miniRedis = mr
    return core, nil
}
```

- [ ] **Step 4: 运行 Task 1 的两个测试，确认转绿**

Run: `go test ./internal/app -run 'TestNewTestCore_(UsesAdminTestMySQLDatabase|ResetsAdminTestDataBetweenRuns)' -count=1 -timeout 60s`
Expected: PASS

### Task 3: 修正依赖 SQLite 方言的包内测试

**Files:**
- Modify: `internal/app/product_goods_schema_test.go`

- [ ] **Step 1: 把手工建表 SQL 改成 MySQL 版本，并用 `SHOW COLUMNS` 断言补列成功**

```go
_, err = core.DB().Exec(ctx, `
CREATE TABLE product_goods (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    goods_code VARCHAR(32) NOT NULL,
    brand_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    goods_type VARCHAR(32) NOT NULL,
    ...
    UNIQUE KEY uk_product_goods_code (goods_code)
)
`)

rows := make([]struct {
    Field string `db:"Field"`
}, 0)
err = core.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM product_goods`)
```

- [ ] **Step 2: 运行相关包测，确认 MySQL 方言下仍能补齐 `subject_id`**

Run: `go test ./internal/app -run TestEnsureProductGoodsSchema_AddsMissingSubjectIDColumn -count=1 -timeout 60s`
Expected: PASS

### Task 4: 同步测试文档

**Files:**
- Modify: `docs/testing.md`
- Modify: `docs/architecture.md`
- Modify: `test/contract/README.md`

- [ ] **Step 1: 更新测试说明，明确 `NewTestCore()` 改为 MySQL `admin_test`**

```md
- MySQL 测试库 `admin_test`
- `miniredis`
- mock 短信 sender
```

- [ ] **Step 2: 在架构和契约测试文档里补充“自动建库 + 每次启动前清库”的行为**

```md
- `NewTestCore()` 会自动创建 MySQL `admin_test`
- 每次启动测试应用前会清空旧表，再执行 bootstrap
```

- [ ] **Step 3: 运行文档相关测试或静态检查入口（如无专门命令则在最终说明中说明）**

Run: `rg -n "SQLite 临时文件|临时 SQLite|admin_test" docs test/contract/README.md`
Expected: 旧的 SQLite 测试描述被替换，新的 `admin_test` 描述出现。

### Task 5: 完整验证

**Files:**
- Modify: `internal/app/core.go`
- Modify: `internal/app/core_config_test.go`
- Modify: `internal/app/product_goods_schema_test.go`
- Modify: `docs/testing.md`
- Modify: `docs/architecture.md`
- Modify: `test/contract/README.md`
- Create: `internal/app/test_mysql.go`

- [ ] **Step 1: 跑完整测试**

Run: `go test ./... -count=1 -timeout 60s`
Expected: PASS

- [ ] **Step 2: 跑构建**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 3: 跑 lint**

Run: `golangci-lint run --timeout=5m`
Expected: PASS
