// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CONN_SECURITY_NONE     = ""
	CONN_SECURITY_PLAIN    = "PLAIN"
	CONN_SECURITY_TLS      = "TLS"
	CONN_SECURITY_STARTTLS = "STARTTLS"

	IMAGE_DRIVER_LOCAL = "local"
	IMAGE_DRIVER_S3    = "amazons3"

	DATABASE_DRIVER_SQLITE   = "sqlite3"
	DATABASE_DRIVER_MYSQL    = "mysql"
	DATABASE_DRIVER_POSTGRES = "postgres"

	MINIO_ACCESS_KEY = "minioaccesskey"
	MINIO_SECRET_KEY = "miniosecretkey"
	MINIO_BUCKET     = "mattermost-test"

	PASSWORD_MAXIMUM_LENGTH = 64
	PASSWORD_MINIMUM_LENGTH = 5

	SERVICE_GITLAB    = "gitlab"
	SERVICE_GOOGLE    = "google"
	SERVICE_OFFICE365 = "office365"

	GENERIC_NO_CHANNEL_NOTIFICATION = "generic_no_channel"
	GENERIC_NOTIFICATION            = "generic"
	FULL_NOTIFICATION               = "full"

	DIRECT_MESSAGE_ANY  = "any"
	DIRECT_MESSAGE_TEAM = "team"

	SHOW_USERNAME          = "username"
	SHOW_NICKNAME_FULLNAME = "nickname_full_name"
	SHOW_FULLNAME          = "full_name"

	PERMISSIONS_ALL           = "all"
	PERMISSIONS_CHANNEL_ADMIN = "channel_admin"
	PERMISSIONS_TEAM_ADMIN    = "team_admin"
	PERMISSIONS_SYSTEM_ADMIN  = "system_admin"

	FAKE_SETTING = "********************************"

	RESTRICT_EMOJI_CREATION_ALL          = "all"
	RESTRICT_EMOJI_CREATION_ADMIN        = "admin"
	RESTRICT_EMOJI_CREATION_SYSTEM_ADMIN = "system_admin"

	PERMISSIONS_DELETE_POST_ALL          = "all"
	PERMISSIONS_DELETE_POST_TEAM_ADMIN   = "team_admin"
	PERMISSIONS_DELETE_POST_SYSTEM_ADMIN = "system_admin"

	ALLOW_EDIT_POST_ALWAYS     = "always"
	ALLOW_EDIT_POST_NEVER      = "never"
	ALLOW_EDIT_POST_TIME_LIMIT = "time_limit"

	GROUP_UNREAD_CHANNELS_DISABLED    = "disabled"
	GROUP_UNREAD_CHANNELS_DEFAULT_ON  = "default_on"
	GROUP_UNREAD_CHANNELS_DEFAULT_OFF = "default_off"

	EMAIL_BATCHING_BUFFER_SIZE = 256
	EMAIL_BATCHING_INTERVAL    = 30

	EMAIL_NOTIFICATION_CONTENTS_FULL    = "full"
	EMAIL_NOTIFICATION_CONTENTS_GENERIC = "generic"

	SITENAME_MAX_LENGTH = 30

	SERVICE_SETTINGS_DEFAULT_SITE_URL           = "http://localhost:8065"
	SERVICE_SETTINGS_DEFAULT_TLS_CERT_FILE      = ""
	SERVICE_SETTINGS_DEFAULT_TLS_KEY_FILE       = ""
	SERVICE_SETTINGS_DEFAULT_READ_TIMEOUT       = 300
	SERVICE_SETTINGS_DEFAULT_WRITE_TIMEOUT      = 300
	SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS = 10
	SERVICE_SETTINGS_DEFAULT_ALLOW_CORS_FROM    = ""
	SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS = ":8065"
	SERVICE_SETTINGS_DEFAULT_GFYCAT_API_KEY     = "2_KtH_W5"
	SERVICE_SETTINGS_DEFAULT_GFYCAT_API_SECRET  = "3wLVZPiswc3DnaiaFoLkDvB4X0IV6CpMkj4tf2inJRsBY6-FnkT08zGmppWFgeof"

	TEAM_SETTINGS_DEFAULT_SITE_NAME                = "Mattermost"
	TEAM_SETTINGS_DEFAULT_MAX_USERS_PER_TEAM       = 50
	TEAM_SETTINGS_DEFAULT_CUSTOM_BRAND_TEXT        = ""
	TEAM_SETTINGS_DEFAULT_CUSTOM_DESCRIPTION_TEXT  = ""
	TEAM_SETTINGS_DEFAULT_USER_STATUS_AWAY_TIMEOUT = 300

	SQL_SETTINGS_DEFAULT_DATA_SOURCE = "mmuser:mostest@tcp(dockerhost:3306)/mattermost_test?charset=utf8mb4,utf8&readTimeout=30s&writeTimeout=30s"

	FILE_SETTINGS_DEFAULT_DIRECTORY = "./data/"

	EMAIL_SETTINGS_DEFAULT_FEEDBACK_ORGANIZATION = ""

	SUPPORT_SETTINGS_DEFAULT_TERMS_OF_SERVICE_LINK = "https://about.mattermost.com/default-terms/"
	SUPPORT_SETTINGS_DEFAULT_PRIVACY_POLICY_LINK   = "https://about.mattermost.com/default-privacy-policy/"
	SUPPORT_SETTINGS_DEFAULT_ABOUT_LINK            = "https://about.mattermost.com/default-about/"
	SUPPORT_SETTINGS_DEFAULT_HELP_LINK             = "https://about.mattermost.com/default-help/"
	SUPPORT_SETTINGS_DEFAULT_REPORT_A_PROBLEM_LINK = "https://about.mattermost.com/default-report-a-problem/"
	SUPPORT_SETTINGS_DEFAULT_SUPPORT_EMAIL         = "feedback@mattermost.com"
	SUPPORT_SETTINGS_DEFAULT_RE_ACCEPTANCE_PERIOD  = 365

	LDAP_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE         = ""
	LDAP_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE          = ""
	LDAP_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE              = ""
	LDAP_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_ID_ATTRIBUTE                 = ""
	LDAP_SETTINGS_DEFAULT_POSITION_ATTRIBUTE           = ""
	LDAP_SETTINGS_DEFAULT_LOGIN_FIELD_NAME             = ""
	LDAP_SETTINGS_DEFAULT_GROUP_DISPLAY_NAME_ATTRIBUTE = ""
	LDAP_SETTINGS_DEFAULT_GROUP_ID_ATTRIBUTE           = ""

	SAML_SETTINGS_DEFAULT_ID_ATTRIBUTE         = ""
	SAML_SETTINGS_DEFAULT_FIRST_NAME_ATTRIBUTE = ""
	SAML_SETTINGS_DEFAULT_LAST_NAME_ATTRIBUTE  = ""
	SAML_SETTINGS_DEFAULT_EMAIL_ATTRIBUTE      = ""
	SAML_SETTINGS_DEFAULT_USERNAME_ATTRIBUTE   = ""
	SAML_SETTINGS_DEFAULT_NICKNAME_ATTRIBUTE   = ""
	SAML_SETTINGS_DEFAULT_LOCALE_ATTRIBUTE     = ""
	SAML_SETTINGS_DEFAULT_POSITION_ATTRIBUTE   = ""

	NATIVEAPP_SETTINGS_DEFAULT_APP_DOWNLOAD_LINK         = "https://about.mattermost.com/downloads/"
	NATIVEAPP_SETTINGS_DEFAULT_ANDROID_APP_DOWNLOAD_LINK = "https://about.mattermost.com/mattermost-android-app/"
	NATIVEAPP_SETTINGS_DEFAULT_IOS_APP_DOWNLOAD_LINK     = "https://about.mattermost.com/mattermost-ios-app/"

	EXPERIMENTAL_SETTINGS_DEFAULT_LINK_METADATA_TIMEOUT_MILLISECONDS = 5000

	ANALYTICS_SETTINGS_DEFAULT_MAX_USERS_FOR_STATISTICS = 2500

	ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_COLOR      = "#f2a93b"
	ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_TEXT_COLOR = "#333333"

	TEAM_SETTINGS_DEFAULT_TEAM_TEXT = "default"

	ELASTICSEARCH_SETTINGS_DEFAULT_CONNECTION_URL                    = "http://dockerhost:9200"
	ELASTICSEARCH_SETTINGS_DEFAULT_USERNAME                          = "elastic"
	ELASTICSEARCH_SETTINGS_DEFAULT_PASSWORD                          = "changeme"
	ELASTICSEARCH_SETTINGS_DEFAULT_POST_INDEX_REPLICAS               = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_POST_INDEX_SHARDS                 = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_CHANNEL_INDEX_REPLICAS            = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_CHANNEL_INDEX_SHARDS              = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_USER_INDEX_REPLICAS               = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_USER_INDEX_SHARDS                 = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_AGGREGATE_POSTS_AFTER_DAYS        = 365
	ELASTICSEARCH_SETTINGS_DEFAULT_POSTS_AGGREGATOR_JOB_START_TIME   = "03:00"
	ELASTICSEARCH_SETTINGS_DEFAULT_INDEX_PREFIX                      = ""
	ELASTICSEARCH_SETTINGS_DEFAULT_LIVE_INDEXING_BATCH_SIZE          = 1
	ELASTICSEARCH_SETTINGS_DEFAULT_BULK_INDEXING_TIME_WINDOW_SECONDS = 3600
	ELASTICSEARCH_SETTINGS_DEFAULT_REQUEST_TIMEOUT_SECONDS           = 30

	DATA_RETENTION_SETTINGS_DEFAULT_MESSAGE_RETENTION_DAYS  = 365
	DATA_RETENTION_SETTINGS_DEFAULT_FILE_RETENTION_DAYS     = 365
	DATA_RETENTION_SETTINGS_DEFAULT_DELETION_JOB_START_TIME = "02:00"

	PLUGIN_SETTINGS_DEFAULT_DIRECTORY        = "./plugins"
	PLUGIN_SETTINGS_DEFAULT_CLIENT_DIRECTORY = "./client/plugins"

	COMPLIANCE_EXPORT_TYPE_CSV         = "csv"
	COMPLIANCE_EXPORT_TYPE_ACTIANCE    = "actiance"
	COMPLIANCE_EXPORT_TYPE_GLOBALRELAY = "globalrelay"
	GLOBALRELAY_CUSTOMER_TYPE_A9       = "A9"
	GLOBALRELAY_CUSTOMER_TYPE_A10      = "A10"

	CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH   = "primary"
	CLIENT_SIDE_CERT_CHECK_SECONDARY_AUTH = "secondary"

	IMAGE_PROXY_TYPE_LOCAL      = "local"
	IMAGE_PROXY_TYPE_ATMOS_CAMO = "atmos/camo"

	PAYMENT_PROXY_TYPE_ALFABANK = "alfa"
	PAYMENT_PROXY_TYPE_SBERBANK = "sber"
)

