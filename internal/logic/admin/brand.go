package adminlogic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
)

type BrandLogic struct{ core *app.Core }

func (l *BrandLogic) List(ctx context.Context, req *adminapi.BrandListReq) (*adminapi.BrandListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	name := strings.TrimSpace(req.Name)
	likeName := "%" + name + "%"

	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE parent_id = 0
  AND (? = '' OR name LIKE ?)
`, name, likeName)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌列表查询失败")
	}

	type brandRow struct {
		ID              int64     `db:"id"`
		ParentID        int64     `db:"parent_id"`
		Name            string    `db:"name"`
		Icon            string    `db:"icon"`
		CredentialImage string    `db:"credential_image"`
		Description     string    `db:"description"`
		IsVisible       int       `db:"is_visible"`
		Sort            int       `db:"sort"`
		GoodsCount      int       `db:"goods_count"`
		HasChildren     int       `db:"has_children"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}
	rows := make([]brandRow, 0)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id,
    b.parent_id,
    b.name,
    b.icon,
    b.credential_image,
    COALESCE(b.description, '') AS description,
    b.is_visible,
    b.sort,
    b.goods_count,
    CASE WHEN EXISTS(SELECT 1 FROM product_brand c WHERE c.parent_id = b.id) THEN 1 ELSE 0 END AS has_children,
    b.created_at,
    b.updated_at
FROM product_brand b
WHERE b.parent_id = 0
  AND (? = '' OR b.name LIKE ?)
ORDER BY b.sort ASC, b.id ASC
LIMIT ? OFFSET ?
`, name, likeName, pageSize, (page-1)*pageSize); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌列表查询失败")
	}

	items := make([]adminapi.BrandListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.BrandListItem{
			ID:              row.ID,
			ParentID:        row.ParentID,
			Name:            row.Name,
			Icon:            row.Icon,
			CredentialImage: row.CredentialImage,
			Description:     row.Description,
			IsVisible:       row.IsVisible,
			Sort:            row.Sort,
			GoodsCount:      row.GoodsCount,
			HasChildren:     row.HasChildren == 1,
			Children:        make([]adminapi.BrandListItem, 0),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return &adminapi.BrandListRes{List: items, Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()}}, nil
}

func (l *BrandLogic) Children(ctx context.Context, req *adminapi.BrandChildrenReq) (*adminapi.BrandChildrenRes, error) {
	if _, err := l.getBrand(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	type brandRow struct {
		ID              int64     `db:"id"`
		ParentID        int64     `db:"parent_id"`
		Name            string    `db:"name"`
		Icon            string    `db:"icon"`
		CredentialImage string    `db:"credential_image"`
		Description     string    `db:"description"`
		IsVisible       int       `db:"is_visible"`
		Sort            int       `db:"sort"`
		GoodsCount      int       `db:"goods_count"`
		HasChildren     int       `db:"has_children"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}
	rows := make([]brandRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id,
    b.parent_id,
    b.name,
    b.icon,
    b.credential_image,
    COALESCE(b.description, '') AS description,
    b.is_visible,
    b.sort,
    b.goods_count,
    CASE WHEN EXISTS(SELECT 1 FROM product_brand c WHERE c.parent_id = b.id) THEN 1 ELSE 0 END AS has_children,
    b.created_at,
    b.updated_at
FROM product_brand b
WHERE b.parent_id = ?
ORDER BY sort ASC, id ASC
`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "子品牌查询失败")
	}
	items := make([]app.BrandListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, app.BrandListItem{
			ID:              row.ID,
			ParentID:        row.ParentID,
			Name:            row.Name,
			Icon:            row.Icon,
			CredentialImage: row.CredentialImage,
			Description:     row.Description,
			IsVisible:       row.IsVisible,
			Sort:            row.Sort,
			GoodsCount:      row.GoodsCount,
			HasChildren:     row.HasChildren == 1,
			Children:        make([]app.BrandListItem, 0),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return &adminapi.BrandChildrenRes{List: items}, nil
}

func (l *BrandLogic) Add(ctx context.Context, req *adminapi.BrandCreateReq, actor app.AdminUser, ip string) (*adminapi.BrandCreateRes, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	req.CredentialImage = strings.TrimSpace(req.CredentialImage)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || (req.IsVisible != 0 && req.IsVisible != 1) || req.ParentID < 0 {
		return nil, apiErr(consts.CodeBadRequest, "品牌参数错误")
	}
	_, level, err := l.validateParent(ctx, req.ParentID)
	if err != nil {
		return nil, err
	}
	exists, err := l.siblingNameExists(ctx, req.ParentID, req.Name, 0)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "同级品牌名称已存在")
	}
	sortValue, err := l.nextSort(ctx, req.ParentID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序初始化失败")
	}
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_brand (
    parent_id, name, icon, credential_image, description, is_visible, sort, goods_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
`, req.ParentID, req.Name, req.Icon, req.CredentialImage, req.Description, req.IsVisible, sortValue, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, l.buildBrandCreateLog(level, req.Name, req.ParentID), ip)
	return &adminapi.BrandCreateRes{ID: id}, nil
}

