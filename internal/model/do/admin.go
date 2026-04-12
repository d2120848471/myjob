package do

import "time"

type AdminUser struct {
	ID            any
	Username      any
	PasswordHash  any
	RealName      any
	Phone         any
	GroupID       any
	Status        any
	BalanceNotify any
	IsBusiness    any
	IsDeleted     any
	LastLoginIP   any
	LastLoginAt   *time.Time
	TokenVersion  any
	DeletedAt     *time.Time
	CreatedAt     any
	UpdatedAt     any
}

type AdminGroup struct {
	ID          any
	Name        any
	Description any
	Status      any
	CreatedAt   any
	UpdatedAt   any
}

type AdminMenu struct {
	ID        any
	ParentID  any
	Name      any
	Code      any
	MenuType  any
	MenuLevel any
	Status    any
	SuperOnly any
	Sort      any
	CreatedAt any
	UpdatedAt any
}

type AdminGroupMenu struct {
	ID        any
	GroupID   any
	MenuID    any
	CreatedAt any
}

type AdminSubject struct {
	ID        any
	Name      any
	HasTax    any
	CreatedAt any
	UpdatedAt any
}

type SystemConfig struct {
	ID          any
	ConfigKey   any
	ConfigValue any
	Description any
	CreatedAt   any
	UpdatedAt   any
}