var ServerTLSSupportedCiphers = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                tls.TLS_RSA_WITH_RC4_128_SHA,
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
}

type ServiceSettings struct {
	SiteURL                                           *string  `restricted:"true"`
	WebsocketURL                                      *string  `restricted:"true"`
	LicenseFileLocation                               *string  `restricted:"true"`
	ListenAddress                                     *string  `restricted:"true"`
	ConnectionSecurity                                *string  `restricted:"true"`
	TLSCertFile                                       *string  `restricted:"true"`
	TLSKeyFile                                        *string  `restricted:"true"`
	TLSMinVer                                         *string  `restricted:"true"`
	TLSStrictTransport                                *bool    `restricted:"true"`
	TLSStrictTransportMaxAge                          *int64   `restricted:"true"`
	TLSOverwriteCiphers                               []string `restricted:"true"`
	UseLetsEncrypt                                    *bool    `restricted:"true"`
	LetsEncryptCertificateCacheFile                   *string  `restricted:"true"`
	Forward80To443                                    *bool    `restricted:"true"`
	ReadTimeout                                       *int     `restricted:"true"`
	WriteTimeout                                      *int     `restricted:"true"`
	MaximumLoginAttempts                              *int     `restricted:"true"`
	GoroutineHealthThreshold                          *int     `restricted:"true"`
	GoogleDeveloperKey                                *string  `restricted:"true"`
	EnableOAuthServiceProvider                        *bool
	EnableIncomingWebhooks                            *bool
	EnableOutgoingWebhooks                            *bool
	EnableCommands                                    *bool
	DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations *bool `json:"EnableOnlyAdminIntegrations"` // This field is deprecated and must not be used.
	EnablePostUsernameOverride                        *bool
	EnablePostIconOverride                            *bool
	EnableLinkPreviews                                *bool
	EnableTesting                                     *bool   `restricted:"true"`
	EnableDeveloper                                   *bool   `restricted:"true"`
	EnableSecurityFixAlert                            *bool   `restricted:"true"`
	EnableInsecureOutgoingConnections                 *bool   `restricted:"true"`
	AllowedUntrustedInternalConnections               *string `restricted:"true"`
	EnableMultifactorAuthentication                   *bool
	EnforceMultifactorAuthentication                  *bool
	EnableUserAccessTokens                            *bool
	AllowCorsFrom                                     *string `restricted:"true"`
	CorsExposedHeaders                                *string `restricted:"true"`
	CorsAllowCredentials                              *bool   `restricted:"true"`
	CorsDebug                                         *bool   `restricted:"true"`
	AllowCookiesForSubdomains                         *bool   `restricted:"true"`
	SessionLengthWebInDays                            *int    `restricted:"true"`
	SessionLengthMobileInDays                         *int    `restricted:"true"`
	SessionLengthSSOInDays                            *int    `restricted:"true"`
	SessionCacheInMinutes                             *int    `restricted:"true"`
	SessionIdleTimeoutInMinutes                       *int    `restricted:"true"`
	WebsocketSecurePort                               *int    `restricted:"true"`
	WebsocketPort                                     *int    `restricted:"true"`
	WebserverMode                                     *string `restricted:"true"`
	EnableCustomEmoji                                 *bool
	EnableEmojiPicker                                 *bool
	EnableGifPicker                                   *bool
	GfycatApiKey                                      *string
	GfycatApiSecret                                   *string
	DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation *string `json:"RestrictCustomEmojiCreation"` // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPostDelete          *string `json:"RestrictPostDelete"`          // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_AllowEditPost               *string `json:"AllowEditPost"`               // This field is deprecated and must not be used.
	PostEditTimeLimit                                 *int
	TimeBetweenUserTypingUpdatesMilliseconds          *int64 `restricted:"true"`
	EnablePostSearch                                  *bool  `restricted:"true"`
	MinimumHashtagLength                              *int   `restricted:"true"`
	EnableUserTypingMessages                          *bool  `restricted:"true"`
	EnableChannelViewedMessages                       *bool  `restricted:"true"`
	EnableUserStatuses                                *bool  `restricted:"true"`
	ExperimentalEnableAuthenticationTransfer          *bool  `restricted:"true"`
	ClusterLogTimeoutMilliseconds                     *int   `restricted:"true"`
	CloseUnusedDirectMessages                         *bool
	EnablePreviewFeatures                             *bool
	EnableTutorial                                    *bool
	ExperimentalEnableDefaultChannelLeaveJoinMessages *bool
	ExperimentalGroupUnreadChannels                   *string
	ExperimentalChannelOrganization                   *bool
	DEPRECATED_DO_NOT_USE_ImageProxyType              *string `json:"ImageProxyType" mapstructure:"ImageProxyType"`       // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_ImageProxyURL               *string `json:"ImageProxyURL" mapstructure:"ImageProxyURL"`         // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_ImageProxyOptions           *string `json:"ImageProxyOptions" mapstructure:"ImageProxyOptions"` // This field is deprecated and must not be used.
	EnableAPITeamDeletion                             *bool
	ExperimentalEnableHardenedMode                    *bool
	DisableLegacyMFA                                  *bool `restricted:"true"`
	ExperimentalStrictCSRFEnforcement                 *bool `restricted:"true"`
	EnableEmailInvitations                            *bool
	ExperimentalLdapGroupSync                         *bool
	DisableBotsWhenOwnerIsDeactivated                 *bool `restricted:"true"`
}

