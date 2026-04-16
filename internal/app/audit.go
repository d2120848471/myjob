package app

import (
	"context"

	modelruntime "myjob/internal/model/runtime"
)

// ResolveRegion 解析 IP 的归属地文本（用于日志/审计展示）。
func (c *Core) ResolveRegion(ip string) string {
	if c.regionResolver == nil {
		return ""
	}
	return c.regionResolver.Resolve(ip)
}

// InsertLoginLog 写入一条登录日志（包含 IP 归属地）。
func (c *Core) InsertLoginLog(ctx context.Context, userID int64, adminName, ip string) error {
	_, err := c.DB().Exec(ctx, `INSERT INTO admin_login_log (admin_id, admin_name, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?)`, userID, adminName, ip, c.ResolveRegion(ip), c.now())
	return err
}

func (c *Core) insertOperationLog(ctx context.Context, evt modelruntime.OperationEvent) error {
	_, err := c.DB().Exec(ctx, `INSERT INTO admin_operation_log (admin_id, admin_name, description, ip, ip_region, created_at) VALUES (?, ?, ?, ?, ?, ?)`, evt.AdminID, evt.AdminName, evt.Description, evt.IP, evt.IPRegion, c.now())
	return err
}

// WriteOperation 写入一条操作审计事件（优先写入 auditWriter，以支持异步/缓冲）。
func (c *Core) WriteOperation(ctx context.Context, actor AdminUser, desc, ip string) {
	evt := modelruntime.OperationEvent{AdminID: actor.ID, AdminName: actor.RealName, Description: desc, IP: ip, IPRegion: c.ResolveRegion(ip)}
	if c.auditWriter != nil {
		c.auditWriter.Write(ctx, evt)
		return
	}
	_ = c.insertOperationLog(ctx, evt)
}
