package contract_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_ProductGoodsPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/products")
	require.Contains(t, res.body, "/api/admin/products/{id}")
	require.Contains(t, res.body, "/api/admin/products/form-options")
	require.Contains(t, res.body, "/api/admin/products/status")
}

func TestProductGoodsSeedsStayInSync(t *testing.T) {
	h := newTestHarness(t)

	var menu struct {
		ID        int64  `db:"id"`
		Name      string `db:"name"`
		Code      string `db:"code"`
		SuperOnly int    `db:"super_only"`
		Sort      int    `db:"sort"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &menu, `
SELECT id, name, code, super_only, sort
FROM admin_menu
WHERE id = ?
`, 13)
	require.NoError(t, err)
	require.EqualValues(t, 13, menu.ID)
	require.Equal(t, "商品管理", menu.Name)
	require.Equal(t, "product.goods", menu.Code)
	require.Equal(t, 0, menu.SuperOnly)
	require.Equal(t, 13, menu.Sort)

	groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, 13)
	require.NoError(t, err)
	require.Equal(t, 1, groupMenuCount.Int())

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'商品管理'")
	require.Contains(t, string(seedFile), "'product.goods'")
	require.Contains(t, string(seedFile), "INSERT INTO admin_group_menu")
	require.Contains(t, string(seedFile), "(1, 13, NOW())")
}

func TestProductGoodsCRUDAndFilters(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	adminMe := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, adminMe.Code)

	var adminMeData struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(adminMe.Data, &adminMeData))
	require.Contains(t, adminMeData.Permissions, "product.goods")

	menuTree := h.getJSON("/api/admin/menus/tree", token)
	require.Equal(t, 0, menuTree.Code)

	var menuTreeData struct {
		List []*menuTreeItem `json:"list"`
	}
	require.NoError(t, json.Unmarshal(menuTree.Data, &menuTreeData))
	require.Contains(t, flattenMenuCodes(menuTreeData.List), "product.goods")

	topA, _, leafA := h.createBrandPath(t, token, "视频会员", "腾讯视频", "腾讯视频周卡")
	_, _, leafB := h.createBrandPath(t, token, "音乐会员", "网易云音乐", "网易云黑胶月卡")
	taxSubjectID := h.createSubject(t, token, "开票主体A", 1)
	nonTaxSubjectID := h.createSubject(t, token, "普通主体B", 0)
	brandIconURL := "/uploads/brands/tencent-video.png"

	updateBrandIcon := h.putJSON("/api/admin/brands/"+int64ToString(topA), map[string]any{
		"name":             "视频会员",
		"icon":             brandIconURL,
		"credential_image": "",
		"description":      "",
		"is_visible":       1,
	}, token)
	require.Equal(t, 0, updateBrandIcon.Code)

	templateA := h.createProductTemplate(t, token, "腾讯视频模板")
	templateB := h.createProductTemplate(t, token, "网易云模板")
	strategyEnabled := h.createPurchaseLimitStrategy(t, token, "启用策略A", 1)
	strategyDisabled := h.createPurchaseLimitStrategy(t, token, "禁用策略", 0)
	strategyEnabledB := h.createPurchaseLimitStrategy(t, token, "启用策略B", 1)

	formOptionsRes := h.getJSON("/api/admin/products/form-options", token)
	require.Equal(t, 0, formOptionsRes.Code)

	var formOptionsData struct {
		Brands []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			IsLeaf   bool   `json:"is_leaf"`
			Children []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				IsLeaf   bool   `json:"is_leaf"`
				Children []struct {
					ID     int64  `json:"id"`
					Name   string `json:"name"`
					IsLeaf bool   `json:"is_leaf"`
				} `json:"children"`
			} `json:"children"`
		} `json:"brands"`
		Templates []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
		} `json:"templates"`
		PurchaseLimitStrategies []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"purchase_limit_strategies"`
		Subjects []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"subjects"`
		GoodsTypes []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"goods_types"`
		SupplyTypes []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"supply_types"`
		BooleanOptions []struct {
			Value int    `json:"value"`
			Label string `json:"label"`
		} `json:"boolean_options"`
		StatusOptions []struct {
			Value int    `json:"value"`
			Label string `json:"label"`
		} `json:"status_options"`
	}
	require.NoError(t, json.Unmarshal(formOptionsRes.Data, &formOptionsData))
	require.Len(t, formOptionsData.Brands, 2)
	require.Equal(t, topA, formOptionsData.Brands[0].ID)
	require.False(t, formOptionsData.Brands[0].IsLeaf)
	require.NotEmpty(t, formOptionsData.Brands[0].Children)
	require.False(t, formOptionsData.Brands[0].Children[0].IsLeaf)
	require.NotEmpty(t, formOptionsData.Brands[0].Children[0].Children)
	require.True(t, formOptionsData.Brands[0].Children[0].Children[0].IsLeaf)
	require.Contains(t, formOptionsData.Templates, struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{ID: templateA, Title: "腾讯视频模板"})
	require.Contains(t, formOptionsData.PurchaseLimitStrategies, struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{ID: strategyEnabled, Name: "启用策略A"})
	for _, item := range formOptionsData.PurchaseLimitStrategies {
		require.NotEqual(t, strategyDisabled, item.ID)
	}
	require.Contains(t, formOptionsData.Subjects, struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{ID: taxSubjectID, Name: "开票主体A"})
	for _, item := range formOptionsData.Subjects {
		require.NotEqual(t, nonTaxSubjectID, item.ID)
	}
	require.Len(t, formOptionsData.GoodsTypes, 2)
	require.Len(t, formOptionsData.SupplyTypes, 1)
	require.Equal(t, "channel", formOptionsData.SupplyTypes[0].Value)
	require.Len(t, formOptionsData.BooleanOptions, 2)
	require.Len(t, formOptionsData.StatusOptions, 2)

	parentBrandErr := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":         topA,
		"name":             "父级品牌非法商品",
		"goods_type":       "card_secret",
		"supply_type":      "channel",
		"is_export":        1,
		"is_douyin":        0,
		"has_tax":          1,
		"exception_notify": 1,
		"balance_limit":    "0",
		"min_purchase_qty": 1,
		"max_purchase_qty": 1,
		"status":           1,
	}, token)
	require.Equal(t, 400, parentBrandErr.Code)

	invalidSupplyType := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":         leafA,
		"name":             "错误供货方式商品",
		"goods_type":       "card_secret",
		"supply_type":      "manual",
		"is_export":        1,
		"is_douyin":        0,
		"has_tax":          1,
		"exception_notify": 1,
		"balance_limit":    "0",
		"min_purchase_qty": 1,
		"max_purchase_qty": 1,
		"status":           1,
	}, token)
	require.Equal(t, 400, invalidSupplyType.Code)

	invalidDisabledStrategy := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":                   leafA,
		"name":                       "禁用策略商品",
		"goods_type":                 "card_secret",
		"supply_type":                "channel",
		"is_export":                  1,
		"is_douyin":                  0,
		"has_tax":                    1,
		"exception_notify":           1,
		"product_template_id":        templateA,
		"purchase_limit_strategy_id": strategyDisabled,
		"balance_limit":              "0",
		"min_purchase_qty":           1,
		"max_purchase_qty":           1,
		"status":                     1,
	}, token)
	require.Equal(t, 400, invalidDisabledStrategy.Code)

	missingTaxSubject := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":         leafA,
		"name":             "缺少主体商品",
		"goods_type":       "card_secret",
		"supply_type":      "channel",
		"is_export":        1,
		"is_douyin":        0,
		"has_tax":          1,
		"exception_notify": 1,
		"balance_limit":    "0",
		"min_purchase_qty": 1,
		"max_purchase_qty": 1,
		"status":           1,
	}, token)
	require.Equal(t, 400, missingTaxSubject.Code)

	invalidNonTaxSubject := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":         leafA,
		"name":             "非含税主体商品",
		"goods_type":       "card_secret",
		"supply_type":      "channel",
		"is_export":        1,
		"is_douyin":        0,
		"has_tax":          1,
		"subject_id":       nonTaxSubjectID,
		"exception_notify": 1,
		"balance_limit":    "0",
		"min_purchase_qty": 1,
		"max_purchase_qty": 1,
		"status":           1,
	}, token)
	require.Equal(t, 400, invalidNonTaxSubject.Code)

	createRes := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":                   leafA,
		"name":                       "腾讯视频周卡商品",
		"goods_type":                 "card_secret",
		"supply_type":                "channel",
		"is_export":                  1,
		"is_douyin":                  0,
		"has_tax":                    1,
		"subject_id":                 taxSubjectID,
		"exception_notify":           1,
		"product_template_id":        templateA,
		"purchase_limit_strategy_id": strategyEnabled,
		"purchase_notice":            "购买须知A",
		"terminal_price_limit":       "29.9000",
		"balance_limit":              "0",
		"default_sell_price":         "19.9000",
		"min_purchase_qty":           1,
		"max_purchase_qty":           5,
		"status":                     1,
	}, token)
	require.Equal(t, 0, createRes.Code)

	var createData struct {
		ID        int64  `json:"id"`
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(createRes.Data, &createData))
	require.NotZero(t, createData.ID)
	require.Equal(t, fmt.Sprintf("GD%010d", createData.ID), createData.GoodsCode)

	disableBoundStrategy := h.patchJSON("/api/admin/purchase-limit-strategies/"+int64ToString(strategyEnabled)+"/status", map[string]any{
		"status": 0,
	}, token)
	require.Equal(t, 0, disableBoundStrategy.Code)

	detailRes := h.getJSON("/api/admin/products/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, detailRes.Code)

	var detailData struct {
		ID                          int64  `json:"id"`
		GoodsCode                   string `json:"goods_code"`
		BrandID                     int64  `json:"brand_id"`
		BrandName                   string `json:"brand_name"`
		Name                        string `json:"name"`
		GoodsType                   string `json:"goods_type"`
		SupplyType                  string `json:"supply_type"`
		ProductTemplateID           *int64 `json:"product_template_id"`
		ProductTemplateTitle        string `json:"product_template_title"`
		PurchaseLimitStrategyID     *int64 `json:"purchase_limit_strategy_id"`
		PurchaseLimitStrategyName   string `json:"purchase_limit_strategy_name"`
		PurchaseLimitStrategyStatus int    `json:"purchase_limit_strategy_status"`
		HasTax                      int    `json:"has_tax"`
		SubjectID                   *int64 `json:"subject_id"`
		SubjectName                 string `json:"subject_name"`
		DefaultSellPrice            string `json:"default_sell_price"`
	}
	require.NoError(t, json.Unmarshal(detailRes.Data, &detailData))
	require.Equal(t, createData.ID, detailData.ID)
	require.Equal(t, createData.GoodsCode, detailData.GoodsCode)
	require.Equal(t, leafA, detailData.BrandID)
	require.Equal(t, "腾讯视频周卡", detailData.BrandName)
	require.Equal(t, "腾讯视频周卡商品", detailData.Name)
	require.Equal(t, "card_secret", detailData.GoodsType)
	require.Equal(t, "channel", detailData.SupplyType)
	require.NotNil(t, detailData.ProductTemplateID)
	require.Equal(t, templateA, *detailData.ProductTemplateID)
	require.Equal(t, "腾讯视频模板", detailData.ProductTemplateTitle)
	require.NotNil(t, detailData.PurchaseLimitStrategyID)
	require.Equal(t, strategyEnabled, *detailData.PurchaseLimitStrategyID)
	require.Equal(t, "启用策略A", detailData.PurchaseLimitStrategyName)
	require.Equal(t, 0, detailData.PurchaseLimitStrategyStatus)
	require.Equal(t, 1, detailData.HasTax)
	require.NotNil(t, detailData.SubjectID)
	require.Equal(t, taxSubjectID, *detailData.SubjectID)
	require.Equal(t, "开票主体A", detailData.SubjectName)
	require.Equal(t, "19.9000", detailData.DefaultSellPrice)

	updateKeepDisabledStrategy := h.putJSON("/api/admin/products/"+int64ToString(createData.ID), map[string]any{
		"brand_id":                   leafA,
		"name":                       "腾讯视频周卡商品-改",
		"goods_type":                 "card_secret",
		"supply_type":                "channel",
		"is_export":                  1,
		"is_douyin":                  1,
		"has_tax":                    1,
		"subject_id":                 taxSubjectID,
		"exception_notify":           0,
		"product_template_id":        templateA,
		"purchase_limit_strategy_id": strategyEnabled,
		"purchase_notice":            "购买须知B",
		"terminal_price_limit":       "39.9000",
		"balance_limit":              "0",
		"default_sell_price":         "29.9000",
		"min_purchase_qty":           2,
		"max_purchase_qty":           6,
		"status":                     1,
	}, token)
	require.Equal(t, 0, updateKeepDisabledStrategy.Code)

	createSecondRes := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":                   leafB,
		"name":                       "网易云月卡商品",
		"goods_type":                 "direct_recharge",
		"supply_type":                "channel",
		"is_export":                  0,
		"is_douyin":                  0,
		"has_tax":                    0,
		"subject_id":                 taxSubjectID,
		"exception_notify":           1,
		"product_template_id":        nil,
		"purchase_limit_strategy_id": nil,
		"purchase_notice":            "",
		"terminal_price_limit":       "",
		"balance_limit":              "10.0000",
		"default_sell_price":         "",
		"min_purchase_qty":           1,
		"max_purchase_qty":           1,
		"status":                     0,
	}, token)
	require.Equal(t, 0, createSecondRes.Code)

	var createSecondData struct {
		ID        int64  `json:"id"`
		GoodsCode string `json:"goods_code"`
	}
	require.NoError(t, json.Unmarshal(createSecondRes.Data, &createSecondData))
	require.NotZero(t, createSecondData.ID)

	secondDetailRes := h.getJSON("/api/admin/products/"+int64ToString(createSecondData.ID), token)
	require.Equal(t, 0, secondDetailRes.Code)

	var secondDetailData struct {
		ProductTemplateID       *int64 `json:"product_template_id"`
		PurchaseLimitStrategyID *int64 `json:"purchase_limit_strategy_id"`
		SubjectID               *int64 `json:"subject_id"`
		SubjectName             string `json:"subject_name"`
	}
	require.NoError(t, json.Unmarshal(secondDetailRes.Data, &secondDetailData))
	require.Nil(t, secondDetailData.ProductTemplateID)
	require.Nil(t, secondDetailData.PurchaseLimitStrategyID)
	require.Nil(t, secondDetailData.SubjectID)
	require.Equal(t, "", secondDetailData.SubjectName)

	updateSecondStrategy := h.putJSON("/api/admin/products/"+int64ToString(createSecondData.ID), map[string]any{
		"brand_id":                   leafB,
		"name":                       "网易云月卡商品",
		"goods_type":                 "direct_recharge",
		"supply_type":                "channel",
		"is_export":                  0,
		"is_douyin":                  0,
		"has_tax":                    0,
		"subject_id":                 taxSubjectID,
		"exception_notify":           1,
		"product_template_id":        templateB,
		"purchase_limit_strategy_id": strategyEnabledB,
		"purchase_notice":            "",
		"terminal_price_limit":       "",
		"balance_limit":              "10.0000",
		"default_sell_price":         "",
		"min_purchase_qty":           1,
		"max_purchase_qty":           1,
		"status":                     0,
	}, token)
	require.Equal(t, 0, updateSecondStrategy.Code)

	updateSecondDisabledStrategy := h.putJSON("/api/admin/products/"+int64ToString(createSecondData.ID), map[string]any{
		"brand_id":                   leafB,
		"name":                       "网易云月卡商品",
		"goods_type":                 "direct_recharge",
		"supply_type":                "channel",
		"is_export":                  0,
		"is_douyin":                  0,
		"has_tax":                    0,
		"subject_id":                 taxSubjectID,
		"exception_notify":           1,
		"product_template_id":        templateB,
		"purchase_limit_strategy_id": strategyDisabled,
		"purchase_notice":            "",
		"terminal_price_limit":       "",
		"balance_limit":              "10.0000",
		"default_sell_price":         "",
		"min_purchase_qty":           1,
		"max_purchase_qty":           1,
		"status":                     0,
	}, token)
	require.Equal(t, 400, updateSecondDisabledStrategy.Code)

	listAll := h.getJSON("/api/admin/products?page=1&page_size=20", token)
	require.Equal(t, 0, listAll.Code)

	var listAllData struct {
		List []struct {
			ID                        int64    `json:"id"`
			GoodsCode                 string   `json:"goods_code"`
			BrandID                   int64    `json:"brand_id"`
			BrandName                 string   `json:"brand_name"`
			BrandIcon                 string   `json:"brand_icon"`
			SubjectID                 *int64   `json:"subject_id"`
			SubjectName               string   `json:"subject_name"`
			Name                      string   `json:"name"`
			GoodsType                 string   `json:"goods_type"`
			ProductTemplateTitle      string   `json:"product_template_title"`
			PurchaseLimitStrategyName string   `json:"purchase_limit_strategy_name"`
			BoundChannels             []string `json:"bound_channels"`
			BoundChannelCount         int      `json:"bound_channel_count"`
			PrimaryChannelName        string   `json:"primary_channel_name"`
			MinChannelCost            string   `json:"min_channel_cost"`
			ChannelAutoPriceStatus    bool     `json:"channel_auto_price_status"`
			HasTax                    int      `json:"has_tax"`
			Status                    int      `json:"status"`
		} `json:"list"`
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(listAll.Data, &listAllData))
	require.Len(t, listAllData.List, 2)
	require.Equal(t, createSecondData.ID, listAllData.List[0].ID)
	require.Equal(t, createData.ID, listAllData.List[1].ID)
	require.Equal(t, 2, listAllData.Pagination.Total)
	require.Equal(t, "网易云月卡商品", listAllData.List[0].Name)
	require.Equal(t, "腾讯视频周卡商品-改", listAllData.List[1].Name)
	require.Nil(t, listAllData.List[0].SubjectID)
	require.Equal(t, "", listAllData.List[0].SubjectName)
	require.Equal(t, brandIconURL, listAllData.List[1].BrandIcon)
	require.NotNil(t, listAllData.List[1].SubjectID)
	require.Equal(t, taxSubjectID, *listAllData.List[1].SubjectID)
	require.Equal(t, "开票主体A", listAllData.List[1].SubjectName)
	require.NotNil(t, listAllData.List[0].BoundChannels)
	require.Len(t, listAllData.List[0].BoundChannels, 0)
	require.Equal(t, 0, listAllData.List[0].BoundChannelCount)
	require.Equal(t, "", listAllData.List[0].PrimaryChannelName)
	require.Equal(t, "", listAllData.List[0].MinChannelCost)
	require.False(t, listAllData.List[0].ChannelAutoPriceStatus)
	require.NotNil(t, listAllData.List[1].BoundChannels)
	require.Len(t, listAllData.List[1].BoundChannels, 0)
	require.Equal(t, 0, listAllData.List[1].BoundChannelCount)
	require.Equal(t, "", listAllData.List[1].PrimaryChannelName)
	require.Equal(t, "", listAllData.List[1].MinChannelCost)
	require.False(t, listAllData.List[1].ChannelAutoPriceStatus)

	listByGoodsCode := h.getJSON("/api/admin/products?page=1&page_size=20&keyword="+createData.GoodsCode, token)
	require.Equal(t, 0, listByGoodsCode.Code)
	requireProductListCount(t, listByGoodsCode.Data, 1)

	listByName := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=网易云月卡", token)
	require.Equal(t, 0, listByName.Code)
	requireProductListCount(t, listByName.Data, 1)

	listByParentBrand := h.getJSON("/api/admin/products?page=1&page_size=20&brand_id="+int64ToString(topA), token)
	require.Equal(t, 0, listByParentBrand.Code)
	requireProductListCount(t, listByParentBrand.Data, 1)

	listByGoodsType := h.getJSON("/api/admin/products?page=1&page_size=20&goods_type=card_secret", token)
	require.Equal(t, 0, listByGoodsType.Code)
	requireProductListCount(t, listByGoodsType.Data, 1)

	listByHasTaxEmpty := h.getJSON("/api/admin/products?page=1&page_size=20&has_tax=", token)
	require.Equal(t, 0, listByHasTaxEmpty.Code)
	requireProductListCount(t, listByHasTaxEmpty.Data, 2)

	listByHasTaxAll := h.getJSON("/api/admin/products?page=1&page_size=20&has_tax=-1", token)
	require.Equal(t, 0, listByHasTaxAll.Code)
	requireProductListCount(t, listByHasTaxAll.Data, 2)

	listByHasTaxNo := h.getJSON("/api/admin/products?page=1&page_size=20&has_tax=0", token)
	require.Equal(t, 0, listByHasTaxNo.Code)
	requireProductListCount(t, listByHasTaxNo.Data, 1)

	listByHasTaxYes := h.getJSON("/api/admin/products?page=1&page_size=20&has_tax=1", token)
	require.Equal(t, 0, listByHasTaxYes.Code)
	requireProductListCount(t, listByHasTaxYes.Data, 1)

	listByStatusEmpty := h.getJSON("/api/admin/products?page=1&page_size=20&status=", token)
	require.Equal(t, 0, listByStatusEmpty.Code)
	requireProductListCount(t, listByStatusEmpty.Data, 2)

	listByStatusAll := h.getJSON("/api/admin/products?page=1&page_size=20&status=-1", token)
	require.Equal(t, 0, listByStatusAll.Code)
	requireProductListCount(t, listByStatusAll.Data, 2)

	listByStatusDisabled := h.getJSON("/api/admin/products?page=1&page_size=20&status=0", token)
	require.Equal(t, 0, listByStatusDisabled.Code)
	requireProductListCount(t, listByStatusDisabled.Data, 1)

	listByStatusEnabled := h.getJSON("/api/admin/products?page=1&page_size=20&status=1", token)
	require.Equal(t, 0, listByStatusEnabled.Code)
	requireProductListCount(t, listByStatusEnabled.Data, 1)

	deleteFirst := h.deleteJSON("/api/admin/products/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, deleteFirst.Code)

	listAfterDelete := h.getJSON("/api/admin/products?page=1&page_size=20", token)
	require.Equal(t, 0, listAfterDelete.Code)
	requireProductListCount(t, listAfterDelete.Data, 1)

	detailAfterDelete := h.getJSON("/api/admin/products/"+int64ToString(createData.ID), token)
	require.NotEqual(t, 0, detailAfterDelete.Code)

	limitedToken := h.createLimitedUserToken(t, token, 13)
	forbiddenRes := h.getJSON("/api/admin/products?page=1&page_size=20", limitedToken)
	require.Equal(t, 0, forbiddenRes.Code)

	limitedNoPermissionToken := h.createLimitedUserToken(t, token, 0)
	forbiddenNoPermissionRes := h.getJSON("/api/admin/products?page=1&page_size=20", limitedNoPermissionToken)
	require.Equal(t, 403, forbiddenNoPermissionRes.Code)
}

func TestProductGoodsReferenceConflicts(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leaf := h.createBrandPath(t, token, "视频权益", "优酷会员", "优酷会员周卡")
	templateID := h.createProductTemplate(t, token, "优酷模板")
	strategyID := h.createPurchaseLimitStrategy(t, token, "优酷策略", 1)
	subjectID := h.createSubject(t, token, "优酷开票主体", 1)

	createRes := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":                   leaf,
		"name":                       "优酷商品",
		"goods_type":                 "card_secret",
		"supply_type":                "channel",
		"is_export":                  1,
		"is_douyin":                  0,
		"has_tax":                    1,
		"subject_id":                 subjectID,
		"exception_notify":           1,
		"product_template_id":        templateID,
		"purchase_limit_strategy_id": strategyID,
		"purchase_notice":            "",
		"terminal_price_limit":       "",
		"balance_limit":              "0",
		"default_sell_price":         "",
		"min_purchase_qty":           1,
		"max_purchase_qty":           1,
		"status":                     1,
	}, token)
	require.Equal(t, 0, createRes.Code)

	deleteBrand := h.deleteJSON("/api/admin/brands/"+int64ToString(leaf), token)
	require.Equal(t, 409, deleteBrand.Code)

	deleteTemplate := h.deleteJSON("/api/admin/product-templates/"+int64ToString(templateID), token)
	require.Equal(t, 409, deleteTemplate.Code)

	deleteStrategy := h.deleteJSON("/api/admin/purchase-limit-strategies/"+int64ToString(strategyID), token)
	require.Equal(t, 409, deleteStrategy.Code)

	disableReferencedStrategy := h.patchJSON("/api/admin/purchase-limit-strategies/"+int64ToString(strategyID)+"/status", map[string]any{
		"status": 0,
	}, token)
	require.Equal(t, 0, disableReferencedStrategy.Code)
}

func TestProductGoodsStatusBatch(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafA := h.createBrandPath(t, token, "影视权益", "爱奇艺会员", "爱奇艺周卡")
	_, _, leafB := h.createBrandPath(t, token, "音频权益", "QQ音乐", "QQ音乐绿钻月卡")

	firstID := h.createProductGoods(t, token, leafA, "爱奇艺周卡商品", 1)
	secondID := h.createProductGoods(t, token, leafB, "QQ音乐月卡商品", 0)

	singleStatusRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{firstID},
		"status": 0,
	}, token)
	require.Equal(t, 0, singleStatusRes.Code)

	var singleStatusData struct {
		SuccessIDs   []int64 `json:"success_ids"`
		SuccessCount int     `json:"success_count"`
		FailedCount  int     `json:"failed_count"`
		Failed       []struct {
			ID     int64  `json:"id"`
			Reason string `json:"reason"`
		} `json:"failed"`
	}
	require.NoError(t, json.Unmarshal(singleStatusRes.Data, &singleStatusData))
	require.Equal(t, []int64{firstID}, singleStatusData.SuccessIDs)
	require.Equal(t, 1, singleStatusData.SuccessCount)
	require.Equal(t, 0, singleStatusData.FailedCount)
	require.Empty(t, singleStatusData.Failed)

	firstDetail := h.getJSON("/api/admin/products/"+int64ToString(firstID), token)
	require.Equal(t, 0, firstDetail.Code)
	var firstDetailData struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.Unmarshal(firstDetail.Data, &firstDetailData))
	require.Equal(t, 0, firstDetailData.Status)

	batchStatusRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{firstID, secondID, secondID},
		"status": 1,
	}, token)
	require.Equal(t, 0, batchStatusRes.Code)

	var batchStatusData struct {
		SuccessIDs   []int64 `json:"success_ids"`
		SuccessCount int     `json:"success_count"`
		FailedCount  int     `json:"failed_count"`
	}
	require.NoError(t, json.Unmarshal(batchStatusRes.Data, &batchStatusData))
	require.Equal(t, []int64{firstID, secondID}, batchStatusData.SuccessIDs)
	require.Equal(t, 2, batchStatusData.SuccessCount)
	require.Equal(t, 0, batchStatusData.FailedCount)

	listByStatusEnabled := h.getJSON("/api/admin/products?page=1&page_size=20&status=1", token)
	require.Equal(t, 0, listByStatusEnabled.Code)
	requireProductListCount(t, listByStatusEnabled.Data, 2)

	partialStatusRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{firstID, 999999},
		"status": 0,
	}, token)
	require.Equal(t, 0, partialStatusRes.Code)

	var partialStatusData struct {
		SuccessIDs   []int64 `json:"success_ids"`
		SuccessCount int     `json:"success_count"`
		FailedCount  int     `json:"failed_count"`
		Failed       []struct {
			ID     int64  `json:"id"`
			Reason string `json:"reason"`
		} `json:"failed"`
	}
	require.NoError(t, json.Unmarshal(partialStatusRes.Data, &partialStatusData))
	require.Equal(t, []int64{firstID}, partialStatusData.SuccessIDs)
	require.Equal(t, 1, partialStatusData.SuccessCount)
	require.Equal(t, 1, partialStatusData.FailedCount)
	require.Len(t, partialStatusData.Failed, 1)
	require.Equal(t, int64(999999), partialStatusData.Failed[0].ID)
	require.Equal(t, "商品不存在", partialStatusData.Failed[0].Reason)

	listByStatusDisabled := h.getJSON("/api/admin/products?page=1&page_size=20&status=0", token)
	require.Equal(t, 0, listByStatusDisabled.Code)
	requireProductListCount(t, listByStatusDisabled.Data, 1)

	emptyIDsRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{},
		"status": 1,
	}, token)
	require.Equal(t, 400, emptyIDsRes.Code)

	invalidIDRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{0},
		"status": 1,
	}, token)
	require.Equal(t, 400, invalidIDRes.Code)

	invalidStatusRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{firstID},
		"status": 2,
	}, token)
	require.Equal(t, 400, invalidStatusRes.Code)

	limitedToken := h.createLimitedUserToken(t, token, 0)
	forbiddenRes := h.patchJSON("/api/admin/products/status", map[string]any{
		"ids":    []int64{firstID},
		"status": 1,
	}, limitedToken)
	require.Equal(t, 403, forbiddenRes.Code)
}

func requireProductListCount(t *testing.T, raw json.RawMessage, want int) {
	t.Helper()
	var payload struct {
		List []json.RawMessage `json:"list"`
	}
	require.NoError(t, json.Unmarshal(raw, &payload))
	require.Len(t, payload.List, want)
}

func (h *testHarness) createBrandPath(t *testing.T, token, topName, childName, leafName string) (int64, int64, int64) {
	t.Helper()
	topID := h.createTopLevelBrand(t, token, topName)

	childRes := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  topID,
		"name":       childName,
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, childRes.Code)

	var childData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(childRes.Data, &childData))
	require.NotZero(t, childData.ID)

	leafRes := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  childData.ID,
		"name":       leafName,
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, leafRes.Code)

	var leafData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(leafRes.Data, &leafData))
	require.NotZero(t, leafData.ID)

	return topID, childData.ID, leafData.ID
}

func (h *testHarness) createProductTemplate(t *testing.T, token, title string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         title,
		"type":          "local",
		"is_shared":     1,
		"account_name":  "充值账号",
		"validate_type": 1,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) createPurchaseLimitStrategy(t *testing.T, token, name string, status int) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        name,
		"limit_type":  1,
		"period_type": 1,
		"period":      1,
		"limit_nums":  1,
		"limit_times": 1,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)

	if status != 1 {
		statusRes := h.patchJSON("/api/admin/purchase-limit-strategies/"+int64ToString(data.ID)+"/status", map[string]any{
			"status": status,
		}, token)
		require.Equal(t, 0, statusRes.Code)
	}
	return data.ID
}

func (h *testHarness) createProductGoods(t *testing.T, token string, brandID int64, name string, status int) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":         brandID,
		"name":             name,
		"goods_type":       "card_secret",
		"supply_type":      "channel",
		"is_export":        1,
		"is_douyin":        0,
		"has_tax":          0,
		"exception_notify": 1,
		"balance_limit":    "0",
		"min_purchase_qty": 1,
		"max_purchase_qty": 1,
		"status":           status,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) createSubject(t *testing.T, token, name string, hasTax int) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/subjects", map[string]any{
		"name":    name,
		"has_tax": hasTax,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) createLimitedUserToken(t *testing.T, adminToken string, menuID int64) string {
	t.Helper()
	groupName := "商品受限组"
	userName := "goodsv1"
	phone := "13800005555"
	if menuID == 0 {
		groupName = "无商品权限组"
		userName = "goodsb1"
		phone = "13800006666"
	}

	createGroup := h.postJSON("/api/admin/groups", map[string]any{
		"name":        groupName,
		"description": groupName,
	}, adminToken)
	require.Equal(t, 0, createGroup.Code)

	var groupData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createGroup.Data, &groupData))
	require.NotZero(t, groupData.ID)

	menuIDs := []int64{}
	if menuID > 0 {
		menuIDs = append(menuIDs, menuID)
	}
	saveAuth := h.patchJSON("/api/admin/groups/"+int64ToString(groupData.ID)+"/permissions", map[string]any{
		"menu_ids": menuIDs,
	}, adminToken)
	require.Equal(t, 0, saveAuth.Code)

	createUser := h.postJSON("/api/admin/users", map[string]any{
		"username":         userName,
		"confirm_username": userName,
		"password":         "Goods_123",
		"confirm_password": "Goods_123",
		"real_name":        groupName,
		"phone":            phone,
		"group_id":         groupData.ID,
	}, adminToken)
	require.Equal(t, 0, createUser.Code)

	loginRes := h.postJSON("/api/admin/auth/login", map[string]any{
		"username": userName,
		"password": "Goods_123",
	}, "")
	require.Equal(t, 0, loginRes.Code)

	var loginData struct {
		Token         string `json:"token"`
		NeedSMSVerify bool   `json:"need_sms_verify"`
		LoginToken    string `json:"login_token"`
	}
	require.NoError(t, json.Unmarshal(loginRes.Data, &loginData))
	if !loginData.NeedSMSVerify {
		require.NotEmpty(t, loginData.Token)
		return loginData.Token
	}

	sendRes := h.postJSON("/api/admin/auth/sms/send", map[string]any{
		"login_token": loginData.LoginToken,
	}, "")
	require.Equal(t, 0, sendRes.Code)
	code := h.lastSMSCode(t, phone)

	verifyRes := h.postJSON("/api/admin/auth/sms/verify", map[string]any{
		"login_token": loginData.LoginToken,
		"sms_code":    code,
	}, "")
	require.Equal(t, 0, verifyRes.Code)

	var verifyData struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(verifyRes.Data, &verifyData))
	require.NotEmpty(t, verifyData.Token)
	return verifyData.Token
}