func (s *ServiceSettings) SetDefaults() {
	if s.EnableEmailInvitations == nil {
		// If the site URL is also not present then assume this is a clean install
		if s.SiteURL == nil {
			s.EnableEmailInvitations = NewBool(false)
		} else {
			s.EnableEmailInvitations = NewBool(true)
		}
	}

	if s.SiteURL == nil {
		s.SiteURL = NewString(SERVICE_SETTINGS_DEFAULT_SITE_URL)
	}

	if s.WebsocketURL == nil {
		s.WebsocketURL = NewString("")
	}

	if s.LicenseFileLocation == nil {
		s.LicenseFileLocation = NewString("")
	}

	if s.ListenAddress == nil {
		s.ListenAddress = NewString(SERVICE_SETTINGS_DEFAULT_LISTEN_AND_ADDRESS)
	}

	if s.EnableLinkPreviews == nil {
		s.EnableLinkPreviews = NewBool(false)
	}

	if s.EnableTesting == nil {
		s.EnableTesting = NewBool(false)
	}

	if s.EnableDeveloper == nil {
		s.EnableDeveloper = NewBool(false)
	}

	if s.EnableSecurityFixAlert == nil {
		s.EnableSecurityFixAlert = NewBool(true)
	}

	if s.EnableInsecureOutgoingConnections == nil {
		s.EnableInsecureOutgoingConnections = NewBool(false)
	}

	if s.AllowedUntrustedInternalConnections == nil {
		s.AllowedUntrustedInternalConnections = NewString("")
	}

	if s.EnableMultifactorAuthentication == nil {
		s.EnableMultifactorAuthentication = NewBool(false)
	}

	if s.EnforceMultifactorAuthentication == nil {
		s.EnforceMultifactorAuthentication = NewBool(false)
	}

	if s.EnableUserAccessTokens == nil {
		s.EnableUserAccessTokens = NewBool(false)
	}

	if s.GoroutineHealthThreshold == nil {
		s.GoroutineHealthThreshold = NewInt(-1)
	}

	if s.GoogleDeveloperKey == nil {
		s.GoogleDeveloperKey = NewString("")
	}

	if s.EnableOAuthServiceProvider == nil {
		s.EnableOAuthServiceProvider = NewBool(false)
	}

	if s.EnableIncomingWebhooks == nil {
		s.EnableIncomingWebhooks = NewBool(true)
	}

	if s.EnableIncomingWebhooks == nil {
		s.EnableIncomingWebhooks = NewBool(true)
	}

	if s.EnableOutgoingWebhooks == nil {
		s.EnableOutgoingWebhooks = NewBool(true)
	}

	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("")
	}

	if s.TLSKeyFile == nil {
		s.TLSKeyFile = NewString(SERVICE_SETTINGS_DEFAULT_TLS_KEY_FILE)
	}

	if s.TLSCertFile == nil {
		s.TLSCertFile = NewString(SERVICE_SETTINGS_DEFAULT_TLS_CERT_FILE)
	}

	if s.TLSMinVer == nil {
		s.TLSMinVer = NewString("1.2")
	}

	if s.TLSStrictTransport == nil {
		s.TLSStrictTransport = NewBool(false)
	}

	if s.TLSStrictTransportMaxAge == nil {
		s.TLSStrictTransportMaxAge = NewInt64(63072000)
	}

	if s.TLSOverwriteCiphers == nil {
		s.TLSOverwriteCiphers = []string{}
	}

	if s.UseLetsEncrypt == nil {
		s.UseLetsEncrypt = NewBool(false)
	}

	if s.LetsEncryptCertificateCacheFile == nil {
		s.LetsEncryptCertificateCacheFile = NewString("./config/letsencrypt.cache")
	}

	if s.ReadTimeout == nil {
		s.ReadTimeout = NewInt(SERVICE_SETTINGS_DEFAULT_READ_TIMEOUT)
	}

	if s.WriteTimeout == nil {
		s.WriteTimeout = NewInt(SERVICE_SETTINGS_DEFAULT_WRITE_TIMEOUT)
	}

	if s.MaximumLoginAttempts == nil {
		s.MaximumLoginAttempts = NewInt(SERVICE_SETTINGS_DEFAULT_MAX_LOGIN_ATTEMPTS)
	}

	if s.Forward80To443 == nil {
		s.Forward80To443 = NewBool(false)
	}

	if s.TimeBetweenUserTypingUpdatesMilliseconds == nil {
		s.TimeBetweenUserTypingUpdatesMilliseconds = NewInt64(5000)
	}

	if s.EnablePostSearch == nil {
		s.EnablePostSearch = NewBool(true)
	}

	if s.MinimumHashtagLength == nil {
		s.MinimumHashtagLength = NewInt(3)
	}

	if s.EnableUserTypingMessages == nil {
		s.EnableUserTypingMessages = NewBool(true)
	}

	if s.EnableChannelViewedMessages == nil {
		s.EnableChannelViewedMessages = NewBool(true)
	}

	if s.EnableUserStatuses == nil {
		s.EnableUserStatuses = NewBool(true)
	}

	if s.ClusterLogTimeoutMilliseconds == nil {
		s.ClusterLogTimeoutMilliseconds = NewInt(2000)
	}

	if s.CloseUnusedDirectMessages == nil {
		s.CloseUnusedDirectMessages = NewBool(false)
	}

	if s.EnableTutorial == nil {
		s.EnableTutorial = NewBool(true)
	}

	if s.SessionLengthWebInDays == nil {
		s.SessionLengthWebInDays = NewInt(180)
	}

	if s.SessionLengthMobileInDays == nil {
		s.SessionLengthMobileInDays = NewInt(180)
	}

	if s.SessionLengthSSOInDays == nil {
		s.SessionLengthSSOInDays = NewInt(30)
	}

	if s.SessionCacheInMinutes == nil {
		s.SessionCacheInMinutes = NewInt(10)
	}

	if s.SessionIdleTimeoutInMinutes == nil {
		s.SessionIdleTimeoutInMinutes = NewInt(43200)
	}

	if s.EnableCommands == nil {
		s.EnableCommands = NewBool(true)
	}

	if s.DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations == nil {
		s.DEPRECATED_DO_NOT_USE_EnableOnlyAdminIntegrations = NewBool(true)
	}

	if s.EnablePostUsernameOverride == nil {
		s.EnablePostUsernameOverride = NewBool(false)
	}

	if s.EnablePostIconOverride == nil {
		s.EnablePostIconOverride = NewBool(false)
	}

	if s.WebsocketPort == nil {
		s.WebsocketPort = NewInt(80)
	}

	if s.WebsocketSecurePort == nil {
		s.WebsocketSecurePort = NewInt(443)
	}

	if s.AllowCorsFrom == nil {
		s.AllowCorsFrom = NewString(SERVICE_SETTINGS_DEFAULT_ALLOW_CORS_FROM)
	}

	if s.CorsExposedHeaders == nil {
		s.CorsExposedHeaders = NewString("")
	}

	if s.CorsAllowCredentials == nil {
		s.CorsAllowCredentials = NewBool(false)
	}

	if s.CorsDebug == nil {
		s.CorsDebug = NewBool(false)
	}

	if s.AllowCookiesForSubdomains == nil {
		s.AllowCookiesForSubdomains = NewBool(false)
	}

	if s.WebserverMode == nil {
		s.WebserverMode = NewString("gzip")
	} else if *s.WebserverMode == "regular" {
		*s.WebserverMode = "gzip"
	}

	if s.EnableCustomEmoji == nil {
		s.EnableCustomEmoji = NewBool(false)
	}

	if s.EnableEmojiPicker == nil {
		s.EnableEmojiPicker = NewBool(true)
	}

	if s.EnableGifPicker == nil {
		s.EnableGifPicker = NewBool(false)
	}

	if s.GfycatApiKey == nil || *s.GfycatApiKey == "" {
		s.GfycatApiKey = NewString(SERVICE_SETTINGS_DEFAULT_GFYCAT_API_KEY)
	}

	if s.GfycatApiSecret == nil || *s.GfycatApiSecret == "" {
		s.GfycatApiSecret = NewString(SERVICE_SETTINGS_DEFAULT_GFYCAT_API_SECRET)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictCustomEmojiCreation = NewString(RESTRICT_EMOJI_CREATION_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPostDelete == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPostDelete = NewString(PERMISSIONS_DELETE_POST_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_AllowEditPost == nil {
		s.DEPRECATED_DO_NOT_USE_AllowEditPost = NewString(ALLOW_EDIT_POST_ALWAYS)
	}

	if s.ExperimentalEnableAuthenticationTransfer == nil {
		s.ExperimentalEnableAuthenticationTransfer = NewBool(true)
	}

	if s.PostEditTimeLimit == nil {
		s.PostEditTimeLimit = NewInt(-1)
	}

	if s.EnablePreviewFeatures == nil {
		s.EnablePreviewFeatures = NewBool(true)
	}

	if s.ExperimentalEnableDefaultChannelLeaveJoinMessages == nil {
		s.ExperimentalEnableDefaultChannelLeaveJoinMessages = NewBool(true)
	}

	if s.ExperimentalGroupUnreadChannels == nil {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DISABLED)
	} else if *s.ExperimentalGroupUnreadChannels == "0" {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DISABLED)
	} else if *s.ExperimentalGroupUnreadChannels == "1" {
		s.ExperimentalGroupUnreadChannels = NewString(GROUP_UNREAD_CHANNELS_DEFAULT_ON)
	}

	if s.ExperimentalChannelOrganization == nil {
		experimentalUnreadEnabled := *s.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DISABLED
		s.ExperimentalChannelOrganization = NewBool(experimentalUnreadEnabled)
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyType == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyType = NewString("")
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyURL == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyURL = NewString("")
	}

	if s.DEPRECATED_DO_NOT_USE_ImageProxyOptions == nil {
		s.DEPRECATED_DO_NOT_USE_ImageProxyOptions = NewString("")
	}

	if s.EnableAPITeamDeletion == nil {
		s.EnableAPITeamDeletion = NewBool(false)
	}

	if s.ExperimentalEnableHardenedMode == nil {
		s.ExperimentalEnableHardenedMode = NewBool(false)
	}

	if s.DisableLegacyMFA == nil {
		s.DisableLegacyMFA = NewBool(false)
	}

	if s.ExperimentalLdapGroupSync == nil {
		s.ExperimentalLdapGroupSync = NewBool(false)
	}

	if s.ExperimentalStrictCSRFEnforcement == nil {
		s.ExperimentalStrictCSRFEnforcement = NewBool(false)
	}

	if s.DisableBotsWhenOwnerIsDeactivated == nil {
		s.DisableBotsWhenOwnerIsDeactivated = NewBool(true)
	}
}

type ClusterSettings struct {
	Enable                      *bool   `restricted:"true"`
	ClusterName                 *string `restricted:"true"`
	OverrideHostname            *string `restricted:"true"`
	UseIpAddress                *bool   `restricted:"true"`
	UseExperimentalGossip       *bool   `restricted:"true"`
	ReadOnlyConfig              *bool   `restricted:"true"`
	GossipPort                  *int    `restricted:"true"`
	StreamingPort               *int    `restricted:"true"`
	MaxIdleConns                *int    `restricted:"true"`
	MaxIdleConnsPerHost         *int    `restricted:"true"`
	IdleConnTimeoutMilliseconds *int    `restricted:"true"`
}

func (s *ClusterSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.ClusterName == nil {
		s.ClusterName = NewString("")
	}

	if s.OverrideHostname == nil {
		s.OverrideHostname = NewString("")
	}

	if s.UseIpAddress == nil {
		s.UseIpAddress = NewBool(true)
	}

	if s.UseExperimentalGossip == nil {
		s.UseExperimentalGossip = NewBool(false)
	}

	if s.ReadOnlyConfig == nil {
		s.ReadOnlyConfig = NewBool(true)
	}

	if s.GossipPort == nil {
		s.GossipPort = NewInt(8074)
	}

	if s.StreamingPort == nil {
		s.StreamingPort = NewInt(8075)
	}

	if s.MaxIdleConns == nil {
		s.MaxIdleConns = NewInt(100)
	}

	if s.MaxIdleConnsPerHost == nil {
		s.MaxIdleConnsPerHost = NewInt(128)
	}

	if s.IdleConnTimeoutMilliseconds == nil {
		s.IdleConnTimeoutMilliseconds = NewInt(90000)
	}
}

type ExperimentalSettings struct {
	ClientSideCertEnable            *bool
	ClientSideCertCheck             *string
	DisablePostMetadata             *bool  `restricted:"true"`
	EnableClickToReply              *bool  `restricted:"true"`
	LinkMetadataTimeoutMilliseconds *int64 `restricted:"true"`
	RestrictSystemAdmin             *bool  `restricted:"true"`
}

