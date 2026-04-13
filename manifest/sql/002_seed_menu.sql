INSERT INTO admin_menu (id, parent_id, name, code, menu_type, menu_level, status, super_only, sort, created_at, updated_at) VALUES
(1, 0, '员工管理', 'admin.list', 'permission', 1, 1, 0, 1, NOW(), NOW()),
(2, 0, '用户组与授权', 'admin.department', 'permission', 1, 1, 0, 2, NOW(), NOW()),
(3, 0, '操作日志', 'admin.action', 'permission', 1, 1, 0, 3, NOW(), NOW()),
(4, 0, '登录日志', 'admin.loginlog', 'permission', 1, 1, 0, 4, NOW(), NOW()),
(5, 0, '主体配置', 'subject.manage', 'permission', 1, 1, 0, 5, NOW(), NOW()),
(6, 0, '短信配置', 'config.sms', 'permission', 1, 1, 1, 6, NOW(), NOW()),
(9, 0, '系统参数配置', 'config.system', 'permission', 1, 1, 1, 9, NOW(), NOW()),
(10, 0, '商品模板管理', 'product.template', 'permission', 1, 1, 0, 10, NOW(), NOW()),
(11, 0, '商品购买数量限制策略', 'product.purchase_limit', 'permission', 1, 1, 0, 11, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  name = VALUES(name),
  code = VALUES(code),
  menu_type = VALUES(menu_type),
  menu_level = VALUES(menu_level),
  status = VALUES(status),
  super_only = VALUES(super_only),
  sort = VALUES(sort),
  updated_at = VALUES(updated_at);
