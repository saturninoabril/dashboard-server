package main

import (
	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/app"
	"github.com/saturninoabril/dashboard-server/model"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(addNewUserCmd)
	addNewUserCmd.Flags().String("email", "", "Email for the user.")
	addNewUserCmd.Flags().String("password", "", "Password for the user.")
	addNewUserCmd.Flags().String("role", "", "Role for the created user. Possible values user and admin")
	addNewUserCmd.Flags().Bool("email-verified", true, "Is the user email verified.")
	userCmd.AddCommand(userRoleCmd)

	// Roles
	userRoleCmd.AddCommand(addUserRoleCmd)
	addUserRoleCmd.Flags().String("email", "", "Email for the user.")
	addUserRoleCmd.Flags().String("role", "", "Role to be added. Possible values user and admin")
	userRoleCmd.AddCommand(removeUserRoleCmd)
	removeUserRoleCmd.Flags().String("email", "", "Email for the user.")
	removeUserRoleCmd.Flags().String("role", "", "Role to be added. Possible values user and admin")
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Perform operations for the users in Dashboard.",
}

var userRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Perform operations with user's roles.",
}

var addNewUserCmd = &cobra.Command{
	Use:     "new",
	Short:   "Generate a new user",
	Long:    "Generate a new user in Dashboard",
	Example: "user new --email test@test.com --password testpassword --role user",
	RunE: func(command *cobra.Command, args []string) error {
		email, _ := command.Flags().GetString("email")
		emailVerified, _ := command.Flags().GetBool("email-verified")
		role, _ := command.Flags().GetString("role")
		password, _ := command.Flags().GetString("password")

		store, err := cmdStore(command)
		if err != nil {
			return err
		}
		userService := app.NewUserService(logger, store)
		userData := &model.User{
			Email:         email,
			EmailVerified: emailVerified,
			Password:      password,
		}
		user, err := userService.Create(userData)
		if err != nil {
			return errors.Wrapf(err, "Error creating user for %s", email)
		}
		err = userService.VerifyEmail(userData.ID, user.Email)
		if err != nil {
			return errors.Wrapf(err, "Error creating user for %s", email)
		}
		if role == model.AdminRoleName {
			roleData, err := store.Role().GetRoleByName(role)
			if err != nil {
				return errors.Wrapf(err, "Error creating user for %s", email)
			}
			err = store.Role().AddUserRole(user.ID, roleData.ID)
			if err != nil {
				return errors.Wrapf(err, "Error creating user for %s", email)
			}
		}
		return nil
	},
}

var addUserRoleCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add role to a user",
	Long:    "Add role to a user",
	Example: "user role add --email test@test.com --role user",
	RunE: func(command *cobra.Command, args []string) error {
		email, _ := command.Flags().GetString("email")
		role, _ := command.Flags().GetString("role")

		store, err := cmdStore(command)
		if err != nil {
			return err
		}
		userService := app.NewUserService(logger, store)
		user, err := userService.GetByEmail(email)
		if err != nil {
			return errors.Wrapf(err, "Error add role %s to user %s", role, email)
		}
		if user == nil {
			return errors.Errorf("User %s doesn't exist", email)
		}
		roleData, err := store.Role().GetRoleByName(role)
		if err != nil {
			return errors.Wrapf(err, "Error add role %s to user %s", role, email)
		}
		err = store.Role().AddUserRole(user.ID, roleData.ID)
		if err != nil {
			return errors.Wrapf(err, "Error add role %s to user %s", role, email)
		}
		logger.WithFields(logrus.Fields{
			"email": email,
			"role":  role,
		}).Info("Role added from user successfully")

		return nil
	},
}

var removeUserRoleCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove role from user",
	Long:    "Remove role from user",
	Example: "user role remove --email test@test.com --role user",
	RunE: func(command *cobra.Command, args []string) error {
		email, _ := command.Flags().GetString("email")
		role, _ := command.Flags().GetString("role")

		store, err := cmdStore(command)
		if err != nil {
			return err
		}
		userService := app.NewUserService(logger, store)
		user, err := userService.GetByEmail(email)
		if err != nil {
			return errors.Wrapf(err, "Error removing role %s to user %s", role, email)
		}
		if user == nil {
			return errors.Errorf("User %s doesn't exist", email)
		}
		roleData, err := store.Role().GetRoleByName(role)
		if err != nil {
			return errors.Wrapf(err, "Error removing role %s to user %s", role, email)
		}
		err = store.Role().DeleteUserRole(user.ID, roleData.ID)
		if err != nil {
			return errors.Wrapf(err, "Error removing role %s to user %s", role, email)
		}
		logger.WithFields(logrus.Fields{
			"email": email,
			"role":  role,
		}).Info("Role removed from user successfully")
		return nil
	},
}
