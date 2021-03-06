package model

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

const (
	// ожидание оплаты (ждем оплату)
	ORDER_STATUS_AWAITING_PAYMENT string = "awaitingPayment"

	// оплаченный заказ, ждем результатов (ждем отправки)
	ORDER_STATUS_AWAITING_FULFILLMENT string = "awaitingFulfillment"

	// заказ упакован, ждем отправки (сформирован)
	ORDER_STATUS_AWAITING_PICKUP string = "awaitingPickup"

	// заказ отправлен
	ORDER_STATUS_AWAITING_SHIPMENT string = "awaitingShipment"

	// заказ доставлен
	ORDER_STATUS_SHIPPED string = "shipped"

	// заказ отменен (отклонен)
	ORDER_STATUS_DECLINED string = "declined"

	// возврат
	ORDER_STATUS_REFUNDED string = "refunded"

	// текущие заказы
	ORDER_STADY_CURRENT string = "current"
	// отложенные заказы
	ORDER_STADY_DEFERRED string = "deferred"
	// закрытые заказы
	ORDER_STADY_CLOSED string = "closed"

	PAYMENT_SYSTEM_CASH     string = "cash"
	PAYMENT_SYSTEM_ALFABANK string = "alfabank"
	PAYMENT_SYSTEM_SBERBANK string = "sberbank"
)

type Order struct {
	Id                   string    `json:"id"`
	Payed                bool      `json:"payed"`
	PayedAt              int64     `json:"payed_at"`
	Canceled             bool      `json:"canceled"`
	CanceledAt           int64     `json:"canceled_at"`
	ReasonCanceled       string    `json:"reason_canceled"`
	Status               string    `json:"status"`
	StatusAt             int64     `json:"status_at"`
	PriceDelivery        float64   `json:"price_delivery"`
	DeliveryAt           int64     `json:"delivery_at"`
	Price                float64   `json:"price"`
	Currency             string    `json:"currency"`
	DiscountValue        float64   `json:"discount_value"`
	UserId               string    `json:"user_id"`
	PaySystemId          string    `json:"pay_system_id"`
	DeliveryId           string    `json:"delivery_id"`
	PaySystemStatus      string    `json:"pay_systems_status"`
	PaySystemCode        string    `json:"pay_system_code"`
	PaySystemDescription string    `json:"pay_system_description"`
	PaySystemMessage     string    `json:"pay_system_message"`
	PaySystemSum         float64   `json:"pay_system_sum"`
	PaySystemCurrency    string    `json:"pay_system_currency"`
	PaySystemResponseAt  int64     `json:"pay_system_response_at"`
	PaySystemOrderNum    string    `json:"pay_system_order_num"`
	CreateAt             int64     `json:"create_at"`
	UpdateAt             int64     `json:"update_at"`
	DeleteAt             int64     `json:"delete_at"`
	Address              string    `json:"address"`
	Comment              string    `json:"comment"`
	Phone                string    `json:"phone"`
	Processing           bool      `json:"processing"`
	Positions            []*Basket `db:"-" json:"positions"`
	Post                 *Post     `db:"-" json:"post,omitempty"`
	User                 *User     `db:"-" json:"user,omitempty"`
}

type OrderPatch struct {
	Status            *string `json:"status"`
	DeliveryAt        *int64  `json:"delivery_at"`
	PaySystemCode     *string `json:"pay_system_code"`
	Processing        *bool   `json:"processing"`
	PaySystemId       *string `json:"pay_system_id"`
	PaySystemOrderNum *string `json:"pay_system_order_num"`
}

func (o *Order) Patch(patch *OrderPatch) {
	if patch.Status != nil {
		o.Status = *patch.Status
	}
	if patch.DeliveryAt != nil {
		o.DeliveryAt = *patch.DeliveryAt
	}
	if patch.PaySystemCode != nil {
		o.PaySystemCode = *patch.PaySystemCode
	}
	if patch.Processing != nil {
		o.Processing = *patch.Processing
	}
	if patch.PaySystemId != nil {
		o.PaySystemId = *patch.PaySystemId
	}
	if patch.PaySystemOrderNum != nil {
		o.PaySystemOrderNum = *patch.PaySystemOrderNum
	}
}

func (order *Order) ToJson() string {
	b, _ := json.Marshal(order)
	return string(b)
}

func OrderFromJson(data io.Reader) *Order {
	var order *Order
	json.NewDecoder(data).Decode(&order)
	return order
}

func OrderPatchFromJson(data io.Reader) *OrderPatch {
	var patch *OrderPatch
	json.NewDecoder(data).Decode(&patch)
	return patch
}

func (o *Order) Clone() *Order {
	copy := *o
	return &copy
}
func (o *Order) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.Status = ORDER_STATUS_AWAITING_PAYMENT

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Order) PreCommit() {

}

func (o *Order) MakeNonNil() {

}

func (o *Order) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Order.IsValid", "model.order.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Order.IsValid", "model.order.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Order.IsValid", "model.order.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.UserId) != 26 {
		return NewAppError("Order.IsValid", "model.order.is_valid.user_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Phone) == 0 {
		return NewAppError("Order.IsValid", "model.order.is_valid.phone.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.Price <= 0 {
		return NewAppError("Order.IsValid", "model.order.is_valid.price.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func (o *Order) NormalizePositions() {
	positions := make(map[string]*Basket)
	var list []*Basket
	for _, entry := range o.Positions {
		if _, value := positions[entry.ProductId]; !value {
			positions[entry.ProductId] = entry
		} else {
			positions[entry.ProductId].Quantity += entry.Quantity
		}
	}
	for _, position := range positions {
		if position.Quantity >= 1 {
			list = append(list, position)
		}
	}
	o.Positions = list
}

func (o *Order) FormatOrderNumber() string {
	create_at := strconv.FormatInt(o.CreateAt, 10)
	if i := len(create_at); i == 0 {
		create_at = strconv.FormatInt(GetMillis(), 10)
	}
	num := create_at[len(create_at)-6:]
	return num[:3] + "-" + num[3:]
}