func (s *ExperimentalSettings) SetDefaults() {
	if s.ClientSideCertEnable == nil {
		s.ClientSideCertEnable = NewBool(false)
	}

	if s.ClientSideCertCheck == nil {
		s.ClientSideCertCheck = NewString(CLIENT_SIDE_CERT_CHECK_SECONDARY_AUTH)
	}

	if s.DisablePostMetadata == nil {
		s.DisablePostMetadata = NewBool(false)
	}

	if s.EnableClickToReply == nil {
		s.EnableClickToReply = NewBool(false)
	}

	if s.LinkMetadataTimeoutMilliseconds == nil {
		s.LinkMetadataTimeoutMilliseconds = NewInt64(EXPERIMENTAL_SETTINGS_DEFAULT_LINK_METADATA_TIMEOUT_MILLISECONDS)
	}

	if s.RestrictSystemAdmin == nil {
		s.RestrictSystemAdmin = NewBool(false)
	}
}

type AnalyticsSettings struct {
	MaxUsersForStatistics *int `restricted:"true"`
}

func (s *AnalyticsSettings) SetDefaults() {
	if s.MaxUsersForStatistics == nil {
		s.MaxUsersForStatistics = NewInt(ANALYTICS_SETTINGS_DEFAULT_MAX_USERS_FOR_STATISTICS)
	}
}

type SSOSettings struct {
	Enable          *bool
	Secret          *string
	Id              *string
	Scope           *string
	AuthEndpoint    *string
	TokenEndpoint   *string
	UserApiEndpoint *string
}

func (s *SSOSettings) setDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.Secret == nil {
		s.Secret = NewString("")
	}

	if s.Id == nil {
		s.Id = NewString("")
	}

	if s.Scope == nil {
		s.Scope = NewString("")
	}

	if s.AuthEndpoint == nil {
		s.AuthEndpoint = NewString("")
	}

	if s.TokenEndpoint == nil {
		s.TokenEndpoint = NewString("")
	}

	if s.UserApiEndpoint == nil {
		s.UserApiEndpoint = NewString("")
	}
}

type SqlSettings struct {
	DriverName                  *string  `restricted:"true"`
	DataSource                  *string  `restricted:"true"`
	DataSourceReplicas          []string `restricted:"true"`
	DataSourceSearchReplicas    []string `restricted:"true"`
	MaxIdleConns                *int     `restricted:"true"`
	ConnMaxLifetimeMilliseconds *int     `restricted:"true"`
	MaxOpenConns                *int     `restricted:"true"`
	Trace                       *bool    `restricted:"true"`
	AtRestEncryptKey            *string  `restricted:"true"`
	QueryTimeout                *int     `restricted:"true"`
}

func (s *SqlSettings) SetDefaults() {
	if s.DriverName == nil {
		s.DriverName = NewString(DATABASE_DRIVER_MYSQL)
	}

	if s.DataSource == nil {
		s.DataSource = NewString(SQL_SETTINGS_DEFAULT_DATA_SOURCE)
	}

	if s.DataSourceReplicas == nil {
		s.DataSourceReplicas = []string{}
	}

	if s.DataSourceSearchReplicas == nil {
		s.DataSourceSearchReplicas = []string{}
	}

	if s.AtRestEncryptKey == nil || len(*s.AtRestEncryptKey) == 0 {
		s.AtRestEncryptKey = NewString(NewRandomString(32))
	}

	if s.MaxIdleConns == nil {
		s.MaxIdleConns = NewInt(20)
	}

	if s.MaxOpenConns == nil {
		s.MaxOpenConns = NewInt(300)
	}

	if s.ConnMaxLifetimeMilliseconds == nil {
		s.ConnMaxLifetimeMilliseconds = NewInt(3600000)
	}

	if s.Trace == nil {
		s.Trace = NewBool(false)
	}

	if s.QueryTimeout == nil {
		s.QueryTimeout = NewInt(30)
	}
}

type LogSettings struct {
	EnableConsole          *bool   `restricted:"true"`
	ConsoleLevel           *string `restricted:"true"`
	ConsoleJson            *bool   `restricted:"true"`
	EnableFile             *bool   `restricted:"true"`
	FileLevel              *string `restricted:"true"`
	FileJson               *bool   `restricted:"true"`
	FileLocation           *string `restricted:"true"`
	EnableWebhookDebugging *bool   `restricted:"true"`
	EnableDiagnostics      *bool   `restricted:"true"`
}

func (s *LogSettings) SetDefaults() {
	if s.EnableConsole == nil {
		s.EnableConsole = NewBool(true)
	}

	if s.ConsoleLevel == nil {
		s.ConsoleLevel = NewString("DEBUG")
	}

	if s.EnableFile == nil {
		s.EnableFile = NewBool(true)
	}

	if s.FileLevel == nil {
		s.FileLevel = NewString("INFO")
	}

	if s.FileLocation == nil {
		s.FileLocation = NewString("")
	}

	if s.EnableWebhookDebugging == nil {
		s.EnableWebhookDebugging = NewBool(true)
	}

	if s.EnableDiagnostics == nil {
		s.EnableDiagnostics = NewBool(true)
	}

	if s.ConsoleJson == nil {
		s.ConsoleJson = NewBool(true)
	}

	if s.FileJson == nil {
		s.FileJson = NewBool(true)
	}
}

type PasswordSettings struct {
	MinimumLength *int
	Lowercase     *bool
	Number        *bool
	Uppercase     *bool
	Symbol        *bool
}

func (s *PasswordSettings) SetDefaults() {
	if s.MinimumLength == nil {
		s.MinimumLength = NewInt(PASSWORD_MINIMUM_LENGTH)
	}

	if s.Lowercase == nil {
		s.Lowercase = NewBool(false)
	}

	if s.Number == nil {
		s.Number = NewBool(false)
	}

	if s.Uppercase == nil {
		s.Uppercase = NewBool(false)
	}

	if s.Symbol == nil {
		s.Symbol = NewBool(false)
	}
}

type FileSettings struct {
	EnableFileAttachments   *bool
	EnableMobileUpload      *bool
	EnableMobileDownload    *bool
	MaxFileSize             *int64
	DriverName              *string `restricted:"true"`
	Directory               *string `restricted:"true"`
	EnablePublicLink        *bool
	PublicLinkSalt          *string
	InitialFont             *string
	AmazonS3AccessKeyId     *string `restricted:"true"`
	AmazonS3SecretAccessKey *string `restricted:"true"`
	AmazonS3Bucket          *string `restricted:"true"`
	AmazonS3Region          *string `restricted:"true"`
	AmazonS3Endpoint        *string `restricted:"true"`
	AmazonS3SSL             *bool   `restricted:"true"`
	AmazonS3SignV2          *bool   `restricted:"true"`
	AmazonS3SSE             *bool   `restricted:"true"`
	AmazonS3Trace           *bool   `restricted:"true"`
}

type PaymentBackendSettings struct {
	Backend    *string `restricted:"true"`
	Password   *string `restricted:"true"`
	UserName   *string `restricted:"true"`
	MerchantId *string `restricted:"true"`
	Currency   *string `restricted:"true"`
	Language   *string `restricted:"true"`
	Sandbox    *bool
}

func (s *FileSettings) SetDefaults() {
	if s.EnableFileAttachments == nil {
		s.EnableFileAttachments = NewBool(true)
	}

	if s.EnableMobileUpload == nil {
		s.EnableMobileUpload = NewBool(true)
	}

	if s.EnableMobileDownload == nil {
		s.EnableMobileDownload = NewBool(true)
	}

	if s.MaxFileSize == nil {
		s.MaxFileSize = NewInt64(52428800) // 50 MB
	}

	if s.DriverName == nil {
		s.DriverName = NewString(IMAGE_DRIVER_LOCAL)
	}

	if s.Directory == nil {
		s.Directory = NewString(FILE_SETTINGS_DEFAULT_DIRECTORY)
	}

	if s.EnablePublicLink == nil {
		s.EnablePublicLink = NewBool(false)
	}

	if s.PublicLinkSalt == nil || len(*s.PublicLinkSalt) == 0 {
		s.PublicLinkSalt = NewString(NewRandomString(32))
	}

	if s.InitialFont == nil {
		// Defaults to "nunito-bold.ttf"
		s.InitialFont = NewString("nunito-bold.ttf")
	}

	if s.AmazonS3AccessKeyId == nil {
		s.AmazonS3AccessKeyId = NewString("")
	}

	if s.AmazonS3SecretAccessKey == nil {
		s.AmazonS3SecretAccessKey = NewString("")
	}

	if s.AmazonS3Bucket == nil {
		s.AmazonS3Bucket = NewString("")
	}

	if s.AmazonS3Region == nil {
		s.AmazonS3Region = NewString("")
	}

	if s.AmazonS3Endpoint == nil || len(*s.AmazonS3Endpoint) == 0 {
		// Defaults to "s3.amazonaws.com"
		s.AmazonS3Endpoint = NewString("s3.amazonaws.com")
	}

	if s.AmazonS3SSL == nil {
		s.AmazonS3SSL = NewBool(true) // Secure by default.
	}

	if s.AmazonS3SignV2 == nil {
		s.AmazonS3SignV2 = new(bool)
		// Signature v2 is not enabled by default.
	}

	if s.AmazonS3SSE == nil {
		s.AmazonS3SSE = NewBool(false) // Not Encrypted by default.
	}

	if s.AmazonS3Trace == nil {
		s.AmazonS3Trace = NewBool(false)
	}
}

type EmailSettings struct {
	EnableSignUpWithEmail             *bool
	EnableSignInWithEmail             *bool
	EnableSignInWithUsername          *bool
	SendEmailNotifications            *bool
	UseChannelInEmailNotifications    *bool
	RequireEmailVerification          *bool
	FeedbackName                      *string
	FeedbackEmail                     *string
	ReplyToAddress                    *string
	FeedbackOrganization              *string
	EnableSMTPAuth                    *bool   `restricted:"true"`
	SMTPUsername                      *string `restricted:"true"`
	SMTPPassword                      *string `restricted:"true"`
	SMTPServer                        *string `restricted:"true"`
	SMTPPort                          *string `restricted:"true"`
	ConnectionSecurity                *string `restricted:"true"`
	SendPushNotifications             *bool
	PushNotificationServer            *string
	PushNotificationContents          *string
	EnableEmailBatching               *bool
	EmailBatchingBufferSize           *int
	EmailBatchingInterval             *int
	EnablePreviewModeBanner           *bool
	SkipServerCertificateVerification *bool `restricted:"true"`
	EmailNotificationContentsType     *string
	LoginButtonColor                  *string
	LoginButtonBorderColor            *string
	LoginButtonTextColor              *string
}

