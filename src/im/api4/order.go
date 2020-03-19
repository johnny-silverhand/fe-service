package api4

import (
	"im/model"
	"im/services/payment"
	"im/services/payment/alfabank"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/currency"
	"net/http"
)

func (api *API) InitOrder() {

	api.BaseRoutes.Orders.Handle("", api.ApiSessionRequired(getAllOrders)).Methods("GET")
	api.BaseRoutes.Orders.Handle("/stats", api.ApiSessionRequired(getOrdersStats)).Methods("GET")
	api.BaseRoutes.Orders.Handle("/invoice", api.ApiSessionRequired(createInvoice)).Methods("POST")
	api.BaseRoutes.Orders.Handle("", api.ApiHandler(createOrder)).Methods("POST")

	api.BaseRoutes.Orders.Handle("/{order_id:[A-Za-z0-9]+}", api.ApiHandler(getOrder)).Methods("GET")
	api.BaseRoutes.Order.Handle("/cancel", api.ApiHandler(cancelOrder)).Methods("GET")
	api.BaseRoutes.Order.Handle("/prepayment", api.ApiHandler(getPaymentOrderUrl)).Methods("GET")
	api.BaseRoutes.Order.Handle("/status", api.ApiHandler(getPaymentOrderStatus)).Methods("GET")
	api.BaseRoutes.Order.Handle("", api.ApiHandler(updateOrder)).Methods("PUT")
	api.BaseRoutes.Order.Handle("", api.ApiHandler(deleteOrder)).Methods("DELETE")
	api.BaseRoutes.User.Handle("/orders", api.ApiSessionRequired(getUserOrders)).Methods("GET")

}

func createInvoice(c *Context, w http.ResponseWriter, r *http.Request) {
	//var user *model.User
	var err *model.AppError
	order := model.OrderFromJson(r.Body)

	user, err := c.App.GetUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}
	application, err := c.App.GetApplication(user.AppId)
	if err != nil {
		c.Err = err
		return
	}
	if len(application.AqType) <= 0 && len(application.AqUsername) <= 0 && len(application.AqPassword) <= 0 {
		c.SetInvalidParam("cash")
		return
	}

	if order == nil {
		c.SetInvalidParam("order")
		return
	}

	orderUser, err := c.App.GetUser(order.UserId)
	if err != nil {
		c.Err = err
		return
	}

	order.Phone = orderUser.Phone

	result, err := c.App.CreateOrderInvoice(order, user)

	if err != nil {
		c.Err = err
		return
	}

	/*if list, err := c.App.GetAllLevelsPage(0, 60, &user.AppId); err == nil {
		list.SortByLvl()

		if u, e := c.App.GetUser(user.InvitedBy); e == nil {
			for _, id := range list.Order {
				accural := math.Floor(result.Price * (list.Levels[id].Value / 100))

				transaction := &model.Transaction{
					UserId:      u.Id,
					OrderId:     result.Id,
					Description: fmt.Sprintf("Начисление по заказу № %s \n", result.FormatOrderNumber()),
					Value:       accural,
					Type:        model.TRANSACTION_TYPE_BONUS,
				}

				if transaction.Value > 0 {
					c.App.AccrualTransaction(transaction)
				}

				if u, e = c.App.GetUser(u.InvitedBy); e != nil {
					break
				}
			}
		}

	}*/

	w.Write([]byte(result.ToJson()))
}

func getOrdersStats(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err != nil {
		return
	}

	user, err := c.App.GetUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}

	stats, err := c.App.GetOrdersStats(model.OrderCountOptions{
		AppId: user.AppId,
	})

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(stats.ToJson()))
}

