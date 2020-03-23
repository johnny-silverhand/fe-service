package api4

import (
	"im/mlog"
	"im/model"
	"im/services/payment"
	"im/services/payment/alfabank"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/currency"
	"im/utils"
	"net/http"
	"strconv"
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
	if c.Err != nil {
		return
	}

	typeOrder := r.URL.Query().Get("type")
	sort := r.URL.Query().Get("sort")

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

	list, err = c.App.GetAllOrdersPage(orderGetOptions)

	if err != nil {
		c.Err = err
		return
	}

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
	sandboxMode := *c.App.Config().ServiceSettings.EnableDeveloper

	if application.AqType == model.SBERBANK_AQUIRING_TYPE { //foodexp-api	foodexp
		var sber payment.SberBankBackend
		if response, err := sber.RegisterOrder(order, sberbank.ClientConfig{
			UserName:           application.AqUsername,
			Password:           application.AqPassword,
			Currency:           currency.RUB,
			Language:           "ru",
			SessionTimeoutSecs: 1200,
			SandboxMode:        sandboxMode,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemId: model.NewString(model.SBERBANK_AQUIRING_TYPE), PaySystemCode: model.NewString(response.OrderId), PaySystemOrderNum: model.NewString(strconv.FormatInt(model.GetMillis(), 10))}, false)
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
			SandboxMode:        sandboxMode,
			SiteURL:            siteURL,
		}); err != nil {
			c.Err = err
			return
		} else {
			c.App.Srv.Go(func() {
				c.App.UpdateOrder(order.Id, &model.OrderPatch{PaySystemId: model.NewString(model.ALFABANK_AQUIRING_TYPE), PaySystemCode: model.NewString(response.OrderId), PaySystemOrderNum: model.NewString(strconv.FormatInt(model.GetMillis(), 10))}, false)
			})

			w.Write([]byte(response.ToJson()))
		}
	} else {
		c.Err = model.NewAppError("GetPaymentOrderUrl", "api.order.get_payment_order_url.app_error", nil, "", http.StatusBadRequest)
		return
	}

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

	if _, err := c.App.DeleteOrder(c.Params.OrderId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func getUserOrders(c *Context, w http.ResponseWriter, r *http.Request) {
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
	c.App.Srv.Go(func() {
		order, err := c.App.GetOrder(c.Params.OrderId)

		if err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		}

		var appId string
		if user, err := c.App.GetUser(order.UserId); err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		} else {
			appId = user.AppId
		}

		var application *model.Application
		if app, err := c.App.GetApplication(appId); err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		} else {
			application = app
		}

		var msg string
		var siteURL string
		siteURL = *c.App.Config().ServiceSettings.SiteURL
		sandboxMode := *c.App.Config().ServiceSettings.EnableDeveloper

		if order.PaySystemId == model.SBERBANK_AQUIRING_TYPE { // foodexp-api	foodexp
			var sber payment.SberBankBackend
			if response, err := sber.GetOrderStatus(order, sberbank.ClientConfig{
				UserName:           application.AqUsername,
				Password:           application.AqPassword,
				Currency:           currency.RUB,
				Language:           "ru",
				SessionTimeoutSecs: 1200,
				SandboxMode:        sandboxMode,
				SiteURL:            siteURL,
			}); err != nil {
				mlog.Warn(err.Error())
			} else {
				if response.OrderStatus == payment.SBERBANK_ORDER_STATUS_PAYED {
					c.App.UpdateOrder(order.Id, &model.OrderPatch{Status: model.NewString(model.ORDER_STATUS_AWAITING_FULFILLMENT)}, false)

					msg = "Оплата банковской картой "
					msg += response.CardAuthInfo.MaskedPan
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
				} else {
					msg = "Оплата банковской картой не произведена"
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
				}
			}
		} else if order.PaySystemId == model.ALFABANK_AQUIRING_TYPE { // yktours-api	yktours*?1
			var alfa payment.AlfaBankBackend
			if response, err := alfa.GetOrderStatus(order, alfabank.ClientConfig{
				UserName:           application.AqUsername,
				Password:           application.AqPassword,
				Currency:           currency.RUB,
				Language:           "ru",
				SessionTimeoutSecs: 1200,
				SandboxMode:        sandboxMode,
				SiteURL:            siteURL,
			}); err != nil {
				mlog.Warn(err.Error())
			} else {
				if response.OrderStatus == payment.ALFABANK_ORDER_STATUS_PAYED {
					c.App.UpdateOrder(order.Id, &model.OrderPatch{Status: model.NewString(model.ORDER_STATUS_AWAITING_FULFILLMENT)}, false)

					msg = "Оплата банковской картой "
					msg += response.CardAuthInfo.MaskedPan
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)

				} else {
					msg = "Оплата банковской картой не произведена"
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
				}
			}
		}
	})

	ReturnStatusOK(w)
}

