package payment

import (
	"context"
	"im/model"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/schema"
	"net/http"
	"strconv"
)

const SBERBANK_ORDER_STATUS_PAYED = 2
const SBERBANK_REFUND_ORDER_STATUS_OK = "0"
const SBERBANK_REVERSE_ORDER_STATUS_OK = "0"

type SberBankBackend struct {
}

func (b *SberBankBackend) sbNew(config sberbank.ClientConfig) (*sberbank.Client, error) {
	cfg := config

	client, err := sberbank.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (b *SberBankBackend) TestConnection(config sberbank.ClientConfig) *model.AppError {
	if _, err := b.sbNew(config); err != nil {
		return model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (b *SberBankBackend) RegisterOrder(order *model.Order, config sberbank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	/*sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}*/

	var client *sberbank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	amount := order.Price * 100
	sbOrder := sberbank.Order{
		OrderNumber: strconv.FormatInt(model.GetMillis(), 10),
		Amount:      int(amount),
		Description: "",
		ReturnURL:   config.SiteURL + "/api/v4/orders/" + order.Id + "/status",
	}

	if result, _, err := client.RegisterOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}

func (b *SberBankBackend) GetOrderStatus(order *model.Order, config sberbank.ClientConfig) (response *schema.OrderStatusResponse, err *model.AppError) {

	var client *sberbank.Client

	if c, err := b.sbNew(config); err != nil {
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

func (b *SberBankBackend) GetRefundOrderResponse(order *model.Order, config sberbank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	var client *sberbank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("services.payment.sberbank", "get_refund_order_response", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	amount := order.Price * 100
	sbOrder := sberbank.Order{
		OrderNumber: order.PaySystemOrderNum,
		Amount:      int(amount),
	}

	if result, _, err := client.RefundOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}

func (b *SberBankBackend) GetReverseOrderResponse(order *model.Order, config sberbank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	var client *sberbank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("services.payment.sberbank", "get_refund_order_response", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	sbOrder := sberbank.Order{
		OrderNumber: order.PaySystemOrderNum,
	}

	if result, _, err := client.ReverseOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}
