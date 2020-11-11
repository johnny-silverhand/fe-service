package app

import (
	"net/http"
	"strings"

	"im/model"
	"im/utils"
)

type TokenLocation int

const (
	TokenLocationNotFound = iota
	TokenLocationHeader
	TokenLocationCookie
	TokenLocationQueryString
)

func (tl TokenLocation) String() string {
	switch tl {
	case TokenLocationNotFound:
		return "Not Found"
	case TokenLocationHeader:
		return "Header"
	case TokenLocationCookie:
		return "Cookie"
	case TokenLocationQueryString:
		return "QueryString"
	default:
		return "Unknown"
	}
}

func (a *App) IsPasswordValid(password string) *model.AppError {
	return utils.IsPasswordValidWithSettings(password, &a.Config().PasswordSettings)
}

func (a *App) CheckPasswordAndAllCriteria(user *model.User, password string, mfaToken string) *model.AppError {
	if err := a.CheckUserPreflightAuthenticationCriteria(user, mfaToken); err != nil {
		return err
	}

	if err := a.checkUserPassword(user, password); err != nil {
		if result := <-a.Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, user.FailedAttempts+1); result.Err != nil {
			return result.Err
		}
		return err
	}

	if result := <-a.Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, 0); result.Err != nil {
		return result.Err
	}

	if err := a.CheckUserPostflightAuthenticationCriteria(user); err != nil {
		return err
	}

	return nil
}

// This to be used for places we check the users password when they are already logged in
func (a *App) DoubleCheckPassword(user *model.User, password string) *model.AppError {
	if err := checkUserLoginAttempts(user, *a.Config().ServiceSettings.MaximumLoginAttempts); err != nil {
		return err
	}

	if err := a.checkUserPassword(user, password); err != nil {
		if result := <-a.Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, user.FailedAttempts+1); result.Err != nil {
			return result.Err
		}
		return err
	}

	if result := <-a.Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, 0); result.Err != nil {
		return result.Err
	}

	return nil
}

func (a *App) checkUserPassword(user *model.User, password string) *model.AppError {
	if !model.ComparePassword(user.Password, password) {
		return model.NewAppError("checkUserPassword", "api.user.check_user_password.invalid.app_error", nil, "user_id="+user.Id, http.StatusUnauthorized)
	}

	return nil
}

func (a *App) CheckUserAllAuthenticationCriteria(user *model.User, mfaToken string) *model.AppError {
	if err := a.CheckUserPreflightAuthenticationCriteria(user, mfaToken); err != nil {
		return err
	}

	if err := a.CheckUserPostflightAuthenticationCriteria(user); err != nil {
		return err
	}

	return nil
}

func (a *App) CheckUserPreflightAuthenticationCriteria(user *model.User, mfaToken string) *model.AppError {
	if err := checkUserNotDisabled(user); err != nil {
		return err
	}

	if err := checkUserLoginAttempts(user, *a.Config().ServiceSettings.MaximumLoginAttempts); err != nil {
		return err
	}

	return nil
}

func (a *App) CheckUserPostflightAuthenticationCriteria(user *model.User) *model.AppError {
	if !user.EmailVerified && *a.Config().EmailSettings.RequireEmailVerification {
		return model.NewAppError("Login", "api.user.login.not_verified.app_error", nil, "user_id="+user.Id, http.StatusUnauthorized)
	}

	return nil
}

func checkUserLoginAttempts(user *model.User, max int) *model.AppError {
	if user.FailedAttempts >= max {
		return model.NewAppError("checkUserLoginAttempts", "api.user.check_user_login_attempts.too_many.app_error", nil, "user_id="+user.Id, http.StatusUnauthorized)
	}

	return nil
}

func checkUserNotDisabled(user *model.User) *model.AppError {
	if user.DeleteAt > 0 {
		return model.NewAppError("Login", "api.user.login.inactive.app_error", nil, "user_id="+user.Id, http.StatusUnauthorized)
	}
	return nil
}

func checkUserNotBot(user *model.User) *model.AppError {
	if user.IsBot {
		return model.NewAppError("Login", "api.user.login.bot_login_forbidden.app_error", nil, "user_id="+user.Id, http.StatusUnauthorized)
	}
	return nil
}

func (a *App) authenticateUser(user *model.User, password, mfaToken string) (*model.User, *model.AppError) {

	if user.AuthService != "" {
		authService := user.AuthService

		err := model.NewAppError("login", "api.user.login.use_auth_service.app_error", map[string]interface{}{"AuthService": authService}, "", http.StatusBadRequest)
		return user, err
	}

	if err := a.CheckPasswordAndAllCriteria(user, password, mfaToken); err != nil {
		err.StatusCode = http.StatusUnauthorized
		return user, err
	}

	return user, nil
}

func ParseAuthTokenFromRequest(r *http.Request) (string, TokenLocation) {
	authHeader := r.Header.Get(model.HEADER_AUTH)

	// Attempt to parse the token from the cookie
	if cookie, err := r.Cookie(model.SESSION_COOKIE_TOKEN); err == nil && len(cookie.Value) > 0 {
		return cookie.Value, TokenLocationCookie
	}

	// Parse the token from the header
	if len(authHeader) > 6 && strings.ToUpper(authHeader[0:6]) == model.HEADER_BEARER {
		// Default session token
		return authHeader[7:], TokenLocationHeader
	}

	if len(authHeader) > 5 && strings.ToLower(authHeader[0:5]) == model.HEADER_TOKEN {
		// OAuth token
		return authHeader[6:], TokenLocationHeader
	}

	// Attempt to parse token out of the query string
	if token := r.URL.Query().Get("access_token"); token != "" {
		return token, TokenLocationQueryString
	}

	return "", TokenLocationNotFound
}