func (l *BrandLogic) Edit(ctx context.Context, req *adminapi.BrandUpdateReq, actor app.AdminUser, ip string) (*adminapi.BrandUpdateRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	req.CredentialImage = strings.TrimSpace(req.CredentialImage)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" || (req.IsVisible != 0 && req.IsVisible != 1) {
		return nil, apiErr(consts.CodeBadRequest, "品牌参数错误")
	}
	exists, err := l.siblingNameExists(ctx, brand.ParentID, req.Name, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌查询失败")
	}
	if exists {
		return nil, apiErr(consts.CodeConflict, "同级品牌名称已存在")
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_brand
SET name = ?, icon = ?, credential_image = ?, description = ?, is_visible = ?, updated_at = ?
WHERE id = ?
`, req.Name, req.Icon, req.CredentialImage, req.Description, req.IsVisible, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑品牌：%d -> %s", req.ID, req.Name), ip)
	return &adminapi.BrandUpdateRes{}, nil
}

func (l *BrandLogic) Delete(ctx context.Context, req *adminapi.BrandDeleteReq, actor app.AdminUser, ip string) (*adminapi.BrandDeleteRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	childCount, countErr := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_brand WHERE parent_id = ?`, req.ID)
	if countErr != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除校验失败")
	}
	if childCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "该品牌下存在子品牌，请先删除子品牌")
	}
	industryRefCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_industry_brand WHERE brand_id = ?`, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除校验失败")
	}
	if industryRefCount.Int() > 0 {
		return nil, apiErr(consts.CodeConflict, "该品牌已被行业关联，请先解除关联")
	}
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := tx.Exec(`DELETE FROM product_brand WHERE id = ?`, req.ID); txErr != nil {
			return txErr
		}
		orderedIDs, txErr := l.loadSiblingOrderTx(tx, brand.ParentID)
		if txErr != nil {
			return txErr
		}
		// 删除后重新压实同级排序，避免留下空洞 sort。
		return l.reorderSiblingSortTx(tx, orderedIDs)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除品牌：%d", req.ID), ip)
	return &adminapi.BrandDeleteRes{}, nil
}

func (l *BrandLogic) Sort(ctx context.Context, req *adminapi.BrandSortReq, actor app.AdminUser, ip string) (*adminapi.BrandSortRes, error) {
	brand, err := l.getBrand(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	action, err := normalizeSortAction(req.Action)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	orderedIDs, err := l.loadSiblingOrder(ctx, brand.ParentID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序读取失败")
	}
	newOrder := moveIDByAction(orderedIDs, req.ID, action)
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 品牌排序统一走“内存换位 + 事务回写”，避免 MySQL/SQLite 方言差异。
		return l.reorderSiblingSortTx(tx, newOrder)
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌排序失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("调整品牌排序：%d -> %s", req.ID, action), ip)
	return &adminapi.BrandSortRes{}, nil
}

func (l *BrandLogic) Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq, actor app.AdminUser, ip string) (*adminapi.BrandVisibilityRes, error) {
	if req.IsVisible != 0 && req.IsVisible != 1 {
		return nil, apiErr(consts.CodeBadRequest, "显示状态错误")
	}
	if _, err := l.getBrand(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	if _, err := l.core.DB().Exec(ctx, `UPDATE product_brand SET is_visible = ?, updated_at = ? WHERE id = ?`, req.IsVisible, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "品牌显隐更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换品牌显示状态：%d -> %d", req.ID, req.IsVisible), ip)
	return &adminapi.BrandVisibilityRes{}, nil
}

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

func (l *BrandLogic) buildBrandCreateLog(level int, name string, parentID int64) string {
	levelLabel := brandLevelLabel(level)
	if parentID == 0 {
		return fmt.Sprintf("添加%s品牌：%s", levelLabel, name)
	}
	return fmt.Sprintf("添加%s品牌：%s（父级ID=%d）", levelLabel, name, parentID)
}

func brandLevelLabel(level int) string {
	switch level {
	case 1:
		return "一级"
	case 2:
		return "二级"
	case 3:
		return "三级"
	default:
		return fmt.Sprintf("%d级", level)
	}
}

func (l *BrandLogic) getBrand(ctx context.Context, id int64) (app.ProductBrand, error) {
	brand := app.ProductBrand{}
	err := l.core.DB().GetCore().GetScan(ctx, &brand, `
SELECT id, parent_id, name, icon, credential_image, COALESCE(description, '') AS description,
       is_visible, sort, goods_count, created_at, updated_at
FROM product_brand
WHERE id = ?
`, id)
	return brand, err
}

func (l *BrandLogic) siblingNameExists(ctx context.Context, parentID int64, name string, excludeID int64) (bool, error) {
	count, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_brand
WHERE parent_id = ? AND name = ? AND id <> ?
`, parentID, name, excludeID)
	if err != nil {
		return false, err
	}
	return count.Int() > 0, nil
}

func (l *BrandLogic) nextSort(ctx context.Context, parentID int64) (int, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COALESCE(MAX(sort), 0) + 1 FROM product_brand WHERE parent_id = ?`, parentID)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

func (l *BrandLogic) loadSiblingOrder(ctx context.Context, parentID int64) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id FROM product_brand WHERE parent_id = ? ORDER BY sort ASC, id ASC`, parentID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (l *BrandLogic) loadSiblingOrderTx(tx gdb.TX, parentID int64) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	if err := tx.GetScan(&rows, `SELECT id FROM product_brand WHERE parent_id = ? ORDER BY sort ASC, id ASC`, parentID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (l *BrandLogic) reorderSiblingSortTx(tx gdb.TX, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.Exec(`UPDATE product_brand SET sort = ?, updated_at = ? WHERE id = ?`, index+1, l.core.Now(), id); err != nil {
			return err
		}
	}
	return nil
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
