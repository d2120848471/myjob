package contract_test

import (
	"bytes"
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

	levelErr := h.postJSON("/api/admin/brands", map[string]any{
		"parent_id":  childData.ID,
		"name":       "非法三级品牌",
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
	require.False(t, childrenData.List[0].HasChildren)

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
	require.NotContains(t, string(selectorRes.Data), "二级")

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
