package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	modelconfig "myjob/internal/model/config"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/stretchr/testify/require"
)

func TestMountUploadStaticPathReturnsErrorWhenDirInitFails(t *testing.T) {
	parentFile, err := os.CreateTemp("", "myjob-upload-parent-*")
	require.NoError(t, err)
	require.NoError(t, parentFile.Close())
	t.Cleanup(func() { _ = os.Remove(parentFile.Name()) })

	server := ghttp.GetServer(fmt.Sprintf("myjob-upload-test-%d", os.Getpid()))
	t.Cleanup(func() { _ = server.Shutdown() })

	err = mountUploadStaticPath(server, modelconfig.UploadConfig{
		LocalDir:     filepath.Join(parentFile.Name(), "uploads"),
		PublicPrefix: "/uploads",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "初始化上传目录失败")
}
