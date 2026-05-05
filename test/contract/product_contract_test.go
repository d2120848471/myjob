package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
	0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestOpenAPI_ProductModulePathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/brands")
	require.Contains(t, res.body, "/api/admin/industries")
	require.Contains(t, res.body, "/api/admin/brands/upload")
	require.Contains(t, res.body, "/api/admin/product-templates")
	require.Contains(t, res.body, "/api/admin/product-templates/validate-types")
}

func TestProductModulePermissionSeedsStayInSync(t *testing.T) {
	h := newTestHarness(t)

	expectedMenus := []struct {
		id   int64
		name string
		code string
		sort int
	}{
		{id: 7, name: "品牌管理", code: "product.brand", sort: 7},
		{id: 8, name: "行业管理", code: "product.industry", sort: 8},
		{id: 10, name: "商品模板管理", code: "product.template", sort: 10},
		{id: 11, name: "商品购买数量限制策略", code: "product.purchase_limit", sort: 11},
		{id: 12, name: "第三方对接", code: "supplier.index", sort: 12},
		{id: 13, name: "商品管理", code: "product.goods", sort: 13},
		{id: 15, name: "充值风控", code: "order.recharge_risk", sort: 15},
	}
	for _, expected := range expectedMenus {
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
`, expected.id)
		require.NoError(t, err)
		require.EqualValues(t, expected.id, menu.ID)
		require.Equal(t, expected.name, menu.Name)
		require.Equal(t, expected.code, menu.Code)
		require.Equal(t, 0, menu.SuperOnly)
		require.Equal(t, expected.sort, menu.Sort)

		groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, expected.id)
		require.NoError(t, err)
		require.Equal(t, 1, groupMenuCount.Int())
	}

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	for _, expected := range expectedMenus {
		require.Contains(t, string(seedFile), fmt.Sprintf("'%s'", expected.name))
		require.Contains(t, string(seedFile), fmt.Sprintf("'%s'", expected.code))
		require.Contains(t, string(seedFile), fmt.Sprintf("(1, %d, NOW())", expected.id))
	}
}

func TestProductBrandFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	addParent := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":        0,
		"name":             "腾讯视频",
		"icon":             "",
		"credential_image": "",
		"description":      "视频会员",
		"is_visible":       1,
	}, token)
	require.Equal(t, 0, addParent.Code)

	var parentData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(addParent.Data, &parentData))
	require.NotZero(t, parentData.ID)

	addChild := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":        parentData.ID,
		"name":             "腾讯超级影视SVIP",
		"icon":             "",
		"credential_image": "",
		"description":      "二级品牌",
		"is_visible":       1,
	}, token)
	require.Equal(t, 0, addChild.Code)

	var childData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(addChild.Data, &childData))
	require.NotZero(t, childData.ID)

	addGrandChild := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  childData.ID,
		"name":       "腾讯视频影视会员周卡",
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, addGrandChild.Code)

	var grandChildData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(addGrandChild.Data, &grandChildData))
	require.NotZero(t, grandChildData.ID)

	levelErr := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  grandChildData.ID,
		"name":       "非法四级品牌",
		"is_visible": 1,
	}, token)
	require.NotEqual(t, 0, levelErr.Code)

	duplicate := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  0,
		"name":       "腾讯视频",
		"is_visible": 1,
	}, token)
	require.Equal(t, 409, duplicate.Code)

	addAnother := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  0,
		"name":       "爱奇艺",
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, addAnother.Code)

	var anotherData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(addAnother.Data, &anotherData))
	require.NotZero(t, anotherData.ID)

	listRes := h.getJSON("/api/admin/brands?page=1&page_size=20&name=腾讯", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			ID          int64 `json:"id"`
			ParentID    int64 `json:"parent_id"`
			HasChildren bool  `json:"has_children"`
			Children    []any `json:"children"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, parentData.ID, listData.List[0].ID)
	require.EqualValues(t, 0, listData.List[0].ParentID)
	require.True(t, listData.List[0].HasChildren)
	require.Empty(t, listData.List[0].Children)

	childrenRes := h.getJSON(fmt.Sprintf("/api/admin/brands/%d/children", parentData.ID), token)
	require.Equal(t, 0, childrenRes.Code)

	var childrenData struct {
		List []struct {
			ID          int64 `json:"id"`
			ParentID    int64 `json:"parent_id"`
			HasChildren bool  `json:"has_children"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(childrenRes.Data, &childrenData))
	require.Len(t, childrenData.List, 1)
	require.Equal(t, childData.ID, childrenData.List[0].ID)
	require.Equal(t, parentData.ID, childrenData.List[0].ParentID)
	require.True(t, childrenData.List[0].HasChildren)

	grandChildrenRes := h.getJSON(fmt.Sprintf("/api/admin/brands/%d/children", childData.ID), token)
	require.Equal(t, 0, grandChildrenRes.Code)

	var grandChildrenData struct {
		List []struct {
			ID          int64 `json:"id"`
			ParentID    int64 `json:"parent_id"`
			HasChildren bool  `json:"has_children"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(grandChildrenRes.Data, &grandChildrenData))
	require.Len(t, grandChildrenData.List, 1)
	require.Equal(t, grandChildData.ID, grandChildrenData.List[0].ID)
	require.Equal(t, childData.ID, grandChildrenData.List[0].ParentID)
	require.False(t, grandChildrenData.List[0].HasChildren)

	visibilityRes := h.patchJSON(fmt.Sprintf("/api/admin/brands/%d/visibility", parentData.ID), map[string]any{
		"is_visible": 0,
	}, token)
	require.Equal(t, 0, visibilityRes.Code)

	sortRes := h.patchJSON(fmt.Sprintf("/api/admin/brands/%d/sort", anotherData.ID), map[string]any{
		"action": "top",
	}, token)
	require.Equal(t, 0, sortRes.Code)

	reordered := h.getJSON("/api/admin/brands?page=1&page_size=20", token)
	require.Equal(t, 0, reordered.Code)

	var reorderedData struct {
		List []struct {
			ID        int64 `json:"id"`
			IsVisible int   `json:"is_visible"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(reordered.Data, &reorderedData))
	require.GreaterOrEqual(t, len(reorderedData.List), 2)
	require.Equal(t, anotherData.ID, reorderedData.List[0].ID)

	var parentVisible int
	for _, item := range reorderedData.List {
		if item.ID == parentData.ID {
			parentVisible = item.IsVisible
		}
	}
	require.Equal(t, 0, parentVisible)

	deleteParentConflict := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", parentData.ID), token)
	require.Equal(t, 409, deleteParentConflict.Code)

	deleteChildConflict := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", childData.ID), token)
	require.Equal(t, 409, deleteChildConflict.Code)

	deleteGrandChild := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", grandChildData.ID), token)
	require.Equal(t, 0, deleteGrandChild.Code)

	deleteChild := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", childData.ID), token)
	require.Equal(t, 0, deleteChild.Code)

	deleteParent := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", parentData.ID), token)
	require.Equal(t, 0, deleteParent.Code)
}

func TestProductIndustryFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	brand1 := h.createTopLevelBrand(t, token, "腾讯视频")
	brand2 := h.createTopLevelBrand(t, token, "优酷")
	brand3 := h.createTopLevelBrand(t, token, "爱奇艺")
	brand1Child := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  brand1,
		"name":       "腾讯视频影视会员",
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, brand1Child.Code)

	var brand1ChildData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(brand1Child.Data, &brand1ChildData))
	require.NotZero(t, brand1ChildData.ID)

	brand1GrandChild := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  brand1ChildData.ID,
		"name":       "腾讯视频影视会员周卡",
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, brand1GrandChild.Code)

	var brand1GrandChildData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(brand1GrandChild.Data, &brand1GrandChildData))
	require.NotZero(t, brand1GrandChildData.ID)

	invalidCreate := h.postJSON("/api/admin/industries", map[string]any{
		"name":      "非法行业",
		"brand_ids": []int64{0},
	}, token)
	require.Equal(t, 400, invalidCreate.Code)

	createIndustry := h.postJSON("/api/admin/industries", map[string]any{
		"name":      "视频会员",
		"brand_ids": []int64{brand1, brand2},
	}, token)
	require.Equal(t, 0, createIndustry.Code)

	var createIndustryData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createIndustry.Data, &createIndustryData))
	require.NotZero(t, createIndustryData.ID)

	listRes := h.getJSON("/api/admin/industries?page=1&page_size=20&name=视频", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			BrandCount int    `json:"brand_count"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, createIndustryData.ID, listData.List[0].ID)
	require.Equal(t, "视频会员", listData.List[0].Name)
	require.Equal(t, 2, listData.List[0].BrandCount)

	selectorRes := h.getJSON("/api/admin/industries/brand-selector?name=视频", token)
	require.Equal(t, 0, selectorRes.Code)
	require.Contains(t, string(selectorRes.Data), "腾讯视频")
	require.NotContains(t, string(selectorRes.Data), "腾讯视频影视会员")
	require.NotContains(t, string(selectorRes.Data), "腾讯视频影视会员周卡")

	brandListRes := h.getJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), token)
	require.Equal(t, 0, brandListRes.Code)

	var brandListData struct {
		IndustryID int64 `json:"industry_id"`
		List       []struct {
			BrandID int64 `json:"brand_id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(brandListRes.Data, &brandListData))
	require.Equal(t, createIndustryData.ID, brandListData.IndustryID)
	require.Len(t, brandListData.List, 2)

	invalidEdit := h.putJSON(fmt.Sprintf("/api/admin/industries/%d", createIndustryData.ID), map[string]any{
		"name":      "视频会员",
		"brand_ids": []int64{0},
	}, token)
	require.Equal(t, 400, invalidEdit.Code)

	afterInvalidEdit := h.getJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), token)
	require.Equal(t, 0, afterInvalidEdit.Code)

	var afterInvalidEditData struct {
		List []struct {
			BrandID int64 `json:"brand_id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterInvalidEdit.Data, &afterInvalidEditData))
	require.Len(t, afterInvalidEditData.List, 2)
	require.Equal(t, brand1, afterInvalidEditData.List[0].BrandID)
	require.Equal(t, brand2, afterInvalidEditData.List[1].BrandID)

	addRelation := h.postJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), map[string]any{
		"brand_ids": []int64{brand3},
	}, token)
	require.Equal(t, 0, addRelation.Code)

	addChildRelation := h.postJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), map[string]any{
		"brand_ids": []int64{brand1ChildData.ID},
	}, token)
	require.Equal(t, 400, addChildRelation.Code)

	addGrandChildRelation := h.postJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), map[string]any{
		"brand_ids": []int64{brand1GrandChildData.ID},
	}, token)
	require.Equal(t, 400, addGrandChildRelation.Code)

	sortRelation := h.patchJSON(fmt.Sprintf("/api/admin/industries/%d/brands/%d/sort", createIndustryData.ID, brand3), map[string]any{
		"action": "top",
	}, token)
	require.Equal(t, 0, sortRelation.Code)

	sortedList := h.getJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), token)
	require.Equal(t, 0, sortedList.Code)

	var sortedData struct {
		List []struct {
			BrandID int64 `json:"brand_id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(sortedList.Data, &sortedData))
	require.Len(t, sortedData.List, 3)
	require.Equal(t, brand3, sortedData.List[0].BrandID)

	deleteRelation := h.deleteJSONWithBody(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), map[string]any{
		"brand_ids": []int64{brand2},
	}, token)
	require.Equal(t, 0, deleteRelation.Code)

	deleteIndustryConflict := h.deleteJSON(fmt.Sprintf("/api/admin/industries/%d", createIndustryData.ID), token)
	require.Equal(t, 409, deleteIndustryConflict.Code)

	deleteBrandConflict := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", brand3), token)
	require.Equal(t, 409, deleteBrandConflict.Code)

	editNameOnly := h.putJSON(fmt.Sprintf("/api/admin/industries/%d", createIndustryData.ID), map[string]any{
		"name": "视频娱乐",
	}, token)
	require.Equal(t, 0, editNameOnly.Code)

	afterNameOnly := h.getJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), token)
	require.Equal(t, 0, afterNameOnly.Code)

	var afterNameOnlyData struct {
		List []struct {
			BrandID int64 `json:"brand_id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterNameOnly.Data, &afterNameOnlyData))
	require.Len(t, afterNameOnlyData.List, 2)

	editClearRelations := h.putJSON(fmt.Sprintf("/api/admin/industries/%d", createIndustryData.ID), map[string]any{
		"name":      "视频娱乐",
		"brand_ids": []int64{},
	}, token)
	require.Equal(t, 0, editClearRelations.Code)

	afterClear := h.getJSON(fmt.Sprintf("/api/admin/industries/%d/brands", createIndustryData.ID), token)
	require.Equal(t, 0, afterClear.Code)

	var afterClearData struct {
		List []struct {
			BrandID int64 `json:"brand_id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterClear.Data, &afterClearData))
	require.Empty(t, afterClearData.List)

	deleteIndustry := h.deleteJSON(fmt.Sprintf("/api/admin/industries/%d", createIndustryData.ID), token)
	require.Equal(t, 0, deleteIndustry.Code)

	deleteBrand := h.deleteJSON(fmt.Sprintf("/api/admin/brands/%d", brand3), token)
	require.Equal(t, 0, deleteBrand.Code)
}

func TestBrandUploadFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	uploadRes := h.multipartPost(t, "/api/admin/brands/upload", token, "icon", "tiny.png", tinyPNG)
	require.Equal(t, 0, uploadRes.Code)

	var uploadData struct {
		URL      string `json:"url"`
		FileName string `json:"file_name"`
		Size     int64  `json:"size"`
	}
	require.NoError(t, json.Unmarshal(uploadRes.Data, &uploadData))
	require.True(t, strings.HasPrefix(uploadData.URL, "/uploads/brands/"))
	require.NotEmpty(t, uploadData.FileName)
	require.EqualValues(t, len(tinyPNG), uploadData.Size)

	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "myjob-upload-*", "brands", "*", uploadData.FileName))
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	info, err := os.Stat(matches[0])
	require.NoError(t, err)
	require.EqualValues(t, len(tinyPNG), info.Size())

	staticRes := h.rawRequest(http.MethodGet, uploadData.URL, nil, "")
	require.Equal(t, http.StatusOK, staticRes.status)
	require.Equal(t, len(tinyPNG), len(staticRes.body))

	oversized := bytes.Repeat([]byte("a"), 3*1024*1024)
	oversizedRes := h.multipartPost(t, "/api/admin/brands/upload", token, "icon", "large.png", oversized)
	require.Equal(t, 400, oversizedRes.Code)

	invalidExtRes := h.multipartPost(t, "/api/admin/brands/upload", token, "credential", "note.txt", []byte("not-an-image"))
	require.Equal(t, 400, invalidExtRes.Code)
}

func TestProductTemplateFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	validateTypesRaw := h.rawRequest(http.MethodGet, "/api/admin/product-templates/validate-types", nil, token)
	require.Equal(t, http.StatusOK, validateTypesRaw.status)
	validateTypesRes := validateTypesRaw.env
	require.Equal(t, 0, validateTypesRes.Code)

	var validateTypesData struct {
		List []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(validateTypesRes.Data, &validateTypesData))
	require.Len(t, validateTypesData.List, 12)
	require.Equal(t, 1, validateTypesData.List[0].ID)
	require.Equal(t, "手机号", validateTypesData.List[0].Title)
	require.Equal(t, 12, validateTypesData.List[11].ID)
	require.Equal(t, "禁止填写邮箱", validateTypesData.List[11].Title)

	invalidTitle := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "   ",
		"type":          "local",
		"is_shared":     1,
		"account_name":  "账号",
		"validate_type": 1,
	}, token)
	require.Equal(t, 400, invalidTitle.Code)

	invalidAccount := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "模板A",
		"type":          "local",
		"is_shared":     1,
		"account_name":  " ",
		"validate_type": 1,
	}, token)
	require.Equal(t, 400, invalidAccount.Code)

	invalidShared := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "模板A",
		"type":          "local",
		"is_shared":     2,
		"account_name":  "账号",
		"validate_type": 1,
	}, token)
	require.Equal(t, 400, invalidShared.Code)

	invalidValidateType := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "模板A",
		"type":          "local",
		"is_shared":     1,
		"account_name":  "账号",
		"validate_type": 99,
	}, token)
	require.Equal(t, 400, invalidValidateType.Code)

	createShared := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "即梦ID",
		"type":          "local",
		"is_shared":     1,
		"account_name":  "即梦账号",
		"validate_type": 6,
	}, token)
	require.Equal(t, 0, createShared.Code)

	var createSharedData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createShared.Data, &createSharedData))
	require.NotZero(t, createSharedData.ID)

	createPrivate := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "手机号模板",
		"type":          "local",
		"is_shared":     0,
		"account_name":  "手机号",
		"validate_type": 1,
	}, token)
	require.Equal(t, 0, createPrivate.Code)

	var createPrivateData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createPrivate.Data, &createPrivateData))
	require.NotZero(t, createPrivateData.ID)

	createBatchDelete := h.postJSON("/api/admin/product-templates", map[string]any{
		"title":         "QQ模板",
		"type":          "local",
		"is_shared":     0,
		"account_name":  "QQ号",
		"validate_type": 2,
	}, token)
	require.Equal(t, 0, createBatchDelete.Code)

	var createBatchDeleteData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createBatchDelete.Data, &createBatchDeleteData))
	require.NotZero(t, createBatchDeleteData.ID)

	listRes := h.getJSON("/api/admin/product-templates?page=1&page_size=20&keyword=模板&type=local&is_shared=0", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			ID            int64  `json:"id"`
			Title         string `json:"title"`
			Type          string `json:"type"`
			TypeLabel     string `json:"type_label"`
			IsShared      int    `json:"is_shared"`
			IsSharedLabel string `json:"is_shared_label"`
			AccountName   string `json:"account_name"`
			ValidateType  int    `json:"validate_type"`
		} `json:"list"`
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Equal(t, 1, listData.Pagination.Page)
	require.Equal(t, 20, listData.Pagination.PageSize)
	require.GreaterOrEqual(t, listData.Pagination.Total, 2)

	foundPrivate := false
	for _, item := range listData.List {
		if item.ID == createPrivateData.ID {
			foundPrivate = true
			require.Equal(t, "手机号模板", item.Title)
			require.Equal(t, "local", item.Type)
			require.Equal(t, "本地模板", item.TypeLabel)
			require.Equal(t, 0, item.IsShared)
			require.Equal(t, "不共享", item.IsSharedLabel)
			require.Equal(t, "手机号", item.AccountName)
			require.Equal(t, 1, item.ValidateType)
		}
		require.NotEqual(t, createSharedData.ID, item.ID)
	}
	require.True(t, foundPrivate)

	editRes := h.putJSON(fmt.Sprintf("/api/admin/product-templates/%d", createSharedData.ID), map[string]any{
		"title":         "即梦数字ID",
		"type":          "local",
		"is_shared":     0,
		"account_name":  "即梦ID账号",
		"validate_type": 1,
	}, token)
	require.Equal(t, 0, editRes.Code)

	afterEditRes := h.getJSON("/api/admin/product-templates?page=1&page_size=20&keyword=即梦数字&type=local&is_shared=0", token)
	require.Equal(t, 0, afterEditRes.Code)

	var afterEditData struct {
		List []struct {
			ID           int64  `json:"id"`
			Title        string `json:"title"`
			AccountName  string `json:"account_name"`
			ValidateType int    `json:"validate_type"`
			IsShared     int    `json:"is_shared"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterEditRes.Data, &afterEditData))
	require.Len(t, afterEditData.List, 1)
	require.Equal(t, createSharedData.ID, afterEditData.List[0].ID)
	require.Equal(t, "即梦数字ID", afterEditData.List[0].Title)
	require.Equal(t, "即梦ID账号", afterEditData.List[0].AccountName)
	require.Equal(t, 1, afterEditData.List[0].ValidateType)
	require.Equal(t, 0, afterEditData.List[0].IsShared)

	deleteSingleRes := h.deleteJSON(fmt.Sprintf("/api/admin/product-templates/%d", createPrivateData.ID), token)
	require.Equal(t, 0, deleteSingleRes.Code)

	afterSingleDeleteRes := h.getJSON("/api/admin/product-templates?page=1&page_size=20&keyword=手机号模板", token)
	require.Equal(t, 0, afterSingleDeleteRes.Code)

	var afterSingleDeleteData struct {
		List []struct {
			ID int64 `json:"id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterSingleDeleteRes.Data, &afterSingleDeleteData))
	require.Empty(t, afterSingleDeleteData.List)

	emptyBatchDelete := h.deleteJSONWithBody("/api/admin/product-templates", map[string]any{
		"ids": []int64{},
	}, token)
	require.Equal(t, 400, emptyBatchDelete.Code)
	require.Equal(t, "请至少选择一个商品模板", emptyBatchDelete.Message)

	invalidBatchDelete := h.deleteJSONWithBody("/api/admin/product-templates", map[string]any{
		"ids": []int64{createSharedData.ID, 0},
	}, token)
	require.Equal(t, 400, invalidBatchDelete.Code)
	require.Equal(t, "模板ID必须是正整数", invalidBatchDelete.Message)

	batchDeleteRes := h.deleteJSONWithBody("/api/admin/product-templates", map[string]any{
		"ids": []int64{createSharedData.ID, createBatchDeleteData.ID},
	}, token)
	require.Equal(t, 0, batchDeleteRes.Code)

	afterBatchDeleteRes := h.getJSON("/api/admin/product-templates?page=1&page_size=20&keyword=模板&type=local", token)
	require.Equal(t, 0, afterBatchDeleteRes.Code)

	var afterBatchDeleteData struct {
		List []struct {
			ID int64 `json:"id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterBatchDeleteRes.Data, &afterBatchDeleteData))
	for _, item := range afterBatchDeleteData.List {
		require.NotEqual(t, createSharedData.ID, item.ID)
		require.NotEqual(t, createBatchDeleteData.ID, item.ID)
	}
}

func (h *testHarness) createTopLevelBrand(t *testing.T, token, name string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  0,
		"name":       name,
		"is_visible": 1,
	}, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) deleteJSONWithBody(path string, body any, token string) apiEnvelope {
	return h.request(http.MethodDelete, path, body, token)
}

func (h *testHarness) multipartPost(t *testing.T, path, token, fileType, filename string, content []byte) apiEnvelope {
	t.Helper()
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	require.NoError(t, writer.WriteField("type", fileType))
	fileWriter, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = io.Copy(fileWriter, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	var env apiEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	return env
}
