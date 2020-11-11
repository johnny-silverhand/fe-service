package api4

import (
	"im/model"
	"im/utils"
	"net/http"
	"regexp"
)

func (api *API) InitNotification() {

	api.BaseRoutes.Notifications.Handle("/verify/send", api.ApiHandler(sendPushVerifyToken)).Methods("POST")

}

func sendPushVerifyToken(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)
	phone := props["phone"]
	appId := props["app_id"]

	if len(appId) == 0 {
		c.RequireAppId()
		if c.Err != nil {
			return
		}
		appId = c.Params.AppId
	}

	reg, _ := regexp.Compile("[^0-9]+")
	phone = reg.ReplaceAllString(phone, "")
	//user, err := c.App.GetUserByPhone(phone)
	user, err := c.App.GetUserApplicationByPhone(phone, appId)

	if err != nil {
		c.Err = err
		return
	}

	//pwd := "1234"                                   //utils.HashDigit(4)
	pwd := utils.HashDigit(4)
	token, err := c.App.CreateStageToken(user, pwd) /*pwd*/

	if err := c.App.SendVerifyFromStageTokenPush(token.Token); err != nil {
		c.Err = err
		return
	}

	ReturnStatusStageTokenOK(w, token.Token)
}
