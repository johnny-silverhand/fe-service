package model

import (
	"encoding/json"
	"io"
)

const (
	PUSH_NOTIFY_APPLE                = "apple"
	PUSH_NOTIFY_ANDROID              = "android"
	PUSH_NOTIFY_APPLE_REACT_NATIVE   = "apple_rn"
	PUSH_NOTIFY_ANDROID_REACT_NATIVE = "android_rn"

	PUSH_TYPE_MESSAGE = "message"
	PUSH_TYPE_CLEAR   = "clear"
	PUSH_MESSAGE_V2   = "v2"

	// The category is set to handle a set of interactive Actions
	// with the push notifications
	CATEGORY_CAN_REPLY = "CAN_REPLY"

	MHPNS = ""
)

type PushNotification struct {
	Platform         string `json:"platform"`
	ServerId         string `json:"server_id"`
	DeviceId         string `json:"device_id"`
	Category         string `json:"category"`
	Sound            string `json:"sound"`
	Message          string `json:"message"`
	Badge            int    `json:"badge"`
	ContentAvailable int    `json:"cont_ava"`
	TeamId           string `json:"team_id"`
	ChannelId        string `json:"channel_id"`
	PostId           string `json:"post_id"`
	RootId           string `json:"root_id"`
	ChannelName      string `json:"channel_name"`
	Type             string `json:"type"`
	SenderId         string `json:"sender_id"`
	OverrideUsername string `json:"override_username"`
	OverrideIconUrl  string `json:"override_icon_url"`
	FromWebhook      string `json:"from_webhook"`
	Version          string `json:"version"`
	PromoId          string `json:"promo_id"`
}

func (me *PushNotification) ToJson() string {
	b, _ := json.Marshal(me)
	return string(b)
}

func (me *PushNotification) SetDeviceIdAndPlatform(deviceId string) {

	/*index := strings.Index(deviceId, ":")

	if index > -1 {
		me.Platform = deviceId[:index]
		me.DeviceId = deviceId[index+1:]
	} else {
		me.DeviceId = deviceId
	}*/
	me.DeviceId = deviceId
}

func PushNotificationFromJson(data io.Reader) *PushNotification {
	var me *PushNotification
	json.NewDecoder(data).Decode(&me)
	return me
}

type SmsNotification struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

func (me *SmsNotification) ToJson() string {
	b, _ := json.Marshal(me)
	return string(b)
}
