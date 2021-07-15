package store

import "github.com/saturninoabril/dashboard-server/model"

// Initialize imports initial values
func (s *SqlStore) Initialize() error {
	roleNames := []string{model.AdminRoleName, model.UserRoleName}
	for _, name := range roleNames {
		if err := s.createRoleIfNotExists(name); err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlStore) createRoleIfNotExists(name string) error {
	role, err := s.GetRoleByName(name)
	if err != nil {
		return err
	}
	if role == nil {
		roleData := &model.Role{
			ID:   model.NewID(),
			Name: name,
		}
		role, err = s.CreateRole(roleData)
		if err != nil {
			return err
		}
	}

	return nil
}
