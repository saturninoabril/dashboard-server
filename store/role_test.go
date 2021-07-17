package store

import (
	"testing"

	"github.com/saturninoabril/dashboard-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoles(t *testing.T) {
	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("get unknown role", func(t *testing.T) {
		role, err := th.SqlStore.Role().GetRoleByName("")
		assert.NoError(t, err)
		assert.Nil(t, role)
	})

	t.Run("should get a default admin role", func(t *testing.T) {
		role, err := th.SqlStore.Role().GetRoleByName(model.AdminRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, model.AdminRoleName, role.Name)
	})

	t.Run("should get a default user role", func(t *testing.T) {
		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, model.UserRoleName, role.Name)
	})

	t.Run("should get empty if not exist", func(t *testing.T) {
		role, err := th.SqlStore.Role().GetRoleByName("fakerole")
		require.NoError(t, err)
		require.Nil(t, role)
	})

	t.Run("should create a new role", func(t *testing.T) {
		newRole := &model.Role{
			ID:   model.NewID(),
			Name: "new_role" + model.NewID(),
		}
		role, err := th.SqlStore.Role().CreateRole(newRole)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, newRole.Name, role.Name)

		role, err = th.SqlStore.Role().GetRoleByName(newRole.Name)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, newRole.Name, role.Name)
	})
}

func TestUserRoles(t *testing.T) {
	th := SetupStoreTestHelper(t)
	defer th.TearDown(t)

	t.Run("should return true or false if user has role or none", func(t *testing.T) {
		user := createTestUser(t, th.SqlStore)

		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)

		err = th.SqlStore.Role().AddUserRole(user.ID, role.ID)
		require.NoError(t, err)

		hasRole, err := th.SqlStore.Role().UserHasRole(user.ID, role.ID)
		require.NoError(t, err)
		require.True(t, hasRole)

		hasRole, err = th.SqlStore.Role().UserHasRole(user.ID, model.NewID())
		require.NoError(t, err)
		require.False(t, hasRole)
	})

	t.Run("should return correct role of the user", func(t *testing.T) {
		user := createTestUser(t, th.SqlStore)

		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)

		err = th.SqlStore.Role().AddUserRole(user.ID, role.ID)
		require.NoError(t, err)

		hasRole, err := th.SqlStore.Role().UserHasRoleByName(user.ID, model.UserRoleName)
		require.NoError(t, err)
		require.True(t, hasRole)

		hasRole, err = th.SqlStore.Role().UserHasRoleByName(user.ID, model.AdminRoleName)
		require.NoError(t, err)
		require.False(t, hasRole)
	})

	t.Run("should add a role to a user", func(t *testing.T) {
		user := createTestUser(t, th.SqlStore)

		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)

		err = th.SqlStore.Role().AddUserRole(user.ID, role.ID)
		require.NoError(t, err)
		hasRole, err := th.SqlStore.Role().UserHasRoleByName(user.ID, model.UserRoleName)
		require.NoError(t, err)
		require.True(t, hasRole)

		// adding non-existing role should not fail but should not change role of a user
		err = th.SqlStore.Role().AddUserRole(user.ID, model.NewID())
		require.NoError(t, err)
		hasRole, err = th.SqlStore.Role().UserHasRoleByName(user.ID, model.UserRoleName)
		require.NoError(t, err)
		require.True(t, hasRole)
	})

	t.Run("should remove a role from a user", func(t *testing.T) {
		user := createTestUser(t, th.SqlStore)

		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)

		err = th.SqlStore.Role().AddUserRole(user.ID, role.ID)
		require.NoError(t, err)

		err = th.SqlStore.Role().DeleteUserRole(user.ID, role.ID)
		require.NoError(t, err)

		hasRole, err := th.SqlStore.Role().UserHasRole(user.ID, role.ID)
		require.NoError(t, err)
		require.False(t, hasRole)
	})

	t.Run("should not remove other roles of a user role on removal of a role", func(t *testing.T) {
		user := createTestUser(t, th.SqlStore)

		role, err := th.SqlStore.Role().GetRoleByName(model.UserRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)
		adminRole, err := th.SqlStore.Role().GetRoleByName(model.AdminRoleName)
		require.NoError(t, err)
		require.NotNil(t, role)

		err = th.SqlStore.Role().AddUserRole(user.ID, role.ID)
		require.NoError(t, err)
		err = th.SqlStore.Role().AddUserRole(user.ID, adminRole.ID)
		require.NoError(t, err)

		err = th.SqlStore.Role().DeleteUserRole(user.ID, role.ID)
		require.NoError(t, err)

		hasRole, err := th.SqlStore.Role().UserHasRole(user.ID, adminRole.ID)
		require.NoError(t, err)
		require.True(t, hasRole)
	})
}
