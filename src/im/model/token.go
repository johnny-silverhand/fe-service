package model

import (
	"encoding/json"
	"net/http"
)

const (
	TOKEN_SIZE            = 64
	MAX_TOKEN_EXIPRY_TIME = 1000 * 60 * 60 * 48 // 48 hour
	TOKEN_TYPE_OAUTH      = "oauth"
	TOKEN_TYPE_DEF        = "app"
	TOKEN_TYPE_INVITE     = "invite"
)

type Token struct {
	Token    string `json:"token"`
	CreateAt int64  `json:"-"`
	Type     string `json:"-"`
	Extra    string `json:"extra"`
	UserId   string `json:"-"`
}

func NewToken(tokentype, extra string) *Token {
	return &Token{
		Token:    NewRandomString(TOKEN_SIZE),
		CreateAt: GetMillis(),
		Type:     tokentype,
		Extra:    extra,
	}
}

func NewInviteToken(userId string, extra string) *Token {
	return &Token{
		Token:    NewRandomString(TOKEN_SIZE),
		CreateAt: GetMillis(),
		Type:     TOKEN_TYPE_INVITE,
		Extra:    extra,
		UserId:   userId,
	}
}

func NewStageToken(userId string, extra string) *Token {
	return &Token{
		Token:    NewRandomString(TOKEN_SIZE),
		CreateAt: GetMillis(),
		Type:     TOKEN_TYPE_DEF,
		Extra:    extra,
		UserId:   userId,
	}
}

func (t *Token) IsValid() *AppError {
	if len(t.Token) != TOKEN_SIZE {
		return NewAppError("Token.IsValid", "model.token.is_valid.size", nil, "", http.StatusInternalServerError)
	}

	if t.CreateAt == 0 {
		return NewAppError("Token.IsValid", "model.token.is_valid.expiry", nil, "", http.StatusInternalServerError)
	}

	return nil
}

func (t *Token) ToJson() string {
	b, _ := json.Marshal(t)
	return string(b)
}
