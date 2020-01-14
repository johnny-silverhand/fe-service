package payment

import (
	"context"
	"im/model"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/currency"
	"im/services/payment/sberbank/schema"
	"net/http"
	"strconv"
)

type SberBankBackend struct {
}

func (b *SberBankBackend) sbNew() (*sberbank.Client, error) {
	cfg := sberbank.ClientConfig{
		UserName:           "foodexp-api", // Replace with your own
		Currency:           currency.RUB,
		Password:           "foodexp", // Replace with your own
		Language:           "ru",
		SessionTimeoutSecs: 1200,
		SandboxMode:        true,
	}

	client, err := sberbank.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (b *SberBankBackend) TestConnection() *model.AppError {

	return nil
}

func (b *SberBankBackend) RegisterOrder(application *model.Application, order *model.Order) (response *schema.OrderResponse, err *model.AppError) {

	/*sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}*/

	var client *sberbank.Client

	if c, err := b.sbNew(); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	amount := order.Price * 100
	sbOrder := sberbank.Order{
		OrderNumber: strconv.FormatInt(model.GetMillis(), 10),
		Amount:      int(amount),
		Description: "",
		ReturnURL:   "http://foodexpress2.russianit.ru/api/v4/orders/" + order.Id + "/status",
	}

	if result, _, err := client.RegisterOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}

func (b *SberBankBackend) GetOrderStatus(application *model.Application, order *model.Order) (response *schema.OrderStatusResponse, err *model.AppError) {

	/*sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}*/

	var client *sberbank.Client

	if c, err := b.sbNew(); err != nil {
		return nil, model.NewAppError("services.payment.sberbank", "get_order_status", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	sbOrder := sberbank.Order{
		OrderNumber: order.PaySystemCode,
	}

	if result, _, err := client.GetOrderStatus(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}
