package einterfaces

import (
	"im/model"
	"io"
)

type OauthProvider interface {
	GetUserFromJson(data io.Reader) *model.User
}

var oauthProviders = make(map[string]OauthProvider)

func RegisterOauthProvider(name string, newProvider OauthProvider) {
	oauthProviders[name] = newProvider
}

func GetOauthProvider(name string) OauthProvider {
	provider, ok := oauthProviders[name]
	if ok {
		return provider
	}
	return nil
}
