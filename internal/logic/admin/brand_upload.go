package adminlogic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/net/ghttp"
)

// Upload 上传品牌图片到本地存储，并返回对外可访问 URL。
func (l *BrandLogic) Upload(ctx context.Context, req *adminapi.BrandUploadReq, actor app.AdminUser, ip string) (*adminapi.BrandUploadRes, error) {
	if err := l.validateUploadType(req.Type); err != nil {
		return nil, err
	}
	if req.File == nil {
		return nil, apiErr(consts.CodeBadRequest, "请上传图片")
	}
	ext, err := l.validateUploadFile(req.File)
	if err != nil {
		return nil, err
	}
	dateDir := l.core.Now().Format("20060102")
	dstDir := filepath.Join(l.core.Config().Upload.LocalDir, "brands", dateDir)
	if err = l.ensureUploadDir(dstDir); err != nil {
		return nil, apiErr(consts.CodeInternalError, "上传目录创建失败")
	}
	fileName, err := l.buildUploadFilename(ext)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "上传文件名生成失败")
	}
	dstPath := filepath.Join(dstDir, fileName)
	if err = l.saveUploadedFile(req.File, dstPath); err != nil {
		return nil, apiErr(consts.CodeInternalError, "图片保存失败")
	}
	publicURL := l.buildPublicURL(path.Join("brands", dateDir, fileName))
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("上传品牌图片：%s", fileName), ip)
	return &adminapi.BrandUploadRes{URL: publicURL, FileName: fileName, Size: req.File.Size}, nil
}

func (l *BrandLogic) ensureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func (l *BrandLogic) buildUploadFilename(ext string) (string, error) {
	var randomPart [4]byte
	if _, err := rand.Read(randomPart[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%s%s", l.core.Now().Unix(), hex.EncodeToString(randomPart[:]), ext), nil
}

func (l *BrandLogic) saveUploadedFile(file *ghttp.UploadFile, dstPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func (l *BrandLogic) buildPublicURL(relativePath string) string {
	prefix := strings.TrimRight(l.core.Config().Upload.PublicPrefix, "/")
	return prefix + "/" + strings.TrimLeft(path.Clean(relativePath), "/")
}