func (s *EmailSettings) SetDefaults() {
	if s.EnableSignUpWithEmail == nil {
		s.EnableSignUpWithEmail = NewBool(true)
	}

	if s.EnableSignInWithEmail == nil {
		s.EnableSignInWithEmail = NewBool(*s.EnableSignUpWithEmail)
	}

	if s.EnableSignInWithUsername == nil {
		s.EnableSignInWithUsername = NewBool(true)
	}

	if s.SendEmailNotifications == nil {
		s.SendEmailNotifications = NewBool(true)
	}

	if s.UseChannelInEmailNotifications == nil {
		s.UseChannelInEmailNotifications = NewBool(false)
	}

	if s.RequireEmailVerification == nil {
		s.RequireEmailVerification = NewBool(false)
	}

	if s.FeedbackName == nil {
		s.FeedbackName = NewString("")
	}

	if s.FeedbackEmail == nil {
		s.FeedbackEmail = NewString("test@example.com")
	}

	if s.ReplyToAddress == nil {
		s.ReplyToAddress = NewString("test@example.com")
	}

	if s.FeedbackOrganization == nil {
		s.FeedbackOrganization = NewString(EMAIL_SETTINGS_DEFAULT_FEEDBACK_ORGANIZATION)
	}

	if s.EnableSMTPAuth == nil {
		if s.ConnectionSecurity == nil || *s.ConnectionSecurity == CONN_SECURITY_NONE {
			s.EnableSMTPAuth = NewBool(false)
		} else {
			s.EnableSMTPAuth = NewBool(true)
		}
	}

	if s.SMTPUsername == nil {
		s.SMTPUsername = NewString("")
	}

	if s.SMTPPassword == nil {
		s.SMTPPassword = NewString("")
	}

	if s.SMTPServer == nil || len(*s.SMTPServer) == 0 {
		s.SMTPServer = NewString("dockerhost")
	}

	if s.SMTPPort == nil || len(*s.SMTPPort) == 0 {
		s.SMTPPort = NewString("2500")
	}

	if s.ConnectionSecurity == nil || *s.ConnectionSecurity == CONN_SECURITY_PLAIN {
		s.ConnectionSecurity = NewString(CONN_SECURITY_NONE)
	}

	if s.SendPushNotifications == nil {
		s.SendPushNotifications = NewBool(false)
	}

	if s.PushNotificationServer == nil {
		s.PushNotificationServer = NewString("")
	}

	if s.PushNotificationContents == nil {
		s.PushNotificationContents = NewString(GENERIC_NOTIFICATION)
	}

	if s.EnableEmailBatching == nil {
		s.EnableEmailBatching = NewBool(false)
	}

	if s.EmailBatchingBufferSize == nil {
		s.EmailBatchingBufferSize = NewInt(EMAIL_BATCHING_BUFFER_SIZE)
	}

	if s.EmailBatchingInterval == nil {
		s.EmailBatchingInterval = NewInt(EMAIL_BATCHING_INTERVAL)
	}

	if s.EnablePreviewModeBanner == nil {
		s.EnablePreviewModeBanner = NewBool(true)
	}

	if s.EnableSMTPAuth == nil {
		if *s.ConnectionSecurity == CONN_SECURITY_NONE {
			s.EnableSMTPAuth = NewBool(false)
		} else {
			s.EnableSMTPAuth = NewBool(true)
		}
	}

	if *s.ConnectionSecurity == CONN_SECURITY_PLAIN {
		*s.ConnectionSecurity = CONN_SECURITY_NONE
	}

	if s.SkipServerCertificateVerification == nil {
		s.SkipServerCertificateVerification = NewBool(false)
	}

	if s.EmailNotificationContentsType == nil {
		s.EmailNotificationContentsType = NewString(EMAIL_NOTIFICATION_CONTENTS_FULL)
	}

	if s.LoginButtonColor == nil {
		s.LoginButtonColor = NewString("#0000")
	}

	if s.LoginButtonBorderColor == nil {
		s.LoginButtonBorderColor = NewString("#2389D7")
	}

	if s.LoginButtonTextColor == nil {
		s.LoginButtonTextColor = NewString("#2389D7")
	}
}

type RateLimitSettings struct {
	Enable           *bool  `restricted:"true"`
	PerSec           *int   `restricted:"true"`
	MaxBurst         *int   `restricted:"true"`
	MemoryStoreSize  *int   `restricted:"true"`
	VaryByRemoteAddr *bool  `restricted:"true"`
	VaryByUser       *bool  `restricted:"true"`
	VaryByHeader     string `restricted:"true"`
}

func (s *RateLimitSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.PerSec == nil {
		s.PerSec = NewInt(10)
	}

	if s.MaxBurst == nil {
		s.MaxBurst = NewInt(100)
	}

	if s.MemoryStoreSize == nil {
		s.MemoryStoreSize = NewInt(10000)
	}

	if s.VaryByRemoteAddr == nil {
		s.VaryByRemoteAddr = NewBool(true)
	}

	if s.VaryByUser == nil {
		s.VaryByUser = NewBool(false)
	}
}

type PrivacySettings struct {
	ShowEmailAddress *bool
	ShowFullName     *bool
}

func (s *PrivacySettings) setDefaults() {
	if s.ShowEmailAddress == nil {
		s.ShowEmailAddress = NewBool(true)
	}

	if s.ShowFullName == nil {
		s.ShowFullName = NewBool(true)
	}
}

type AnnouncementSettings struct {
	EnableBanner         *bool
	BannerText           *string
	BannerColor          *string
	BannerTextColor      *string
	AllowBannerDismissal *bool
}

func (s *AnnouncementSettings) SetDefaults() {
	if s.EnableBanner == nil {
		s.EnableBanner = NewBool(false)
	}

	if s.BannerText == nil {
		s.BannerText = NewString("")
	}

	if s.BannerColor == nil {
		s.BannerColor = NewString(ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_COLOR)
	}

	if s.BannerTextColor == nil {
		s.BannerTextColor = NewString(ANNOUNCEMENT_SETTINGS_DEFAULT_BANNER_TEXT_COLOR)
	}

	if s.AllowBannerDismissal == nil {
		s.AllowBannerDismissal = NewBool(true)
	}
}

type ThemeSettings struct {
	EnableThemeSelection *bool
	DefaultTheme         *string
	AllowCustomThemes    *bool
	AllowedThemes        []string
}

func (s *ThemeSettings) SetDefaults() {
	if s.EnableThemeSelection == nil {
		s.EnableThemeSelection = NewBool(true)
	}

	if s.DefaultTheme == nil {
		s.DefaultTheme = NewString(TEAM_SETTINGS_DEFAULT_TEAM_TEXT)
	}

	if s.AllowCustomThemes == nil {
		s.AllowCustomThemes = NewBool(true)
	}

	if s.AllowedThemes == nil {
		s.AllowedThemes = []string{}
	}
}

type TeamSettings struct {
	SiteName                                                  *string
	MaxUsersPerTeam                                           *int
	DEPRECATED_DO_NOT_USE_EnableTeamCreation                  *bool `json:"EnableTeamCreation"` // This field is deprecated and must not be used.
	EnableUserCreation                                        *bool
	EnableOpenServer                                          *bool
	EnableUserDeactivation                                    *bool
	RestrictCreationToDomains                                 *string
	EnableCustomBrand                                         *bool
	CustomBrandText                                           *string
	CustomDescriptionText                                     *string
	RestrictDirectMessage                                     *string
	DEPRECATED_DO_NOT_USE_RestrictTeamInvite                  *string `json:"RestrictTeamInvite"`                  // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement     *string `json:"RestrictPublicChannelManagement"`     // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement    *string `json:"RestrictPrivateChannelManagement"`    // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation       *string `json:"RestrictPublicChannelCreation"`       // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation      *string `json:"RestrictPrivateChannelCreation"`      // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion       *string `json:"RestrictPublicChannelDeletion"`       // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion      *string `json:"RestrictPrivateChannelDeletion"`      // This field is deprecated and must not be used.
	DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers *string `json:"RestrictPrivateChannelManageMembers"` // This field is deprecated and must not be used.
	EnableXToLeaveChannelsFromLHS                             *bool
	UserStatusAwayTimeout                                     *int64
	MaxChannelsPerTeam                                        *int64
	MaxNotificationsPerChannel                                *int64
	EnableConfirmNotificationsToChannel                       *bool
	TeammateNameDisplay                                       *string
	ExperimentalViewArchivedChannels                          *bool
	ExperimentalEnableAutomaticReplies                        *bool
	ExperimentalHideTownSquareinLHS                           *bool
	ExperimentalTownSquareIsReadOnly                          *bool
	ExperimentalPrimaryTeam                                   *string
	ExperimentalDefaultChannels                               []string
}

