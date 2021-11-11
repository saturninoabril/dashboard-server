package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
)

type SqlRoleStore struct {
	*SqlStore
}

func newSqlRoleStore(sqlStore *SqlStore) RoleStore {
	s := &SqlRoleStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) Role() RoleStore {
	return s.stores.role
}

var (
	roleSelect     sq.SelectBuilder
	userRoleSelect sq.SelectBuilder
)

func init() {
	roleSelect = sq.
		Select(
			"id",
			"name",
			"create_at",
			"update_at",
		)
	userRoleSelect = sq.
		Select(
			"user_id",
			"role_id",
			"create_at",
			"update_at",
		)
}

func (s *SqlRoleStore) getRoleTable() string {
	return s.tablePrefix + "role"
}

func (s *SqlRoleStore) getUserRoleTable() string {
	return s.tablePrefix + "user_role"
}

// CreateRole inserts a new role.
func (s *SqlRoleStore) CreateRole(role *model.Role) (*model.Role, error) {
	role.CreatePreSave()

	_, err := s.execBuilder(
		s.db,
		sq.Insert(s.getRoleTable()).
			SetMap(map[string]interface{}{
				"id":        role.ID,
				"name":      role.Name,
				"create_at": role.CreateAt,
				"update_at": role.UpdateAt,
			}),
	)
	if err != nil {
		if isUniqueConstraintError(err, []string{"Name", "name"}) {
			return nil, errors.New("name exists")
		}
		return nil, errors.Wrap(err, "failed to create role")
	}

	return role, nil
}

// GetRoleByName returns the role entity for the provided name.
func (s *SqlRoleStore) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := s.getBuilder(
		s.db,
		&role,
		roleSelect.From(s.getRoleTable()).
			Where("name = $1", name),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to get role by name %s", name)
	}

	return &role, nil
}

func (s *SqlRoleStore) UserHasRole(userID, roleID string) (bool, error) {
	var userRole model.UserRole
	err := s.getBuilder(
		s.db, &userRole,
		userRoleSelect.From(s.getUserRoleTable()).
			Where("user_id = $1 AND role_id = $2", userID, roleID),
	)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "failed to user role for user %s and role %s", userID, roleID)
	}

	return true, nil
}

func (s *SqlRoleStore) UserHasRoleByName(userID, roleName string) (bool, error) {
	var userRole model.UserRole
	roleSelectByName := sq.Select("ur.user_id", "ur.role_id").
		From(s.getUserRoleTable() + " ur").
		Join(s.getRoleTable() + " r ON ur.role_id = r.id")
	err := s.getBuilder(s.db, &userRole, roleSelectByName.
		Where("ur.user_id = $1 AND r.name = $2", userID, roleName))
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(err, "failed to get role by name %s", roleName)
	}

	return true, nil
}

func (s *SqlRoleStore) AddUserRole(userID, roleID string) error {
	_, err := s.execBuilder(s.db, sq.
		Insert(s.getUserRoleTable()).
		SetMap(map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to add role %s to user %s", roleID, userID)
	}

	return nil
}

func (s *SqlRoleStore) DeleteUserRole(userID, roleID string) error {
	query := sq.Delete(s.getUserRoleTable()).
		Where("user_id = $1 AND role_id = $2", userID, roleID)
	if _, err := s.execBuilder(s.db, query); err != nil {
		return errors.Wrapf(err, "failed to add role %s to user %s", roleID, userID)
	}

	return nil
}
