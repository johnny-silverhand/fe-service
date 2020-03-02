package app

import (
	"fmt"
	"im/mlog"
	"im/model"
	smsaero "im/utils"
)

func (a *App) SendVerifySms(phone, locale, msg string) *model.AppError {

	message := model.SmsNotification{
		Phone:   phone,
		Message: msg,
	}

	a.Srv.Go(func() {
		print(message.Message)
		fmt.Println(message)
		//a.sendToSmsProxy(message)
		if phone == `79991892951` {
			a.sendToSmsProxy(message)
		}
	})

	return nil
}

func (a *App) sendToSmsProxy(msg model.SmsNotification) {

	//client := smsaero.NewClient("ivan@russianit.ru", "mmRjD5mOoMkvVsuIAMiVwX6i9czQ", "", "")
	//client := smsaero.NewClient("ndmitry.web@gmail.com", "y2VplrTeh2wE9eKy6A11NWmy3FA1", "", "")
	client := smsaero.NewClient("osmary@bk.ru", "QxXk6d6055rG7bgttUUeQEtHrsmn", "", "")
	msgAero := smsaero.MessageRequest{
		Numbers: []string{msg.Phone},
		Sign:    "SMS Aero",
		Text:    msg.Message,
		Channel: smsaero.ChannelDirect,
	}
	resp, err := client.Send(msgAero)
	if err != nil {
		mlog.Error(err.Error())
	}

	fmt.Printf("%#v\n", resp)

	return
}
