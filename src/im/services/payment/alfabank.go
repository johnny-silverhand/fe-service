package payment

import (
	"context"
	"im/model"
	"im/services/payment/alfabank"
	"im/services/payment/alfabank/schema"
	"net/http"
	"strconv"
)

/*0	Заказ зарегистрирован, но не оплачен
1	Предавторизованная сумма захолдирована (для двухстадийных платежей)
2	Проведена полная авторизация суммы заказа
3	Авторизация отменена
4	По транзакции была проведена операция возврата
5	Инициирована авторизация через ACS банка-эмитента
6	Авторизация отклонена*/
//const ALFABANK_ORDER_STATUS_REGISTER = 0

const ALFABANK_ORDER_STATUS_PAYED = 2
const ALFABANK_REFUND_ORDER_STATUS_OK = "0"
const ALFABANK_REVERSE_ORDER_STATUS_OK = "0"

type AlfaBankBackend struct {
}

func (b *AlfaBankBackend) sbNew(config alfabank.ClientConfig) (*alfabank.Client, error) {
	cfg := config

	client, err := alfabank.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (b *AlfaBankBackend) TestConnection(config alfabank.ClientConfig) *model.AppError {
	if _, err := b.sbNew(config); err != nil {
		return model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (b *AlfaBankBackend) RegisterOrder(order *model.Order, config alfabank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	/*sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}*/

	var client *alfabank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	amount := order.Price * 100
	sbOrder := alfabank.Order{
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

func (b *AlfaBankBackend) GetOrderStatus(order *model.Order, config alfabank.ClientConfig) (response *schema.OrderStatusResponse, err *model.AppError) {

	var client *alfabank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("services.payment.alfabank", "get_order_status", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	sbOrder := alfabank.Order{
		OrderNumber: order.PaySystemCode,
	}

	if result, _, err := client.GetOrderStatus(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}

func (b *AlfaBankBackend) GetRefundOrderResponse(order *model.Order, config alfabank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	var client *alfabank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("services.payment.sberbank", "get_refund_order_response", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}
	amount := order.Price * 100
	sbOrder := alfabank.Order{
		OrderNumber: order.PaySystemOrderNum,
		Amount:      int(amount),
	}

	if result, _, err := client.RefundOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}

func (b *AlfaBankBackend) GetReverseOrderResponse(order *model.Order, config alfabank.ClientConfig) (response *schema.OrderResponse, err *model.AppError) {

	var client *alfabank.Client

	if c, err := b.sbNew(config); err != nil {
		return nil, model.NewAppError("services.payment.sberbank", "get_refund_order_response", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}
	sbOrder := alfabank.Order{
		OrderNumber: order.PaySystemOrderNum,
	}

	if result, _, err := client.ReverseOrder(context.Background(), sbOrder); err != nil {
		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return result, nil
	}

}
