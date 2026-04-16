package adminlogic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/net/ghttp"
)

func (l *BrandLogic) validateParent(ctx context.Context, parentID int64) (app.ProductBrand, int, error) {
	if parentID == 0 {
		return app.ProductBrand{}, 1, nil
	}
	parent, err := l.getBrand(ctx, parentID)
	if err != nil {
		return app.ProductBrand{}, 0, apiErr(consts.CodeBadRequest, "父级品牌不存在")
	}
	// 沿着父链回溯层级，最多允许新增到三级品牌。
	parentLevel, err := l.brandLevel(ctx, parent)
	if err != nil {
		return app.ProductBrand{}, 0, err
	}
	if parentLevel >= 3 {
		return app.ProductBrand{}, 0, apiErr(consts.CodeBadRequest, "品牌层级最多支持三级")
	}
	return parent, parentLevel + 1, nil
}

func (l *BrandLogic) brandLevel(ctx context.Context, brand app.ProductBrand) (int, error) {
	level := 1
	visited := map[int64]struct{}{}
	current := brand
	for current.ParentID != 0 {
		if _, ok := visited[current.ID]; ok {
			return 0, apiErr(consts.CodeBadRequest, "父级品牌层级异常")
		}
		visited[current.ID] = struct{}{}
		parent, err := l.getBrand(ctx, current.ParentID)
		if err != nil {
			return 0, apiErr(consts.CodeBadRequest, "父级品牌不存在")
		}
		level++
		current = parent
	}
	return level, nil
}

func (l *BrandLogic) validateUploadType(uploadType string) error {
	uploadType = strings.TrimSpace(strings.ToLower(uploadType))
	if uploadType == "" || uploadType == "icon" || uploadType == "credential" {
		return nil
	}
	return apiErr(consts.CodeBadRequest, "图片用途错误")
}

func (l *BrandLogic) validateUploadFile(file *ghttp.UploadFile) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]struct{}{".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}}
	if _, ok := allowedExt[ext]; !ok {
		return "", apiErr(consts.CodeBadRequest, "仅支持 jpg/jpeg/png/gif/webp 图片")
	}
	maxSize := int64(l.core.Config().Upload.MaxImageSizeMB) * 1024 * 1024
	if file.Size > maxSize {
		return "", apiErr(consts.CodeBadRequest, fmt.Sprintf("图片大小不能超过 %dMB", l.core.Config().Upload.MaxImageSizeMB))
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(header[:n])
	allowedContentTypes := map[string]struct{}{"image/jpeg": {}, "image/png": {}, "image/gif": {}, "image/webp": {}}
	if _, ok := allowedContentTypes[contentType]; !ok {
		return "", apiErr(consts.CodeBadRequest, "仅支持 jpg/jpeg/png/gif/webp 图片")
	}
	return ext, nil
}
