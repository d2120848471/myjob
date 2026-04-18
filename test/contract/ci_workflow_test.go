package contract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCIWorkflow_TestJobProvidesMySQLService 约束 CI 的 test job 必须显式启动 MySQL。
//
// 这里的契约是为了锁住测试基础设施：既然测试态已经统一依赖 MySQL `admin_test`，
// 那么远端 Actions 也必须提供兼容的 3306 服务，否则 `go test ./...` 会在冷启动时直接失败。
func TestCIWorkflow_TestJobProvidesMySQLService(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", "..", ".github", "workflows", "ci.yml")
	content, err := os.ReadFile(workflowPath)
	require.NoError(t, err)

	workflow := string(content)
	require.Contains(t, workflow, "test:")
	require.Contains(t, workflow, "services:")
	require.Contains(t, workflow, "mysql:")
	require.Contains(t, workflow, "image: mysql:")
	require.Contains(t, workflow, "MYSQL_ROOT_PASSWORD: root123456")
	require.Contains(t, workflow, "MYSQL_DATABASE: admin")
	require.Contains(t, workflow, "- 3306:3306")
	require.Contains(t, workflow, "mysqladmin ping")
}