func getAllOrders(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	typeOrder := r.URL.Query().Get("type")
	/*afterOrder := r.URL.Query().Get("after")
	beforeOrder := r.URL.Query().Get("before")
	sinceString := r.URL.Query().Get("since")*/
	sort := r.URL.Query().Get("sort")

	/*var since int64
	var parseError error

	if len(sinceString) > 0 {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}*/

	/*	if !c.App.SessionHasPermissionToChannel(c.Session, c.Params.ChannelId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}*/

	var list *model.OrderList
	var err *model.AppError
	//etag := ""

	user, err := c.App.GetUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}

	orderGetOptions := &model.OrderGetOptions{
		Sort:    sort,
		Page:    c.Params.Page,
		PerPage: c.Params.PerPage,
		AppId:   user.AppId,
	}

	switch typeOrder {
	case model.ORDER_STADY_CURRENT:
		orderGetOptions.Status = model.ORDER_STADY_CURRENT
	case model.ORDER_STADY_DEFERRED:
		orderGetOptions.Status = model.ORDER_STADY_DEFERRED
	case model.ORDER_STADY_CLOSED:
		orderGetOptions.Status = model.ORDER_STADY_CLOSED
	default:
		orderGetOptions.Status = model.ORDER_STADY_CURRENT
	}

	/*if since > 0 {
		list, err = c.App.GetAllOrdersSince(since, orderGetOptions)
	} else if len(afterOrder) > 0 {

		list, err = c.App.GetAllOrdersAfterOrder(afterOrder, orderGetOptions)
	} else if len(beforeOrder) > 0 {

		list, err = c.App.GetAllOrdersBeforeOrder(beforeOrder, orderGetOptions)
	} else {*/
	list, err = c.App.GetAllOrdersPage(orderGetOptions)
	/*}*/

	if err != nil {
		c.Err = err
		return
	}

	/*	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}*/

	w.Write([]byte(c.App.PrepareOrderListForClient(list).ToJson()))
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

	var appId string
	if user, err := c.App.GetUser(order.UserId); err != nil {
		c.Err = err
		return
	} else {
		appId = user.AppId
	}

	var application *model.Application
	if app, err := c.App.GetApplication(appId); err != nil {
		c.Err = err
		return
	} else {
		application = app
	}

	var siteURL string
	siteURL = *c.App.Config().ServiceSettings.SiteURL

	if application.AqType == model.SBERBANK_AQUIRING_TYPE { //foodexp-api	foodexp
		var sber payment.SberBankBackend
		if response, err := sber.RegisterOrder(order, sberbank.ClientConfig{
			UserName:           application.AqUsername,
			Password:           application.AqPassword,
			Currency:           currency.RUB,
			Language:           "ru",
			SessionTimeoutSecs: 1200,
			SandboxMode:        true,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemCode: model.NewString(response.OrderId)}, false)
			})

			w.Write([]byte(response.ToJson()))
		}
	} else if application.AqType == model.ALFABANK_AQUIRING_TYPE { // yktours-api	yktours*?1
		var alfa payment.AlfaBankBackend
		if response, err := alfa.RegisterOrder(order, alfabank.ClientConfig{
			UserName:           application.AqUsername,
			Password:           application.AqPassword,
			Currency:           currency.RUB,
			Language:           "ru",
			SessionTimeoutSecs: 1200,
			SandboxMode:        true,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemCode: model.NewString(response.OrderId)}, false)
			})

			w.Write([]byte(response.ToJson()))
		}
	} else {
		c.Err = model.NewAppError("GetPaymentOrderUrl", "api.order.get_payment_order_url.app_error", nil, "", http.StatusBadRequest)
		return
	}

	/*if response, err := registerOrder(order); err != nil {
		c.Err = err
		return
	} else {
		c.App.Srv.Go(func() {
			order.PaySystemCode = *response.OrderId
			c.App.UpdateOrder(order, false)
		})

		/*c.App.Srv.Go(func() {
			c.App.SetOrderPayed(c.Params.OrderId)
		})

		w.Write([]byte(response.ToJson()))
	}*/

}

func updateOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	patch := model.OrderPatchFromJson(r.Body)

	if patch == nil {
		c.SetInvalidParam("order")
		return
	}

	rorder, err := c.App.UpdateOrder(c.Params.OrderId, patch, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rorder.ToJson()))
}

func createOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	var user *model.User
	var err *model.AppError
	order := model.OrderFromJson(r.Body)

	if order == nil {
		c.SetInvalidParam("order")
		return
	}

	if len(c.App.Session.UserId) == 0 {

		if len(order.Phone) > 0 {

			user, err = c.App.GetUserApplicationByPhone(order.Phone, c.Params.AppId)

			if err != nil {

				newUser := &model.User{
					Phone:         order.Phone,
					Username:      order.Phone,
					AppId:         c.Params.AppId,
					EmailVerified: true,
				}

				user, _ = c.App.AutoCreateUser(newUser)

			} else {
				order.UserId = user.Id
			}
		} else {
			c.SetInvalidParam("phone")
			return
		}

	} else {
		order.UserId = c.App.Session.UserId
		user, _ = c.App.GetUser(order.UserId)
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

	/**/

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
	/*c.RequireUserId()
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

	w.Write([]byte(list.ToJson()))*/

	c.RequireUserId()
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	var list *model.OrderList
	var err *model.AppError
	//etag := ""
	sort := r.URL.Query().Get("sort")
	orderGetOptions := &model.OrderGetOptions{
		Sort:    sort,
		Page:    c.Params.Page,
		PerPage: c.Params.PerPage,
		AppId:   c.Params.AppId,
		UserId:  c.Params.UserId,
	}

	list, err = c.App.GetUserOrders(orderGetOptions)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}

func getPaymentOrderStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()
	if c.Err != nil {
		return
	}

	order, err := c.App.GetOrder(c.Params.OrderId)

	if err != nil {
		c.Err = err
		return
	}

	var appId string
	if user, err := c.App.GetUser(order.UserId); err != nil {
		c.Err = err
		return
	} else {
		appId = user.AppId
	}

	var application *model.Application
	if app, err := c.App.GetApplication(appId); err != nil {
		c.Err = err
		return
	} else {
		application = app
	}

	var siteURL string
	siteURL = *c.App.Config().ServiceSettings.SiteURL

	if application.AqType == model.SBERBANK_AQUIRING_TYPE { // foodexp-api	foodexp
		var sber payment.SberBankBackend
		if response, err := sber.GetOrderStatus(order, sberbank.ClientConfig{
			UserName:           application.AqUsername,
			Password:           application.AqPassword,
			Currency:           currency.RUB,
			Language:           "ru",
			SessionTimeoutSecs: 1200,
			SandboxMode:        true,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				//c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemCode: model.NewString(response.)}, false)
			})

			w.Write([]byte(response.ToJson()))
		}
	} else if application.AqType == model.ALFABANK_AQUIRING_TYPE { // yktours-api	yktours*?1
		var alfa payment.AlfaBankBackend
		if response, err := alfa.GetOrderStatus(order, alfabank.ClientConfig{
			UserName:           application.AqUsername,
			Password:           application.AqPassword,
			Currency:           currency.RUB,
			Language:           "ru",
			SessionTimeoutSecs: 1200,
			SandboxMode:        true,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				//c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemCode: model.NewString(response.)}, false)
			})

			w.Write([]byte(response.ToJson()))
		}
	} else {
		c.Err = model.NewAppError("GetPaymentOrderStatus", "api.order.get_payment_order_status.app_error", nil, "", http.StatusBadRequest)
		return
	}

	/*var sber payment.SberBankBackend
	if response, err := sber.GetOrderStatus(order); err != nil {
		c.Err = err
		return
	} else {

		c.App.Srv.Go(func() {
			c.App.SetOrderPayed(c.Params.OrderId, response)
		})
	}

	if c.Err != nil {
		return
	}
	*/
	ReturnStatusOK(w)
}

func cancelOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()

	c.App.Srv.Go(func() {
		c.App.SetOrderCancel(c.Params.OrderId)
	})

	if c.Err != nil {
		return
	}

	ReturnStatusOK(w)
}