func (s *TeamSettings) SetDefaults() {

	if s.SiteName == nil || *s.SiteName == "" {
		s.SiteName = NewString(TEAM_SETTINGS_DEFAULT_SITE_NAME)
	}

	if s.MaxUsersPerTeam == nil {
		s.MaxUsersPerTeam = NewInt(TEAM_SETTINGS_DEFAULT_MAX_USERS_PER_TEAM)
	}

	if s.DEPRECATED_DO_NOT_USE_EnableTeamCreation == nil {
		s.DEPRECATED_DO_NOT_USE_EnableTeamCreation = NewBool(true)
	}

	if s.EnableUserCreation == nil {
		s.EnableUserCreation = NewBool(true)
	}

	if s.EnableOpenServer == nil {
		s.EnableOpenServer = NewBool(false)
	}

	if s.RestrictCreationToDomains == nil {
		s.RestrictCreationToDomains = NewString("")
	}

	if s.EnableCustomBrand == nil {
		s.EnableCustomBrand = NewBool(false)
	}

	if s.EnableUserDeactivation == nil {
		s.EnableUserDeactivation = NewBool(false)
	}

	if s.CustomBrandText == nil {
		s.CustomBrandText = NewString(TEAM_SETTINGS_DEFAULT_CUSTOM_BRAND_TEXT)
	}

	if s.CustomDescriptionText == nil {
		s.CustomDescriptionText = NewString(TEAM_SETTINGS_DEFAULT_CUSTOM_DESCRIPTION_TEXT)
	}

	if s.RestrictDirectMessage == nil {
		s.RestrictDirectMessage = NewString(DIRECT_MESSAGE_ANY)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictTeamInvite == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictTeamInvite = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement = NewString(PERMISSIONS_ALL)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = new(string)
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		if *s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement == PERMISSIONS_CHANNEL_ADMIN {
			*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = PERMISSIONS_TEAM_ADMIN
		} else {
			*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelCreation = *s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement
		}
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		if *s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement == PERMISSIONS_CHANNEL_ADMIN {
			s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation = NewString(PERMISSIONS_TEAM_ADMIN)
		} else {
			s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelCreation = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement)
		}
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelDeletion = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPublicChannelManagement)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion == nil {
		// If this setting does not exist, assume migration from <3.6, so use management setting as default.
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelDeletion = NewString(*s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManagement)
	}

	if s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers == nil {
		s.DEPRECATED_DO_NOT_USE_RestrictPrivateChannelManageMembers = NewString(PERMISSIONS_ALL)
	}

	if s.EnableXToLeaveChannelsFromLHS == nil {
		s.EnableXToLeaveChannelsFromLHS = NewBool(false)
	}

	if s.UserStatusAwayTimeout == nil {
		s.UserStatusAwayTimeout = NewInt64(TEAM_SETTINGS_DEFAULT_USER_STATUS_AWAY_TIMEOUT)
	}

	if s.MaxChannelsPerTeam == nil {
		s.MaxChannelsPerTeam = NewInt64(2000)
	}

	if s.MaxNotificationsPerChannel == nil {
		s.MaxNotificationsPerChannel = NewInt64(1000)
	}

	if s.EnableConfirmNotificationsToChannel == nil {
		s.EnableConfirmNotificationsToChannel = NewBool(true)
	}

	if s.ExperimentalEnableAutomaticReplies == nil {
		s.ExperimentalEnableAutomaticReplies = NewBool(false)
	}

	if s.ExperimentalHideTownSquareinLHS == nil {
		s.ExperimentalHideTownSquareinLHS = NewBool(false)
	}

	if s.ExperimentalTownSquareIsReadOnly == nil {
		s.ExperimentalTownSquareIsReadOnly = NewBool(false)
	}

	if s.ExperimentalPrimaryTeam == nil {
		s.ExperimentalPrimaryTeam = NewString("")
	}

	if s.ExperimentalDefaultChannels == nil {
		s.ExperimentalDefaultChannels = []string{}
	}

	if s.DEPRECATED_DO_NOT_USE_EnableTeamCreation == nil {
		s.DEPRECATED_DO_NOT_USE_EnableTeamCreation = NewBool(true)
	}

	if s.EnableUserCreation == nil {
		s.EnableUserCreation = NewBool(true)
	}

	if s.ExperimentalViewArchivedChannels == nil {
		s.ExperimentalViewArchivedChannels = NewBool(false)
	}
}

type ClientRequirements struct {
	AndroidLatestVersion string `restricted:"true"`
	AndroidMinVersion    string `restricted:"true"`
	DesktopLatestVersion string `restricted:"true"`
	DesktopMinVersion    string `restricted:"true"`
	IosLatestVersion     string `restricted:"true"`
	IosMinVersion        string `restricted:"true"`
}

type LocalizationSettings struct {
	DefaultServerLocale *string
	DefaultClientLocale *string
	AvailableLocales    *string
}

func (s *LocalizationSettings) SetDefaults() {
	if s.DefaultServerLocale == nil {
		s.DefaultServerLocale = NewString(DEFAULT_LOCALE)
	}

	if s.DefaultClientLocale == nil {
		s.DefaultClientLocale = NewString(DEFAULT_LOCALE)
	}

	if s.AvailableLocales == nil {
		s.AvailableLocales = NewString("")
	}
}

type ElasticsearchSettings struct {
	ConnectionUrl                 *string `restricted:"true"`
	Username                      *string `restricted:"true"`
	Password                      *string `restricted:"true"`
	EnableIndexing                *bool   `restricted:"true"`
	EnableSearching               *bool   `restricted:"true"`
	EnableAutocomplete            *bool   `restricted:"true"`
	Sniff                         *bool   `restricted:"true"`
	PostIndexReplicas             *int    `restricted:"true"`
	PostIndexShards               *int    `restricted:"true"`
	ChannelIndexReplicas          *int    `restricted:"true"`
	ChannelIndexShards            *int    `restricted:"true"`
	UserIndexReplicas             *int    `restricted:"true"`
	UserIndexShards               *int    `restricted:"true"`
	AggregatePostsAfterDays       *int    `restricted:"true"`
	PostsAggregatorJobStartTime   *string `restricted:"true"`
	IndexPrefix                   *string `restricted:"true"`
	LiveIndexingBatchSize         *int    `restricted:"true"`
	BulkIndexingTimeWindowSeconds *int    `restricted:"true"`
	RequestTimeoutSeconds         *int    `restricted:"true"`
}

func (s *ElasticsearchSettings) SetDefaults() {
	if s.ConnectionUrl == nil {
		s.ConnectionUrl = NewString(ELASTICSEARCH_SETTINGS_DEFAULT_CONNECTION_URL)
	}

	if s.Username == nil {
		s.Username = NewString(ELASTICSEARCH_SETTINGS_DEFAULT_USERNAME)
	}

	if s.Password == nil {
		s.Password = NewString(ELASTICSEARCH_SETTINGS_DEFAULT_PASSWORD)
	}

	if s.EnableIndexing == nil {
		s.EnableIndexing = NewBool(false)
	}

	if s.EnableSearching == nil {
		s.EnableSearching = NewBool(false)
	}

	if s.EnableAutocomplete == nil {
		s.EnableAutocomplete = NewBool(false)
	}

	if s.Sniff == nil {
		s.Sniff = NewBool(true)
	}

	if s.PostIndexReplicas == nil {
		s.PostIndexReplicas = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_POST_INDEX_REPLICAS)
	}

	if s.PostIndexShards == nil {
		s.PostIndexShards = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_POST_INDEX_SHARDS)
	}

	if s.ChannelIndexReplicas == nil {
		s.ChannelIndexReplicas = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_CHANNEL_INDEX_REPLICAS)
	}

	if s.ChannelIndexShards == nil {
		s.ChannelIndexShards = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_CHANNEL_INDEX_SHARDS)
	}

	if s.UserIndexReplicas == nil {
		s.UserIndexReplicas = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_USER_INDEX_REPLICAS)
	}

	if s.UserIndexShards == nil {
		s.UserIndexShards = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_USER_INDEX_SHARDS)
	}

	if s.AggregatePostsAfterDays == nil {
		s.AggregatePostsAfterDays = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_AGGREGATE_POSTS_AFTER_DAYS)
	}

	if s.PostsAggregatorJobStartTime == nil {
		s.PostsAggregatorJobStartTime = NewString(ELASTICSEARCH_SETTINGS_DEFAULT_POSTS_AGGREGATOR_JOB_START_TIME)
	}

	if s.IndexPrefix == nil {
		s.IndexPrefix = NewString(ELASTICSEARCH_SETTINGS_DEFAULT_INDEX_PREFIX)
	}

	if s.LiveIndexingBatchSize == nil {
		s.LiveIndexingBatchSize = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_LIVE_INDEXING_BATCH_SIZE)
	}

	if s.BulkIndexingTimeWindowSeconds == nil {
		s.BulkIndexingTimeWindowSeconds = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_BULK_INDEXING_TIME_WINDOW_SECONDS)
	}

	if s.RequestTimeoutSeconds == nil {
		s.RequestTimeoutSeconds = NewInt(ELASTICSEARCH_SETTINGS_DEFAULT_REQUEST_TIMEOUT_SECONDS)
	}
}

type DataRetentionSettings struct {
	EnableMessageDeletion *bool
	EnableFileDeletion    *bool
	MessageRetentionDays  *int
	FileRetentionDays     *int
	DeletionJobStartTime  *string
}

func (s *DataRetentionSettings) SetDefaults() {
	if s.EnableMessageDeletion == nil {
		s.EnableMessageDeletion = NewBool(false)
	}

	if s.EnableFileDeletion == nil {
		s.EnableFileDeletion = NewBool(false)
	}

	if s.MessageRetentionDays == nil {
		s.MessageRetentionDays = NewInt(DATA_RETENTION_SETTINGS_DEFAULT_MESSAGE_RETENTION_DAYS)
	}

	if s.FileRetentionDays == nil {
		s.FileRetentionDays = NewInt(DATA_RETENTION_SETTINGS_DEFAULT_FILE_RETENTION_DAYS)
	}

	if s.DeletionJobStartTime == nil {
		s.DeletionJobStartTime = NewString(DATA_RETENTION_SETTINGS_DEFAULT_DELETION_JOB_START_TIME)
	}
}

