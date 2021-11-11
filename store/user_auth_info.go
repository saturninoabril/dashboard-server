package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"github.com/saturninoabril/dashboard-server/model"
)

type SqlUserAuthInfoStore struct {
	*SqlStore
}

func newSqlUserAuthInfoStore(sqlStore *SqlStore) UserAuthInfoStore {
	s := &SqlUserAuthInfoStore{
		SqlStore: sqlStore,
	}

	return s
}

func (s *SqlStore) UserAuthInfo() UserAuthInfoStore {
	return s.stores.user_auth_info
}

var userAuthInfoSelect sq.SelectBuilder

func init() {
	userAuthInfoSelect = sq.
		Select(
			"user_id",
			"oauth_provider",
			"token",
			"username",
			"email",
			"name",
			"avatar_url",
			"create_at",
			"update_at",
			"delete_at",
		)
}

func (s *SqlUserAuthInfoStore) getUserAuthInfoTable() string {
	return s.tablePrefix + "user_auth_info"
}

// CreateUserAuthInfo inserts a new user.
func (s *SqlUserAuthInfoStore) CreateUserAuthInfo(userAuthInfo *model.UserAuthInfo) (*model.UserAuthInfo, error) {
	userAuthInfo.PreSave()

	userAuthInfoTable := s.getUserAuthInfoTable()
	_, err := s.execBuilder(
		s.db,
		sq.Insert(userAuthInfoTable).
			SetMap(map[string]interface{}{
				"user_id":        userAuthInfo.UserID,
				"oauth_provider": userAuthInfo.OAuthProvider,
				"token":          userAuthInfo.Token,
				"username":       userAuthInfo.Username,
				"email":          userAuthInfo.Email,
				"name":           userAuthInfo.Name,
				"avatar_url":     userAuthInfo.AvatarURL,
				"create_at":      userAuthInfo.CreateAt,
				"update_at":      userAuthInfo.UpdateAt,
				"delete_at":      userAuthInfo.DeleteAt,
			}),
	)
	if err != nil {
		if isUniqueConstraintError(
			err,
			[]string{
				"UserID",
				userAuthInfoTable + "_user_id_key",
				"OAuthProvider",
				userAuthInfoTable + "_oauth_provider_key",
			}) {
			return nil, errors.New("user_id with given oauth_provider exists")
		}
		return nil, errors.Wrap(err, "failed to create user")
	}

	return userAuthInfo, nil
}

// GetUser fetches the given user by id.
func (s *SqlUserAuthInfoStore) GetUserAuthInfo(userID, oauthProvider string) (*model.UserAuthInfo, error) {
	var userAuthInfo model.UserAuthInfo
	err := s.getBuilder(
		s.db,
		&userAuthInfo,
		userAuthInfoSelect.
			From(s.getUserAuthInfoTable()).
			Where("user_id = $1 AND oauth_provider = $2", userID, oauthProvider),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get user by id")
	}

	return &userAuthInfo, nil
}

// UpdateUserAuthInfoToken updates the given user.
func (s *SqlUserAuthInfoStore) UpdateUserAuthInfoToken(userAuthInfo *model.UserAuthInfo) error {
	_, err := s.execBuilder(
		s.db,
		sq.Update(s.getUserAuthInfoTable()).
			Where("user_id = $1 AND oauth_provider = $2", userAuthInfo.UserID, userAuthInfo.OAuthProvider).
			SetMap(map[string]interface{}{
				"token":     userAuthInfo.Token,
				"update_at": model.GetMillis(),
			}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to update user token")
	}

	return nil
}
