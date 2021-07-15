package model

const (
	UserRoleName  = "user"
	AdminRoleName = "admin"
)

type Role struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
}

type UserRole struct {
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
}

// CreatePreSave will set the correct values for a new role that is about to be
// saved.
func (r *Role) CreatePreSave() {
	if r.ID == "" {
		r.ID = NewID()
	}

	now := GetMillis()
	r.CreateAt = now
	r.UpdateAt = now
}