func cancelOrder(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOrderId()

	if c.Err != nil {
		return
	}

	c.App.Srv.Go(func() {
		order, err := c.App.GetOrder(c.Params.OrderId)

		if err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		}
		var msg string
		if order.PaySystemId == model.PAYMENT_SYSTEM_CASH || !utils.StringInSlice(order.PaySystemId, []string{model.PAYMENT_SYSTEM_ALFABANK, model.PAYMENT_SYSTEM_SBERBANK}) {
			if err := c.App.SetOrderCancel(c.Params.OrderId); err != nil {
				c.Err = err
				mlog.Warn(err.Error())
				return
			}
			msg = "Заказ № " + order.FormatOrderNumber() + " отменен."

			post := &model.Post{
				UserId:   order.UserId,
				Message:  msg,
				CreateAt: model.GetMillis() + 1,
				Type:     model.POST_WITH_TRANSACTION,
			}

			c.App.CreatePostWithTransaction(post, false)
			return
		} /* else if !utils.StringInSlice(order.PaySystemId, []string{model.PAYMENT_SYSTEM_ALFABANK, model.PAYMENT_SYSTEM_SBERBANK}) {
			c.Err = model.NewAppError("CancelOrder", "api.order.cancel_order.app_error", nil, "", http.StatusBadRequest)
			mlog.Warn(c.Err.Error())
			return
		}*/

		var appId string
		if user, err := c.App.GetUser(order.UserId); err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		} else {
			appId = user.AppId
		}

		var application *model.Application
		if app, err := c.App.GetApplication(appId); err != nil {
			c.Err = err
			mlog.Warn(err.Error())
			return
		} else {
			application = app
		}

		var siteURL string
		siteURL = *c.App.Config().ServiceSettings.SiteURL
		sandboxMode := *c.App.Config().ServiceSettings.EnableDeveloper

		if order.PaySystemId == model.SBERBANK_AQUIRING_TYPE { // foodexp-api	foodexp
			var sber payment.SberBankBackend
			if response, err := sber.GetReverseOrderResponse(order, sberbank.ClientConfig{
				UserName:           application.AqUsername,
				Password:           application.AqPassword,
				Currency:           currency.RUB,
				Language:           "ru",
				SessionTimeoutSecs: 1200,
				SandboxMode:        sandboxMode,
				SiteURL:            siteURL,
			}); err != nil {
				c.Err = err
				mlog.Warn(err.Error())
				return
			} else {
				if response.ErrorCode == payment.SBERBANK_REVERSE_ORDER_STATUS_OK {
					c.App.UpdateOrder(order.Id, &model.OrderPatch{Status: model.NewString(model.ORDER_STATUS_DECLINED)}, false)

					msg = "Оплата банковской картой по транзакции: "
					msg += response.OrderId + " отменена"
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
				} else {
					mlog.Warn(response.ToJson())

					if err := c.App.SetOrderCancel(c.Params.OrderId); err != nil {
						c.Err = err
						mlog.Warn(err.Error())
						return
					}
					msg = "Заказ № " + order.FormatOrderNumber() + " отменен."

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
					return
				}
			}
		} else if order.PaySystemId == model.ALFABANK_AQUIRING_TYPE { // yktours-api	yktours*?1
			var alfa payment.AlfaBankBackend
			if response, err := alfa.GetReverseOrderResponse(order, alfabank.ClientConfig{
				UserName:           application.AqUsername,
				Password:           application.AqPassword,
				Currency:           currency.RUB,
				Language:           "ru",
				SessionTimeoutSecs: 1200,
				SandboxMode:        sandboxMode,
				SiteURL:            siteURL,
			}); err != nil {
				c.Err = err
				mlog.Warn(err.Error())
				return
			} else {
				if response.ErrorCode == payment.ALFABANK_REVERSE_ORDER_STATUS_OK {
					c.App.UpdateOrder(order.Id, &model.OrderPatch{Status: model.NewString(model.ORDER_STATUS_DECLINED)}, false)

					msg = "Оплата банковской картой по транзакции: "
					msg += response.OrderId + " отменена"
					msg += ". № заказа " + order.FormatOrderNumber()

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
				} else {
					mlog.Warn(response.ToJson())

					if err := c.App.SetOrderCancel(c.Params.OrderId); err != nil {
						c.Err = err
						mlog.Warn(err.Error())
						return
					}
					msg = "Заказ № " + order.FormatOrderNumber() + " отменен."

					post := &model.Post{
						UserId:   order.UserId,
						Message:  msg,
						CreateAt: model.GetMillis() + 1,
						Type:     model.POST_WITH_TRANSACTION,
					}

					c.App.CreatePostWithTransaction(post, false)
					return
				}
			}
		} else {
			c.Err = model.NewAppError("CancelOrder", "api.order.cancel_order.app_error", nil, "", http.StatusBadRequest)
			mlog.Warn(c.Err.Error())
			return
		}
	})

	ReturnStatusOK(w)
}
