package app

import "net/http"

func (a *Application) handleMenuTree(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	items := make([]AdminMenu, 0)
	if err := a.db.SelectContext(r.Context(), &items, `
SELECT id, parent_id, name, code, menu_type, menu_level, status, super_only, sort
FROM admin_menu
WHERE status = 1 AND super_only = 0
ORDER BY sort ASC, id ASC
`); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "菜单树查询失败")
		return
	}
	writeSuccess(w, buildMenuTree(items, 0))
}
