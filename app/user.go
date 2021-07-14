package app

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/store"
	"github.com/saturninoabril/dashboard-server/utils"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	Create(user *model.User) (*model.User, error)
	Get(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) (*model.User, error)
	AuthenticateUserForLogin(email, password string) (*model.User, error)
	Login(w http.ResponseWriter, r *http.Request, user *model.User) error
	Logout(w http.ResponseWriter, r *http.Request, sessionID string)
	GetSession(tokenOrID string) (*model.Session, error)
	VerifyEmail(id, email string) error
	UnverifyEmail(id, email string) error
	HasAdminPermission(id string) (bool, error)
}

type userService struct {
	logger logrus.FieldLogger
	store  store.Store
}

var _ UserService = &userService{}

func NewUserService(logger logrus.FieldLogger, store store.Store) UserService {
	logFields := logrus.Fields{"package": "app", "type": "user"}
	return &userService{
		logger: logger.WithFields(logFields),
		store:  store,
	}
}

func (u *userService) Logger(user *model.User) logrus.FieldLogger {
	return u.logger.WithField("id", user.ID).WithField("email", user.Email)
}

func (u *userService) Create(user *model.User) (*model.User, error) {
	user.Email = strings.ToLower(user.Email)
	user, err := u.store.User().CreateUser(user)
	if err != nil {
		return nil, err
	}

	role, err := u.store.GetRoleByName(model.UserRoleName)
	if err != nil || role == nil {
		return nil, err
	}

	if err = u.store.AddUserRole(user.ID, role.ID); err != nil {
		return nil, err
	}

	u.Logger(user).Info("create")

	return user, nil
}

func (u *userService) Get(id string) (*model.User, error) {
	user, err := u.store.User().GetUser(id)
	if err != nil || user == nil {
		return nil, err
	}
	isAdmin, err := u.store.UserHasRoleByName(user.ID, model.AdminRoleName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting user with id %s", user.ID)
	}
	user.IsAdmin = isAdmin
	return user, nil
}

func (u *userService) GetByEmail(email string) (*model.User, error) {
	return u.store.User().GetUserByEmail(email)
}

func (u *userService) AuthenticateUserForLogin(email, password string) (*model.User, error) {
	if len(password) == 0 {
		return nil, errors.New("blank password")
	}

	email = strings.ToLower(email)

	user, err := u.store.User().GetUserByEmail(email)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user for login")
	}
	if user == nil {
		return nil, errors.New("no user")
	}

	if !model.ComparePassword(user.Password, password) {
		return nil, errors.New("bad password")
	}

	isAdmin, err := u.store.UserHasRoleByName(user.ID, model.AdminRoleName)
	if err != nil {
		return nil, errors.New("error getting admin role for user")
	}
	user.IsAdmin = isAdmin

	return user, nil
}

func (u *userService) Update(user *model.User) (*model.User, error) {
	err := u.store.User().UpdateUser(user)
	if err != nil {
		return nil, err
	}

	u.Logger(user).Info("update")

	return u.store.User().GetUser(user.ID)
}

func (u *userService) Login(w http.ResponseWriter, r *http.Request, user *model.User) error {
	session := &model.Session{UserID: user.ID}

	session, err := u.store.CreateSession(session)
	if err != nil {
		return err
	}

	w.Header().Set(model.SessionHeader, session.Token)

	if r.Header.Get(model.HeaderRequestedWith) == model.HeaderRequestedWithXML {
		utils.AttachSessionCookies(w, r, session)
	}

	u.Logger(user).Info("login")

	return nil
}

func (u *userService) Logout(w http.ResponseWriter, r *http.Request, sessionID string) {
	utils.DeleteSessionCookies(w, r)
	_ = u.store.DeleteSession(sessionID)
}

func (u *userService) GetSession(tokenOrID string) (*model.Session, error) {
	return u.store.GetSession(tokenOrID)
}

func (u *userService) VerifyEmail(id, email string) error {
	return u.store.User().VerifyEmail(id, email)
}

func (u *userService) UnverifyEmail(id, email string) error {
	return u.store.User().UnverifyEmail(id, email)
}

func (u *userService) HasAdminPermission(id string) (bool, error) {
	hasRole, err := u.store.UserHasRoleByName(id, model.AdminRoleName)
	if err != nil {
		return false, errors.Wrapf(err, "error verifying admin permissions for user %s", id)
	}

	return hasRole, nil
}