type JobSettings struct {
	RunJobs      *bool `restricted:"true"`
	RunScheduler *bool `restricted:"true"`
}

func (s *JobSettings) SetDefaults() {
	if s.RunJobs == nil {
		s.RunJobs = NewBool(true)
	}

	if s.RunScheduler == nil {
		s.RunScheduler = NewBool(true)
	}
}

type DisplaySettings struct {
	CustomUrlSchemes     []string
	ExperimentalTimezone *bool
}

func (s *DisplaySettings) SetDefaults() {
	if s.CustomUrlSchemes == nil {
		customUrlSchemes := []string{}
		s.CustomUrlSchemes = customUrlSchemes
	}

	if s.ExperimentalTimezone == nil {
		s.ExperimentalTimezone = NewBool(false)
	}
}

type ImageProxySettings struct {
	Enable                  *bool
	ImageProxyType          *string
	RemoteImageProxyURL     *string
	RemoteImageProxyOptions *string
}

func (ips *ImageProxySettings) SetDefaults(ss ServiceSettings) {
	if ips.Enable == nil {
		if ss.DEPRECATED_DO_NOT_USE_ImageProxyType == nil || *ss.DEPRECATED_DO_NOT_USE_ImageProxyType == "" {
			ips.Enable = NewBool(false)
		} else {
			ips.Enable = NewBool(true)
		}
	}

	if ips.ImageProxyType == nil {
		if ss.DEPRECATED_DO_NOT_USE_ImageProxyType == nil || *ss.DEPRECATED_DO_NOT_USE_ImageProxyType == "" {
			ips.ImageProxyType = NewString(IMAGE_PROXY_TYPE_LOCAL)
		} else {
			ips.ImageProxyType = ss.DEPRECATED_DO_NOT_USE_ImageProxyType
		}
	}

	if ips.RemoteImageProxyURL == nil {
		if ss.DEPRECATED_DO_NOT_USE_ImageProxyURL == nil {
			ips.RemoteImageProxyURL = NewString("")
		} else {
			ips.RemoteImageProxyURL = ss.DEPRECATED_DO_NOT_USE_ImageProxyURL
		}
	}

	if ips.RemoteImageProxyOptions == nil {
		if ss.DEPRECATED_DO_NOT_USE_ImageProxyOptions == nil {
			ips.RemoteImageProxyOptions = NewString("")
		} else {
			ips.RemoteImageProxyOptions = ss.DEPRECATED_DO_NOT_USE_ImageProxyOptions
		}
	}
}

type ConfigFunc func() *Config

type Config struct {
	ServiceSettings      ServiceSettings
	TeamSettings         TeamSettings
	ClientRequirements   ClientRequirements
	SqlSettings          SqlSettings
	LogSettings          LogSettings
	PasswordSettings     PasswordSettings
	FileSettings         FileSettings
	EmailSettings        EmailSettings
	RateLimitSettings    RateLimitSettings
	PrivacySettings      PrivacySettings
	AnnouncementSettings AnnouncementSettings
	ThemeSettings        ThemeSettings
	GitLabSettings       SSOSettings
	GoogleSettings       SSOSettings
	Office365Settings    SSOSettings

	LocalizationSettings LocalizationSettings

	ClusterSettings ClusterSettings

	ExperimentalSettings  ExperimentalSettings
	AnalyticsSettings     AnalyticsSettings
	ElasticsearchSettings ElasticsearchSettings
	DataRetentionSettings DataRetentionSettings

	JobSettings JobSettings

	DisplaySettings    DisplaySettings
	ImageProxySettings ImageProxySettings
}

func (o *Config) Clone() *Config {
	var ret Config
	if err := json.Unmarshal([]byte(o.ToJson()), &ret); err != nil {
		panic(err)
	}
	return &ret
}

func (o *Config) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Config) GetSSOService(service string) *SSOSettings {
	switch service {
	case SERVICE_GITLAB:
		return &o.GitLabSettings
	case SERVICE_GOOGLE:
		return &o.GoogleSettings
	case SERVICE_OFFICE365:
		return &o.Office365Settings
	}

	return nil
}

func ConfigFromJson(data io.Reader) *Config {
	var o *Config
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *Config) SetDefaults() {

	if o.TeamSettings.TeammateNameDisplay == nil {
		o.TeamSettings.TeammateNameDisplay = NewString(SHOW_USERNAME)

	}

	o.SqlSettings.SetDefaults()
	o.FileSettings.SetDefaults()
	o.EmailSettings.SetDefaults()
	o.PrivacySettings.setDefaults()
	o.Office365Settings.setDefaults()
	o.GitLabSettings.setDefaults()
	o.GoogleSettings.setDefaults()
	o.ServiceSettings.SetDefaults()
	o.PasswordSettings.SetDefaults()
	o.TeamSettings.SetDefaults()
	o.ExperimentalSettings.SetDefaults()
	o.AnnouncementSettings.SetDefaults()
	o.ThemeSettings.SetDefaults()
	o.ClusterSettings.SetDefaults()
	o.AnalyticsSettings.SetDefaults()
	o.LocalizationSettings.SetDefaults()
	o.ElasticsearchSettings.SetDefaults()
	o.DataRetentionSettings.SetDefaults()
	o.RateLimitSettings.SetDefaults()
	o.LogSettings.SetDefaults()
	o.JobSettings.SetDefaults()
	o.DisplaySettings.SetDefaults()
	o.ImageProxySettings.SetDefaults(o.ServiceSettings)
}

func (o *Config) IsValid() *AppError {
	if len(*o.ServiceSettings.SiteURL) == 0 && *o.EmailSettings.EnableEmailBatching {
		return NewAppError("Config.IsValid", "model.config.is_valid.site_url_email_batching.app_error", nil, "", http.StatusBadRequest)
	}

	if *o.ClusterSettings.Enable && *o.EmailSettings.EnableEmailBatching {
		return NewAppError("Config.IsValid", "model.config.is_valid.cluster_email_batching.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*o.ServiceSettings.SiteURL) == 0 && *o.ServiceSettings.AllowCookiesForSubdomains {
		return NewAppError("Config.IsValid", "model.config.is_valid.allow_cookies_for_subdomains.app_error", nil, "", http.StatusBadRequest)
	}

	if err := o.TeamSettings.isValid(); err != nil {
		return err
	}

	if err := o.SqlSettings.isValid(); err != nil {
		return err
	}

	if err := o.FileSettings.isValid(); err != nil {
		return err
	}

	if err := o.EmailSettings.isValid(); err != nil {
		return err
	}

	if *o.PasswordSettings.MinimumLength < PASSWORD_MINIMUM_LENGTH || *o.PasswordSettings.MinimumLength > PASSWORD_MAXIMUM_LENGTH {
		return NewAppError("Config.IsValid", "model.config.is_valid.password_length.app_error", map[string]interface{}{"MinLength": PASSWORD_MINIMUM_LENGTH, "MaxLength": PASSWORD_MAXIMUM_LENGTH}, "", http.StatusBadRequest)
	}

	if err := o.RateLimitSettings.isValid(); err != nil {
		return err
	}

	if err := o.ServiceSettings.isValid(); err != nil {
		return err
	}

	if err := o.ElasticsearchSettings.isValid(); err != nil {
		return err
	}

	if err := o.DataRetentionSettings.isValid(); err != nil {
		return err
	}

	if err := o.LocalizationSettings.isValid(); err != nil {
		return err
	}

	if err := o.DisplaySettings.isValid(); err != nil {
		return err
	}

	if err := o.ImageProxySettings.isValid(); err != nil {
		return err
	}

	return nil
}

func (ts *TeamSettings) isValid() *AppError {
	if *ts.MaxUsersPerTeam <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.max_users.app_error", nil, "", http.StatusBadRequest)
	}

	if *ts.MaxChannelsPerTeam <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.max_channels.app_error", nil, "", http.StatusBadRequest)
	}

	if *ts.MaxNotificationsPerChannel <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.max_notify_per_channel.app_error", nil, "", http.StatusBadRequest)
	}

	if !(*ts.RestrictDirectMessage == DIRECT_MESSAGE_ANY || *ts.RestrictDirectMessage == DIRECT_MESSAGE_TEAM) {
		return NewAppError("Config.IsValid", "model.config.is_valid.restrict_direct_message.app_error", nil, "", http.StatusBadRequest)
	}

	if !(*ts.TeammateNameDisplay == SHOW_FULLNAME || *ts.TeammateNameDisplay == SHOW_NICKNAME_FULLNAME || *ts.TeammateNameDisplay == SHOW_USERNAME) {
		return NewAppError("Config.IsValid", "model.config.is_valid.teammate_name_display.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*ts.SiteName) == 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sitename_empty.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*ts.SiteName) > SITENAME_MAX_LENGTH {
		return NewAppError("Config.IsValid", "model.config.is_valid.sitename_length.app_error", map[string]interface{}{"MaxLength": SITENAME_MAX_LENGTH}, "", http.StatusBadRequest)
	}

	return nil
}

