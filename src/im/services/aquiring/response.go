package aquiring

import (
	"encoding/json"
	"im/model"
)

type ErrorDetails struct {
	ErrorCode    *string `json:"errorCode"`
	ErrorMessage *string `json:"errorMessage"`
}

type RegistrationData struct {
	OrderId *string `json:"orderId"` //Номер заказа в платёжной системе
	FormUrl *string `json:"formUrl"` //URL платёжной формы, на который надо перенаправить броузер клиента
}

func (reg *RegistrationData) ToJson() string {
	b, _ := json.Marshal(reg)
	return string(b)
}

func (err *ErrorDetails) ToJson() string {
	b, _ := json.Marshal(err)
	return string(b)
}

type ResponseRegistration struct {
	RegistrationData
	ErrorDetails
}

func (res *ResponseRegistration) ToJson() string {
	b, _ := json.Marshal(res)
	return string(b)
}

func (response ResponseRegistration) GetError() *ErrorDetails {
	return &ErrorDetails{
		ErrorCode:    response.ErrorCode,
		ErrorMessage: response.ErrorMessage,
	}
}

func (response ResponseRegistration) GetData() *RegistrationData {
	return &RegistrationData{
		OrderId: response.OrderId,
		FormUrl: response.FormUrl,
	}
}

type ResponseOrderStatus struct {
	OrderStatus int `json:"orderStatus"` //Состояние заказа в платёжной системе

	ErrorCode    int    `json:"errorCode"`    //Код ошибки
	ErrorMessage string `json:"errorMessage"` //Описание ошибки

	OrderNumber    string `json:"orderNumber"`    //Номер (идентификатор) заказа в системе магазина
	Pan            string `json:"pan"`            //Маскированный номер карты, которая использовалась для оплаты
	Expiration     int    `json:"expiration"`     //Срок истечения действия карты в формате YYYYMM
	CardholderName string `json:"cardholderName"` //Имя держателя карты

	Amount       int    `json:"amount"`       //Сумма платежа в копейках (или центах)
	Currency     int    `json:"currency"`     //Код валюты платежа ISO 4217. Если не указан, считается равным 810 (российские рубли)
	ApprovalCode string `json:"approvalCode"` //Код авторизации МПС
	Ip           string `json:"ip"`           //IP адрес пользователя, который оплачивал заказ

}

func (response ResponseOrderStatus) GetOrderStatus() string {
	switch response.OrderStatus {
	case 0, 3, 5, 6:
		return model.ORDER_STATUS_AWAITING_PAYMENT
	case 1, 2:
		return model.ORDER_STATUS_AWAITING_FULFILLMENT
	case 4:
		return model.ORDER_STATUS_REFUNDED
	default:
		return model.ORDER_STATUS_AWAITING_PAYMENT
	}
}
