package api4

import (
	"encoding/json"
	"im/model"
	"im/services/aquiring"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitOrder() {

	api.BaseRoutes.Orders.Handle("", api.ApiHandler(getAllOrders)).Methods("GET")
	api.BaseRoutes.Orders.Handle("", api.ApiHandler(createOrder)).Methods("POST")

	api.BaseRoutes.Order.Handle("", api.ApiHandler(getOrder)).Methods("GET")
	api.BaseRoutes.Order.Handle("/prepayment", api.ApiHandler(getPaymentOrderUrl)).Methods("GET")
	api.BaseRoutes.Order.Handle("/status", api.ApiHandler(getPaymentOrderStatus)).Methods("GET")
	api.BaseRoutes.Order.Handle("", api.ApiHandler(updateOrder)).Methods("PUT")
	api.BaseRoutes.Order.Handle("", api.ApiHandler(deleteOrder)).Methods("DELETE")
	api.BaseRoutes.User.Handle("/orders", api.ApiSessionRequired(getUserOrders)).Methods("GET")

}

func getAllOrders(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterOrder := r.URL.Query().Get("after")
	beforeOrder := r.URL.Query().Get("before")
	sinceString := r.URL.Query().Get("since")

	var since int64
	var parseError error

	if len(sinceString) > 0 {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}

	/*	if !c.App.SessionHasPermissionToChannel(c.Session, c.Params.ChannelId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}*/

	var list *model.OrderList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllOrdersSince(since)
	} else if len(afterOrder) > 0 {

		list, err = c.App.GetAllOrdersAfterOrder(afterOrder, c.Params.Page, c.Params.PerPage)
	} else if len(beforeOrder) > 0 {

		list, err = c.App.GetAllOrdersBeforeOrder(beforeOrder, c.Params.Page, c.Params.PerPage)
	} else {
		list, err = c.App.GetAllOrdersPage(c.Params.Page, c.Params.PerPage)
	}

	if err != nil {
		c.Err = err
		return
	}

	/*	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}*/

	w.Write([]byte(list.ToJson()))
}

func getOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	order, err := c.App.GetOrder(c.Params.OrderId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(order.ToJson()))

}

func getPaymentOrderUrl(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	order, err := c.App.GetOrder(c.Params.OrderId)

	if err != nil {
		c.Err = err
		return
	}
	if response, err := registerOrder(order); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(response.ToJson()))
	}

}

func updateOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	order := model.OrderFromJson(r.Body)

	if order == nil {
		c.SetInvalidParam("order")
		return
	}

	// The order being updated in the payload must be the same one as indicated in the URL.
	if order.Id != c.Params.OrderId {
		c.SetInvalidParam("id")
		return
	}

	order.Id = c.Params.OrderId

	rorder, err := c.App.UpdateOrder(order, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rorder.ToJson()))
}

func createOrder(c *Context, w http.ResponseWriter, r *http.Request) {

	order := model.OrderFromJson(r.Body)

	if order == nil {
		c.SetInvalidParam("order")
		return
	}

	if len(c.App.Session.UserId) == 0 {

		if len(order.Phone) > 0 {

			user, err := c.App.GetUserByPhone(order.Phone)

			if err != nil {

				newUser := &model.User{
					Phone:         order.Phone,
					Username:      order.Phone,
					EmailVerified: true,
				}

				c.App.AutoCreateUser(newUser)

			} else {
				order.UserId = user.Id
			}
		} else {
			c.SetInvalidParam("phone")
			return
		}

	} else {
		order.UserId = c.App.Session.UserId
	}

	/*	if (order.Positions == nil) {
		c.SetInvalidParam("positions")
		return
	}*/

	result, err := c.App.CreateOrder(order)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(result.ToJson()))
}

func deleteOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetOrder(c.Params.OrderId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == order.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, order.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, order.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteOrder(c.Params.OrderId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func getUserOrders(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	var list *model.OrderList
	var err *model.AppError
	//etag := ""

	list, err = c.App.GetUserOrders(c.Params.UserId, c.Params.Page, c.Params.PerPage, "")

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}

func registerOrder(order *model.Order) (*aquiring.ResponseRegistration, *model.AppError) {
	var client *aquiring.Client

	/*if order.PaySystemId == "sberbank" {
		client = aquiring.NewSberClient("foodexp-api", "foodexp")
	} else {*/
	client = aquiring.NewAlfaClient("yktours-api", "yktours*?1")
	//}

	/*var sber payment.SberBankBackend
	sber.RegisterOrder()*/

	var requestRegistration = aquiring.RequestRegistration{
		OrderNumber: strconv.FormatInt(time.Now().UnixNano(), 10),
		Description: "",
		Amount:      "100000", // потому что нужно значение в копейках
		ReturnUrl:   "http://foodexpress2.russianit.ru/api/v4/order/" + order.Id + "/status",
	}

	if r, err := client.PostRequest("/register.do", requestRegistration); err != nil {

		return nil, model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		var responseReg *aquiring.ResponseRegistration
		json.NewDecoder(r.Body).Decode(&responseReg)

		return responseReg, nil
	}
}

func getPaymentOrderStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()

	c.App.SetOrderPayed(c.Params.OrderId)

	if c.Err != nil {
		return
	}

	ReturnStatusOK(w)
}
