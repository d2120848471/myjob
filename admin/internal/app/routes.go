package app

func (a *Application) registerRoutes() {
	a.mux.HandleFunc("POST /api/admin/login", a.handleLogin)
	a.mux.HandleFunc("POST /api/admin/login/sms/send", a.handleLoginSMSSend)
	a.mux.HandleFunc("POST /api/admin/login/sms/verify", a.handleLoginSMSVerify)
	a.mux.HandleFunc("POST /api/admin/me", a.withAuth("", false, a.handleMe))
	a.mux.HandleFunc("POST /api/admin/logout", a.withAuth("", false, a.handleLogout))

	a.mux.HandleFunc("GET /api/admin/user/list", a.withAuth("admin.list", false, a.handleUserList))
	a.mux.HandleFunc("POST /api/admin/user/add", a.withAuth("admin.list", false, a.handleUserAdd))
	a.mux.HandleFunc("PUT /api/admin/user/{id}", a.withAuth("admin.list", false, a.handleUserEdit))
	a.mux.HandleFunc("DELETE /api/admin/user/{id}", a.withAuth("admin.list", false, a.handleUserDelete))
	a.mux.HandleFunc("PUT /api/admin/user/{id}/status", a.withAuth("admin.list", false, a.handleUserStatus))
	a.mux.HandleFunc("PUT /api/admin/user/{id}/notify", a.withAuth("admin.list", false, a.handleUserNotify))
	a.mux.HandleFunc("POST /api/admin/user/setBusiness", a.withAuth("admin.list", false, a.handleUserSetBusiness))
	a.mux.HandleFunc("POST /api/admin/user/cancelBusiness", a.withAuth("admin.list", false, a.handleUserCancelBusiness))
	a.mux.HandleFunc("GET /api/admin/user/trash", a.withAuth("admin.list", false, a.handleUserTrash))
	a.mux.HandleFunc("PUT /api/admin/user/{id}/restore", a.withAuth("admin.list", false, a.handleUserRestore))

	a.mux.HandleFunc("GET /api/admin/subject/list", a.withAuth("subject.manage", false, a.handleSubjectList))
	a.mux.HandleFunc("POST /api/admin/subject/add", a.withAuth("subject.manage", false, a.handleSubjectAdd))
	a.mux.HandleFunc("PUT /api/admin/subject/{id}", a.withAuth("subject.manage", false, a.handleSubjectEdit))

	a.mux.HandleFunc("GET /api/admin/group/list", a.withAuth("admin.department", false, a.handleGroupList))
	a.mux.HandleFunc("POST /api/admin/group/add", a.withAuth("admin.department", false, a.handleGroupAdd))
	a.mux.HandleFunc("PUT /api/admin/group/{id}", a.withAuth("admin.department", false, a.handleGroupEdit))
	a.mux.HandleFunc("DELETE /api/admin/group/{id}", a.withAuth("admin.department", false, a.handleGroupDelete))
	a.mux.HandleFunc("PUT /api/admin/group/{id}/status", a.withAuth("admin.department", false, a.handleGroupStatus))
	a.mux.HandleFunc("GET /api/admin/group/{id}/auth", a.withAuth("admin.department", false, a.handleGroupAuthGet))
	a.mux.HandleFunc("PUT /api/admin/group/{id}/auth", a.withAuth("admin.department", false, a.handleGroupAuthSave))
	a.mux.HandleFunc("GET /api/admin/menu/tree", a.withAuth("admin.department", false, a.handleMenuTree))

	a.mux.HandleFunc("GET /api/admin/config/sms", a.withAuth("", true, a.handleSMSConfigGet))
	a.mux.HandleFunc("PUT /api/admin/config/sms", a.withAuth("", true, a.handleSMSConfigSave))
	a.mux.HandleFunc("GET /api/admin/log/operation", a.withAuth("admin.action", false, a.handleOperationLogList))
	a.mux.HandleFunc("GET /api/admin/log/login", a.withAuth("admin.loginlog", false, a.handleLoginLogList))
}
