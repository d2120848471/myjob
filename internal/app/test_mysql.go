package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
	modelconfig "myjob/internal/model/config"
)

const testMySQLDatabase = "admin_test"
const testMySQLLockName = "myjob_admin_test_lock"

// newTestMySQLConfig 返回测试态默认配置。
//
// 测试库固定落到 MySQL `admin_test`，避免污染日常运行使用的 `admin`。
func newTestMySQLConfig() (modelconfig.Config, error) {
	cfg := modelconfig.Default()
	cfg.AppEnv = "test"
	cfg.Database.Driver = "mysql"
	cfg.SMS.Provider = "mock"
	cfg.Audit.Async = false

	dsn, err := withMySQLDatabase(cfg.Database.DSN, testMySQLDatabase)
	if err != nil {
		return modelconfig.Config{}, err
	}
	cfg.Database.DSN = dsn
	return cfg, nil
}

// ensureAndResetTestMySQLDatabase 确保测试库存在，并在每次测试启动前清空旧表。
//
// 这里显式清库，而不是依赖事务回滚，是因为契约测试和应用级测试会跨多次请求写入数据。
func prepareTestMySQLDatabase(dsn string) (*sql.DB, *sql.Conn, error) {
	lockDB, lockConn, err := acquireTestMySQLLock(dsn)
	if err != nil {
		return nil, nil, err
	}
	if err = ensureAndResetTestMySQLDatabase(dsn); err != nil {
		_ = lockConn.Close()
		_ = lockDB.Close()
		return nil, nil, err
	}
	return lockDB, lockConn, nil
}

func ensureAndResetTestMySQLDatabase(dsn string) error {
	serverDB, err := openMySQLServer(dsn)
	if err != nil {
		return err
	}
	defer serverDB.Close()

	if _, err = serverDB.Exec(
		fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			quoteMySQLIdentifier(testMySQLDatabase),
		),
	); err != nil {
		return err
	}

	testDB, schemaName, err := openMySQLDatabase(dsn)
	if err != nil {
		return err
	}
	defer testDB.Close()

	return resetMySQLSchema(testDB, schemaName)
}

// acquireTestMySQLLock 通过 MySQL 命名锁串行化 `admin_test` 的生命周期。
//
// `go test ./...` 会并发跑多个 package；如果多个测试进程同时重置同一测试库，
// 就会在 bootstrap 种子写入阶段互相踩数据，所以这里必须在测试 Core 整个生命周期内持锁。
func acquireTestMySQLLock(dsn string) (*sql.DB, *sql.Conn, error) {
	db, err := openMySQLServer(dsn)
	if err != nil {
		return nil, nil, err
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		db.Close()
		return nil, nil, err
	}

	var locked sql.NullInt64
	if err = conn.QueryRowContext(context.Background(), `SELECT GET_LOCK(?, 60)`, testMySQLLockName).Scan(&locked); err != nil {
		conn.Close()
		db.Close()
		return nil, nil, err
	}
	if !locked.Valid || locked.Int64 != 1 {
		conn.Close()
		db.Close()
		return nil, nil, fmt.Errorf("获取 MySQL 测试库锁失败: %s", testMySQLLockName)
	}
	return db, conn, nil
}

func withMySQLDatabase(dsn, database string) (string, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	cfg.DBName = database
	return cfg.FormatDSN(), nil
}

func openMySQLServer(dsn string) (*sql.DB, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	cfg.DBName = ""
	return openMySQLServerWithDB(cfg.FormatDSN())
}

func openMySQLServerWithDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func openMySQLDatabase(dsn string) (*sql.DB, string, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, "", err
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, "", err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, "", err
	}
	return db, cfg.DBName, nil
}

func resetMySQLSchema(db *sql.DB, schemaName string) (err error) {
	rows, err := db.Query(`
SELECT table_name
FROM information_schema.tables
WHERE table_schema = ?
ORDER BY table_name ASC
`, schemaName)
	if err != nil {
		return err
	}
	defer rows.Close()

	tableNames := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			return err
		}
		tableNames = append(tableNames, tableName)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if len(tableNames) == 0 {
		return nil
	}

	if _, err = db.Exec(`SET FOREIGN_KEY_CHECKS = 0`); err != nil {
		return err
	}
	defer func() {
		_, restoreErr := db.Exec(`SET FOREIGN_KEY_CHECKS = 1`)
		if err == nil {
			err = restoreErr
		}
	}()

	for _, tableName := range tableNames {
		if _, err = db.Exec(`DROP TABLE IF EXISTS ` + quoteMySQLIdentifier(tableName)); err != nil {
			return err
		}
	}
	return nil
}

func quoteMySQLIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
