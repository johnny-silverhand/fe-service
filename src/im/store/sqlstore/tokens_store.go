package sqlstore

import (
	"database/sql"
	"net/http"

	"im/mlog"
	"im/model"
	"im/store"
)

type SqlTokenStore struct {
	SqlStore
}

func NewSqlTokenStore(sqlStore SqlStore) store.TokenStore {
	s := &SqlTokenStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Token{}, "Tokens").SetKeys(false, "Token")
		table.ColMap("Token").SetMaxSize(64)
		table.ColMap("Type").SetMaxSize(64)
		table.ColMap("Extra").SetMaxSize(128)
	}

	return s
}

func (s SqlTokenStore) CreateIndexesIfNotExists() {
}

func (s SqlTokenStore) Save(token *model.Token) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if result.Err = token.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(token); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.Save", "store.sql_recover.save.app_error", nil, "", http.StatusInternalServerError)
		}
	})
}

func (s SqlTokenStore) Delete(token string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec("DELETE FROM Tokens WHERE Token = :Token", map[string]interface{}{"Token": token}); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.Delete", "store.sql_recover.delete.app_error", nil, "", http.StatusInternalServerError)
		}
	})
}

func (s SqlTokenStore) GetByToken(tokenString string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		token := model.Token{}

		if err := s.GetReplica().SelectOne(&token, "SELECT * FROM Tokens WHERE Token = :Token", map[string]interface{}{"Token": tokenString}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlTokenStore.GetByToken", "store.sql_recover.get_by_code.app_error", nil, err.Error(), http.StatusBadRequest)
			} else {
				result.Err = model.NewAppError("SqlTokenStore.GetByToken", "store.sql_recover.get_by_code.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		}

		result.Data = &token
	})
}

func (s SqlTokenStore) GetByUserInviteToken(userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := s.getQueryBuilder().
			Select("*").
			From("Tokens t").
			Where("t.UserId = ? AND t.Type = ?", userId, model.TOKEN_TYPE_INVITE).
			OrderBy("t.CreateAt DESC")

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlTokenStore.GetByUserInviteToken", "store.sql_token.get_user_invite_token.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var tokens []*model.Token
		if _, err := s.GetMaster().Select(&tokens, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.GetByUserInviteToken", "store.sql_token.get_user_invite_token.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(tokens) > 1 || len(tokens) == 0 {
			result.Err = model.NewAppError("SqlTokenStore.GetByUserInviteToken", store.INVITE_TOKEN_NOT_FOUND, nil, "userId="+userId, http.StatusInternalServerError)
			return
		}

		result.Data = tokens[0]

		/*token := model.Token{}

		if err := s.GetReplica().SelectOne(&token, "SELECT * FROM Tokens WHERE UserId = :UserId AND Type = :Type ORDER BY CreateAt DESC", map[string]interface{}{"UserId": userId, "Type": model.TOKEN_TYPE_INVITE}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlTokenStore.GetByUserInviteToken", "store.sql_token.get_user_invite_token.not_found", nil, err.Error(), http.StatusBadRequest)
			} else {
				result.Err = model.NewAppError("SqlTokenStore.GetByUserInviteToken", "store.sql_token.get_user_invite_token.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		}

		result.Data = &token*/
	})
}

func (s SqlTokenStore) GetByApplicationInviteCode(appId string, code string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		token := model.Token{}

		if err := s.GetReplica().SelectOne(&token, "SELECT T.* FROM Tokens T JOIN Users U ON U.Id = T.UserId WHERE T.Type = :Type AND T.Extra = :Extra AND U.AppId = :AppId",
			map[string]interface{}{"Type": model.TOKEN_TYPE_INVITE, "AppId": appId, "Extra": code}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlTokenStore.GetByToken", "store.sql_recover.get_by_code.app_error", nil, err.Error(), http.StatusBadRequest)
			} else {
				result.Err = model.NewAppError("SqlTokenStore.GetByToken", "store.sql_recover.get_by_code.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		}

		result.Data = &token
	})
}

func (s SqlTokenStore) Cleanup() {
	mlog.Debug("Cleaning up token store.")
	deltime := model.GetMillis() - model.MAX_TOKEN_EXIPRY_TIME
	if _, err := s.GetMaster().Exec("DELETE FROM Tokens WHERE CreateAt < :DelTime", map[string]interface{}{"DelTime": deltime}); err != nil {
		mlog.Error("Unable to cleanup token store.")
	}
}

func (s SqlTokenStore) RemoveAllTokensByType(tokenType string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec("DELETE FROM Tokens WHERE Type = :TokenType", map[string]interface{}{"TokenType": tokenType}); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.RemoveAllTokensByType", "store.sql_recover.remove_all_tokens_by_type.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (s SqlTokenStore) RemoveUserTokensByType(tokenType string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec("DELETE FROM Tokens WHERE Type = :TokenType AND UserId = :UserId ", map[string]interface{}{"TokenType": tokenType, "UserId": userId}); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.RemoveUserTokensByType", "store.sql_recover.remove_user_tokens_by_type.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (us SqlTokenStore) UpdateExtra(token, extra string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//updateAt := model.GetMillis()

		if _, err := us.GetMaster().Exec("UPDATE Tokens SET Extra = :Extra , CreateAt = :CreateAt WHERE Token = :Token", map[string]interface{}{"Token": token, "Extra": extra, "CreateAt": model.GetMillis()}); err != nil {
			result.Err = model.NewAppError("SqlTokenStore.UpdateExtra", "store.sql_user.update_extra.app_error", nil, "token="+token+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = token
		}
	})
}
