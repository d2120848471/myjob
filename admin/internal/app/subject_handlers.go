package app

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *Application) handleSubjectList(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	items := make([]AdminSubject, 0)
	if err := a.db.SelectContext(r.Context(), &items, `SELECT id, name, has_tax, created_at, updated_at FROM admin_subject ORDER BY id DESC`); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "主体列表查询失败")
		return
	}
	writeSuccess(w, map[string]interface{}{"list": items})
}

func (a *Application) handleSubjectAdd(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	var req struct {
		Name   string `json:"name"`
		HasTax int    `json:"has_tax"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		writeError(w, http.StatusBadRequest, 400, "主体参数错误")
		return
	}
	var exists int
	if err := a.db.GetContext(r.Context(), &exists, `SELECT COUNT(*) FROM admin_subject WHERE name = ?`, req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "主体查询失败")
		return
	}
	if exists > 0 {
		writeError(w, http.StatusConflict, 409, "主体名称已存在")
		return
	}
	result, err := a.db.ExecContext(r.Context(), `INSERT INTO admin_subject (name, has_tax, created_at, updated_at) VALUES (?, ?, ?, ?)`, req.Name, req.HasTax, a.now(), a.now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "主体新增失败")
		return
	}
	id, _ := result.LastInsertId()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("添加主体：%s", req.Name), requestIP(r))
	writeSuccess(w, map[string]interface{}{"id": id})
}

func (a *Application) handleSubjectEdit(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "主体ID错误")
		return
	}
	var req struct {
		Name   string `json:"name"`
		HasTax int    `json:"has_tax"`
	}
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || (req.HasTax != 0 && req.HasTax != 1) {
		writeError(w, http.StatusBadRequest, 400, "主体参数错误")
		return
	}
	var exists int
	if err = a.db.GetContext(r.Context(), &exists, `SELECT COUNT(*) FROM admin_subject WHERE name = ? AND id <> ?`, req.Name, id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "主体查询失败")
		return
	}
	if exists > 0 {
		writeError(w, http.StatusConflict, 409, "主体名称已存在")
		return
	}
	if _, err = a.db.ExecContext(r.Context(), `UPDATE admin_subject SET name = ?, has_tax = ?, updated_at = ? WHERE id = ?`, req.Name, req.HasTax, a.now(), id); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "主体编辑失败")
		return
	}
	a.writeOperation(r.Context(), actor, fmt.Sprintf("编辑主体：%s", req.Name), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}
