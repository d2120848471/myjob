package app

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var createTablePattern = regexp.MustCompile(`(?i)CREATE TABLE IF NOT EXISTS\s+([a-zA-Z0-9_]+)\s*\(`)
var tableCommentPattern = regexp.MustCompile(`(?is)\)\s*COMMENT\s*=\s*'[^']+'`)
var columnCommentPattern = regexp.MustCompile(`(?i)\bCOMMENT\b\s+'[^']+'`)

func TestMySQLSchemaIncludesTableAndColumnComments(t *testing.T) {
	assertMySQLCreateTableComments(t, "internal/app/schema.go:mysqlSchema", mysqlSchema)
}

func TestManifestMySQLSchemaFilesIncludeTableAndColumnComments(t *testing.T) {
	for _, name := range []string{
		"001_schema.sql", "005_supplier_platform.sql", "006_product_goods_channel_binding.sql",
		"007_product_goods_channel_config.sql", "009_supplier_product_push.sql",
	} {
		path := filepath.Join("..", "..", "manifest", "sql", name)
		content, err := os.ReadFile(path)
		require.NoError(t, err)

		assertMySQLCreateTableComments(t, name, string(content))
	}
}

func assertMySQLCreateTableComments(t *testing.T, source, schema string) {
	t.Helper()

	createTableCount := 0
	for _, stmt := range splitSQLStatements(schema) {
		if !strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE IF NOT EXISTS") {
			continue
		}
		createTableCount++

		match := createTablePattern.FindStringSubmatch(stmt)
		require.Lenf(t, match, 2, "%s: cannot parse table name from statement: %s", source, stmt)

		tableName := match[1]
		require.Truef(t, tableCommentPattern.MatchString(stmt), "%s: table %s must declare COMMENT", source, tableName)

		for _, line := range strings.Split(stmt, "\n") {
			trimmed := strings.TrimSpace(line)
			if !isMySQLColumnDefinitionLine(trimmed) {
				continue
			}
			require.Truef(
				t,
				columnCommentPattern.MatchString(trimmed),
				"%s: table %s column definition missing COMMENT: %s",
				source,
				tableName,
				trimmed,
			)
		}
	}

	require.Greaterf(t, createTableCount, 0, "%s: no CREATE TABLE statement found", source)
}

func splitSQLStatements(schema string) []string {
	parts := strings.Split(schema, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		statements = append(statements, trimmed)
	}
	return statements
}

func isMySQLColumnDefinitionLine(line string) bool {
	if line == "" {
		return false
	}

	upper := strings.ToUpper(line)
	switch {
	case strings.HasPrefix(upper, "CREATE TABLE"):
		return false
	case strings.HasPrefix(line, ")"):
		return false
	case strings.HasPrefix(upper, "PRIMARY KEY"):
		return false
	case strings.HasPrefix(upper, "UNIQUE KEY"):
		return false
	case strings.HasPrefix(upper, "KEY "):
		return false
	case strings.HasPrefix(upper, "CONSTRAINT "):
		return false
	default:
		return true
	}
}
