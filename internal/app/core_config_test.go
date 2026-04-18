package app

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	modelconfig "myjob/internal/model/config"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoadConfig_UsesFixedBootstrapAdminFromLocalConfig(t *testing.T) {
	t.Setenv("SUPER_ADMIN_PHONE", "")
	t.Setenv("SUPER_ADMIN_PASSWORD", "")

	_, _, cfg, err := loadConfig("../../manifest/config/config.local.yaml")
	require.NoError(t, err)

	require.Equal(t, "15881767197", cfg.Bootstrap.SuperAdminPhone)
	require.Equal(t, "abc123", cfg.Bootstrap.SuperAdminPassword)
}

func TestDefaultConfig_UsesLocalMySQLPort3306(t *testing.T) {
	cfg := modelconfig.Default()
	require.Equal(t, "root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local", cfg.Database.DSN)
}

func TestNewTestCore_UsesAdminTestMySQLDatabase(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	require.Equal(t, "mysql", core.Config().Database.Driver)

	schema, err := core.DB().GetCore().GetValue(context.Background(), `SELECT DATABASE()`)
	require.NoError(t, err)
	require.Equal(t, "admin_test", schema.String())
}

func TestNewTestCore_ResetsAdminTestDataBetweenRuns(t *testing.T) {
	db := openAdminTestMySQL(t)
	_, err := db.Exec(`DROP TABLE IF EXISTS test_residue`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE test_residue (id BIGINT PRIMARY KEY)`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS test_residue`)
		_ = db.Close()
	})

	second, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = second.Close() })

	count, err := second.DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = DATABASE() AND table_name = ?
`, "test_residue")
	require.NoError(t, err)
	require.EqualValues(t, 0, count.Int64())
}

func TestNewTestCore_HoldsAdminTestMySQLLock(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	lockDB, lockConn := openAdminTestMySQLLockConn(t)
	defer func() {
		_ = lockConn.Close()
		_ = lockDB.Close()
	}()

	var locked sql.NullInt64
	err = lockConn.QueryRowContext(
		context.Background(),
		`SELECT GET_LOCK(?, 0)`,
		"myjob_admin_test_lock",
	).Scan(&locked)
	require.NoError(t, err)
	require.EqualValues(t, 0, locked.Int64)
}

func TestNewTestCore_CreatesAdminTestDatabaseWhenMissing(t *testing.T) {
	serverDB := openMySQLServerConn(t)
	_, err := serverDB.Exec(`DROP DATABASE IF EXISTS ` + "`" + testMySQLDatabase + "`")
	require.NoError(t, err)

	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	schema, err := core.DB().GetCore().GetValue(context.Background(), `SELECT DATABASE()`)
	require.NoError(t, err)
	require.Equal(t, testMySQLDatabase, schema.String())
}

func openAdminTestMySQL(t *testing.T) *sql.DB {
	t.Helper()

	baseCfg, err := mysql.ParseDSN(modelconfig.Default().Database.DSN)
	require.NoError(t, err)

	serverCfg := *baseCfg
	serverCfg.DBName = ""
	serverDB, err := sql.Open("mysql", serverCfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverDB.Close() })
	require.NoError(t, serverDB.Ping())
	_, err = serverDB.Exec("CREATE DATABASE IF NOT EXISTS `admin_test` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	require.NoError(t, err)

	testCfg := *baseCfg
	testCfg.DBName = "admin_test"
	db, err := sql.Open("mysql", testCfg.FormatDSN())
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func openAdminTestMySQLLockConn(t *testing.T) (*sql.DB, *sql.Conn) {
	t.Helper()

	baseCfg, err := mysql.ParseDSN(modelconfig.Default().Database.DSN)
	require.NoError(t, err)
	baseCfg.DBName = ""

	db, err := sql.Open("mysql", baseCfg.FormatDSN())
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	conn, err := db.Conn(context.Background())
	require.NoError(t, err)
	return db, conn
}

func openMySQLServerConn(t *testing.T) *sql.DB {
	t.Helper()

	baseCfg, err := mysql.ParseDSN(modelconfig.Default().Database.DSN)
	require.NoError(t, err)
	baseCfg.DBName = ""

	db, err := sql.Open("mysql", baseCfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())
	return db
}

func TestDefaultConfigPath_FindsRepoConfigFromNestedDir(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	nestedDir := filepath.Join(wd, "..", "..", "test", "integration")
	require.NoError(t, os.Chdir(nestedDir))
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	path, err := defaultConfigPath()
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(path))
	require.True(t, strings.HasSuffix(path, filepath.Join("manifest", "config", "config.local.yaml")))
}

func TestEnsureSuperAdmin_UpdatesExistingAdminCredentials(t *testing.T) {
	core, err := NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })

	ctx := context.Background()
	legacyHash, err := bcryptGenerate("legacy123")
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
UPDATE admin_user
SET phone = ?, password_hash = ?, updated_at = ?
WHERE username = ?
`, "13800000000", legacyHash, core.Now(), "admin")
	require.NoError(t, err)

	core.cfg.Bootstrap.SuperAdminPhone = "15881767197"
	core.cfg.Bootstrap.SuperAdminPassword = "abc123"
	require.NoError(t, core.ensureSuperAdmin(ctx))

	user, err := core.GetUserByUsername(ctx, "admin")
	require.NoError(t, err)
	require.Equal(t, "15881767197", user.Phone)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("abc123")))
}
