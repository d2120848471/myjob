package adminlogic

import (
	"database/sql"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
)

func customerTimeString(value sql.NullTime) string {
	if !value.Valid || value.Time.IsZero() {
		return ""
	}
	return value.Time.In(time.Local).Format("2006-01-02 15:04:05")
}

func mapCustomerDetail(customer app.CustomerUser) *adminapi.CustomerDetailRes {
	return &adminapi.CustomerDetailRes{
		ID:          customer.ID,
		CompanyName: customer.CompanyName,
		Phone:       customer.Phone,
		Status:      customer.Status,
		LastLoginIP: customer.LastLoginIP,
		LastLoginAt: customerTimeString(customer.LastLoginAt),
		CreatedAt:   customer.CreatedAt.In(time.Local).Format("2006-01-02 15:04:05"),
		UpdatedAt:   customer.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05"),
	}
}

type customerListRow struct {
	ID          int64        `db:"id"`
	CompanyName string       `db:"company_name"`
	Phone       string       `db:"phone"`
	Status      int          `db:"status"`
	LastLoginIP string       `db:"last_login_ip"`
	LastLoginAt sql.NullTime `db:"last_login_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

func mapCustomerListRows(rows []customerListRow) []adminapi.CustomerListItem {
	items := make([]adminapi.CustomerListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.CustomerListItem{
			ID:          row.ID,
			CompanyName: row.CompanyName,
			Phone:       row.Phone,
			Status:      row.Status,
			LastLoginIP: row.LastLoginIP,
			LastLoginAt: customerTimeString(row.LastLoginAt),
			CreatedAt:   row.CreatedAt.In(time.Local).Format("2006-01-02 15:04:05"),
			UpdatedAt:   row.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05"),
		})
	}
	return items
}
