package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
)

var (
	roleSelect     sq.SelectBuilder
	userRoleSelect sq.SelectBuilder
)

func init() {
	roleSelect = sq.
		Select("ID", "Name", "CreateAt", "UpdateAt").From("Roles")
	userRoleSelect = sq.
		Select("UserID", "RoleID", "CreateAt", "UpdateAt").From("UserRoles")
}

// GetRoleByName returns the role entity for the provided name.
func (store *DashboardStore) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := store.getBuilder(store.db, &role, roleSelect.Where("Name = $1", name))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get role by name %s", name)
	}

	return &role, nil
}

func (store *DashboardStore) UserHasRole(userID, roleID string) (bool, error) {
	var userRole model.UserRole
	err := store.getBuilder(store.db, &userRole, userRoleSelect.Where("UserID = $1 AND RoleID = $2", userID, roleID))
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "failed to user role for user %s and role %s", userID, roleID)
	}

	return true, nil
}

func (store *DashboardStore) UserHasRoleByName(userID, roleName string) (bool, error) {
	var userRole model.UserRole
	roleSelectByName := sq.Select("u.UserID", "u.RoleID").From("UserRoles u").Join("Roles r ON u.RoleID = r.ID")
	err := store.getBuilder(store.db, &userRole, roleSelectByName.Where("u.UserID = $1 AND r.Name = $2", userID, roleName))
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "failed to get role by name %s", roleName)
	}

	return true, nil
}

func (store *DashboardStore) AddUserRole(userID, roleID string) error {
	_, err := store.execBuilder(store.db, sq.
		Insert("UserRoles").
		SetMap(map[string]interface{}{
			"UserID": userID,
			"RoleID": roleID,
		}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to add role %s to user %s", roleID, userID)
	}

	return nil
}

func (store *DashboardStore) DeleteUserRole(userID, roleID string) error {
	query := sq.Delete("UserRoles").Where("UserID = $1 AND RoleID = $2", userID, roleID)
	if _, err := store.execBuilder(store.db, query); err != nil {
		return errors.Wrapf(err, "failed to add role %s to user %s", roleID, userID)
	}

	return nil
}
