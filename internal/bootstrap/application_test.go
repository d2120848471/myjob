package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	modelconfig "myjob/internal/model/config"

	"github.com/alicebob/miniredis/v2"
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

func TestApplicationStartsAndClosesBackgroundWorkers(t *testing.T) {
	redisServer, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(redisServer.Close)

	cfg := modelconfig.Default()
	cfg.Database.Driver = "sqlite"
	cfg.Database.DSN = filepath.Join(t.TempDir(), "worker.db")
	cfg.Redis.Addr = redisServer.Addr()
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0
	cfg.Upload.LocalDir = filepath.Join(t.TempDir(), "uploads")
	cfg.OpenOrder.WorkerEnabled = true
	cfg.OpenOrder.SubmitScanIntervalSeconds = 1
	cfg.OpenOrder.PollIntervalSeconds = 1

	app, err := NewApplicationFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, app.Core())
	require.NoError(t, app.Close())
}