func (ss *SqlSettings) isValid() *AppError {
	if len(*ss.AtRestEncryptKey) < 32 {
		return NewAppError("Config.IsValid", "model.config.is_valid.encrypt_sql.app_error", nil, "", http.StatusBadRequest)
	}

	if !(*ss.DriverName == DATABASE_DRIVER_MYSQL || *ss.DriverName == DATABASE_DRIVER_POSTGRES) {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_driver.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.MaxIdleConns <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_idle.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.ConnMaxLifetimeMilliseconds < 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_conn_max_lifetime_milliseconds.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.QueryTimeout <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_query_timeout.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*ss.DataSource) == 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_data_src.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.MaxOpenConns <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_max_conn.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (fs *FileSettings) isValid() *AppError {
	if *fs.MaxFileSize <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.max_file_size.app_error", nil, "", http.StatusBadRequest)
	}

	if !(*fs.DriverName == IMAGE_DRIVER_LOCAL || *fs.DriverName == IMAGE_DRIVER_S3) {
		return NewAppError("Config.IsValid", "model.config.is_valid.file_driver.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*fs.PublicLinkSalt) < 32 {
		return NewAppError("Config.IsValid", "model.config.is_valid.file_salt.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (es *EmailSettings) isValid() *AppError {
	if !(*es.ConnectionSecurity == CONN_SECURITY_NONE || *es.ConnectionSecurity == CONN_SECURITY_TLS || *es.ConnectionSecurity == CONN_SECURITY_STARTTLS || *es.ConnectionSecurity == CONN_SECURITY_PLAIN) {
		return NewAppError("Config.IsValid", "model.config.is_valid.email_security.app_error", nil, "", http.StatusBadRequest)
	}

	if *es.EmailBatchingBufferSize <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.email_batching_buffer_size.app_error", nil, "", http.StatusBadRequest)
	}

	if *es.EmailBatchingInterval < 30 {
		return NewAppError("Config.IsValid", "model.config.is_valid.email_batching_interval.app_error", nil, "", http.StatusBadRequest)
	}

	if !(*es.EmailNotificationContentsType == EMAIL_NOTIFICATION_CONTENTS_FULL || *es.EmailNotificationContentsType == EMAIL_NOTIFICATION_CONTENTS_GENERIC) {
		return NewAppError("Config.IsValid", "model.config.is_valid.email_notification_contents_type.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (rls *RateLimitSettings) isValid() *AppError {
	if *rls.MemoryStoreSize <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.rate_mem.app_error", nil, "", http.StatusBadRequest)
	}

	if *rls.PerSec <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.rate_sec.app_error", nil, "", http.StatusBadRequest)
	}

	if *rls.MaxBurst <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.max_burst.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (ss *ServiceSettings) isValid() *AppError {
	if !(*ss.ConnectionSecurity == CONN_SECURITY_NONE || *ss.ConnectionSecurity == CONN_SECURITY_TLS) {
		return NewAppError("Config.IsValid", "model.config.is_valid.webserver_security.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.ConnectionSecurity == CONN_SECURITY_TLS && *ss.UseLetsEncrypt == false {
		appErr := NewAppError("Config.IsValid", "model.config.is_valid.tls_cert_file.app_error", nil, "", http.StatusBadRequest)

		if *ss.TLSCertFile == "" {
			return appErr
		} else if _, err := os.Stat(*ss.TLSCertFile); os.IsNotExist(err) {
			return appErr
		}

		appErr = NewAppError("Config.IsValid", "model.config.is_valid.tls_key_file.app_error", nil, "", http.StatusBadRequest)

		if *ss.TLSKeyFile == "" {
			return appErr
		} else if _, err := os.Stat(*ss.TLSKeyFile); os.IsNotExist(err) {
			return appErr
		}
	}

	if len(ss.TLSOverwriteCiphers) > 0 {
		for _, cipher := range ss.TLSOverwriteCiphers {
			if _, ok := ServerTLSSupportedCiphers[cipher]; !ok {
				return NewAppError("Config.IsValid", "model.config.is_valid.tls_overwrite_cipher.app_error", map[string]interface{}{"name": cipher}, "", http.StatusBadRequest)
			}
		}
	}

	if *ss.ReadTimeout <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.read_timeout.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.WriteTimeout <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.write_timeout.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.TimeBetweenUserTypingUpdatesMilliseconds < 1000 {
		return NewAppError("Config.IsValid", "model.config.is_valid.time_between_user_typing.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.MaximumLoginAttempts <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.login_attempts.app_error", nil, "", http.StatusBadRequest)
	}

	if len(*ss.SiteURL) != 0 {
		if _, err := url.ParseRequestURI(*ss.SiteURL); err != nil {
			return NewAppError("Config.IsValid", "model.config.is_valid.site_url.app_error", nil, "", http.StatusBadRequest)
		}
	}

	if len(*ss.WebsocketURL) != 0 {
		if _, err := url.ParseRequestURI(*ss.WebsocketURL); err != nil {
			return NewAppError("Config.IsValid", "model.config.is_valid.websocket_url.app_error", nil, "", http.StatusBadRequest)
		}
	}

	host, port, _ := net.SplitHostPort(*ss.ListenAddress)
	var isValidHost bool
	if host == "" {
		isValidHost = true
	} else {
		isValidHost = (net.ParseIP(host) != nil) || IsDomainName(host)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil || !isValidHost || portInt < 0 || portInt > math.MaxUint16 {
		return NewAppError("Config.IsValid", "model.config.is_valid.listen_address.app_error", nil, "", http.StatusBadRequest)
	}

	if *ss.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DISABLED &&
		*ss.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DEFAULT_ON &&
		*ss.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DEFAULT_OFF {
		return NewAppError("Config.IsValid", "model.config.is_valid.group_unread_channels.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (ess *ElasticsearchSettings) isValid() *AppError {
	if *ess.EnableIndexing {
		if len(*ess.ConnectionUrl) == 0 {
			return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.connection_url.app_error", nil, "", http.StatusBadRequest)
		}
	}

	if *ess.EnableSearching && !*ess.EnableIndexing {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.enable_searching.app_error", nil, "", http.StatusBadRequest)
	}

	if *ess.EnableAutocomplete && !*ess.EnableIndexing {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.enable_autocomplete.app_error", nil, "", http.StatusBadRequest)
	}

	if *ess.AggregatePostsAfterDays < 1 {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.aggregate_posts_after_days.app_error", nil, "", http.StatusBadRequest)
	}

	if _, err := time.Parse("15:04", *ess.PostsAggregatorJobStartTime); err != nil {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.posts_aggregator_job_start_time.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	if *ess.LiveIndexingBatchSize < 1 {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.live_indexing_batch_size.app_error", nil, "", http.StatusBadRequest)
	}

	if *ess.BulkIndexingTimeWindowSeconds < 1 {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.bulk_indexing_time_window_seconds.app_error", nil, "", http.StatusBadRequest)
	}

	if *ess.RequestTimeoutSeconds < 1 {
		return NewAppError("Config.IsValid", "model.config.is_valid.elastic_search.request_timeout_seconds.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (drs *DataRetentionSettings) isValid() *AppError {
	if *drs.MessageRetentionDays <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.data_retention.message_retention_days_too_low.app_error", nil, "", http.StatusBadRequest)
	}

	if *drs.FileRetentionDays <= 0 {
		return NewAppError("Config.IsValid", "model.config.is_valid.data_retention.file_retention_days_too_low.app_error", nil, "", http.StatusBadRequest)
	}

	if _, err := time.Parse("15:04", *drs.DeletionJobStartTime); err != nil {
		return NewAppError("Config.IsValid", "model.config.is_valid.data_retention.deletion_job_start_time.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (ls *LocalizationSettings) isValid() *AppError {
	if len(*ls.AvailableLocales) > 0 {
		if !strings.Contains(*ls.AvailableLocales, *ls.DefaultClientLocale) {
			return NewAppError("Config.IsValid", "model.config.is_valid.localization.available_locales.app_error", nil, "", http.StatusBadRequest)
		}
	}

	return nil
}

func (ds *DisplaySettings) isValid() *AppError {
	if len(ds.CustomUrlSchemes) != 0 {
		validProtocolPattern := regexp.MustCompile(`(?i)^\s*[a-z][a-z0-9-]*\s*$`)

		for _, scheme := range ds.CustomUrlSchemes {
			if !validProtocolPattern.MatchString(scheme) {
				return NewAppError(
					"Config.IsValid",
					"model.config.is_valid.display.custom_url_schemes.app_error",
					map[string]interface{}{"Scheme": scheme},
					"",
					http.StatusBadRequest,
				)
			}
		}
	}

	return nil
}

func (ips *ImageProxySettings) isValid() *AppError {
	if *ips.Enable {
		switch *ips.ImageProxyType {
		case IMAGE_PROXY_TYPE_LOCAL:
			// No other settings to validate
		case IMAGE_PROXY_TYPE_ATMOS_CAMO:
			if *ips.RemoteImageProxyURL == "" {
				return NewAppError("Config.IsValid", "model.config.is_valid.atmos_camo_image_proxy_url.app_error", nil, "", http.StatusBadRequest)
			}

			if *ips.RemoteImageProxyOptions == "" {
				return NewAppError("Config.IsValid", "model.config.is_valid.atmos_camo_image_proxy_options.app_error", nil, "", http.StatusBadRequest)
			}
		default:
			return NewAppError("Config.IsValid", "model.config.is_valid.image_proxy_type.app_error", nil, "", http.StatusBadRequest)
		}
	}

	return nil
}

func (o *Config) GetSanitizeOptions() map[string]bool {
	options := map[string]bool{}
	options["fullname"] = *o.PrivacySettings.ShowFullName
	options["email"] = *o.PrivacySettings.ShowEmailAddress

	return options
}

func (o *Config) Sanitize() {

	*o.FileSettings.PublicLinkSalt = FAKE_SETTING
	if len(*o.FileSettings.AmazonS3SecretAccessKey) > 0 {
		*o.FileSettings.AmazonS3SecretAccessKey = FAKE_SETTING
	}

	if len(*o.GitLabSettings.Secret) > 0 {
		*o.GitLabSettings.Secret = FAKE_SETTING
	}

	*o.SqlSettings.DataSource = FAKE_SETTING
	*o.SqlSettings.AtRestEncryptKey = FAKE_SETTING

	for i := range o.SqlSettings.DataSourceReplicas {
		o.SqlSettings.DataSourceReplicas[i] = FAKE_SETTING
	}

	for i := range o.SqlSettings.DataSourceSearchReplicas {
		o.SqlSettings.DataSourceSearchReplicas[i] = FAKE_SETTING
	}

	*o.ElasticsearchSettings.Password = FAKE_SETTING
}
