package app

import (
	"im/model"
	smsaero "im/utils"
	"fmt"
	"im/mlog"
)

func (a *App) SendVerifySms(phone, locale,  msg string) *model.AppError {

	message := model.SmsNotification{
		Phone:   phone,
		Message: msg,
	}

	a.Srv.Go(func() {
		a.sendToSmsProxy(message)
	})

	return nil
}




func (a *App) sendToSmsProxy(msg model.SmsNotification) {

	client := smsaero.NewClient("ivan@russianit.ru", "mmRjD5mOoMkvVsuIAMiVwX6i9czQ", "", "")
	msgAero := smsaero.MessageRequest{
		Numbers: []string{msg.Phone},
		Sign: "RSIT",
		Text: msg.Message,
		Channel: smsaero.ChannelDirect,
	}
	resp, err := client.Send(msgAero)
	if err != nil {
		mlog.Error(err.Error())
	}

	fmt.Printf("%#v\n", resp)

	return
}
