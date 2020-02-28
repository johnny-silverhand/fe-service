package api4

import (
	"encoding/json"
	"fmt"
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitTransaction() {
	api.BaseRoutes.Transactions.Handle("/validate", api.ApiSessionRequired(validateTransactionForOrderUser)).Methods("POST")

	api.BaseRoutes.Transactions.Handle("/mailing", api.ApiSessionRequired(createMailingTransactions)).Methods("POST")
	api.BaseRoutes.Transactions.Handle("/discard", api.ApiSessionRequired(discardTransactionUser)).Methods("POST")
	api.BaseRoutes.Transactions.Handle("/charge", api.ApiSessionRequired(chargeTransactionUser)).Methods("POST")

	api.BaseRoutes.Transactions.Handle("", api.ApiHandler(getAllTransactions)).Methods("GET")
	api.BaseRoutes.Transactions.Handle("", api.ApiHandler(createTransaction)).Methods("POST")

	api.BaseRoutes.Transactions.Handle("/{transaction_id:[A-Za-z0-9_-]+}", api.ApiHandler(getTransaction)).Methods("GET")
	api.BaseRoutes.Transaction.Handle("", api.ApiHandler(updateTransaction)).Methods("PUT")
	api.BaseRoutes.Transaction.Handle("", api.ApiHandler(deleteTransaction)).Methods("DELETE")
	api.BaseRoutes.User.Handle("/transactions", api.ApiSessionRequired(getUserTransactions)).Methods("GET")
}

func validateTransactionForOrderUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}
	order := model.OrderFromJson(r.Body)
	if order.Price < 0 {
		c.SetInvalidParam("price")
		return
	}
	appId := c.Params.AppId
	if app, err := c.App.GetApplication(appId); err != nil {
		c.Err = err
		return
	} else {

		discount := order.Price * (float64(app.MaxDiscount) / 100)
		cashback := order.Price * (app.Cashback / 100)

		b, _ := json.Marshal(struct {
			DiscountValue int64 `json:"discount_value"`
			Cashback      int64 `json:"cashback"`
		}{
			DiscountValue: int64(discount),
			Cashback:      int64(cashback),
		})
		w.Write([]byte(string(b)))
	}
}

func createMailingTransactions(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_CREATE_POST_PUBLIC) {
		c.SetPermissionError(model.PERMISSION_CREATE_POST_PUBLIC)
		return
	}
	transaction := model.TransactionFromJson(r.Body)
	if transaction == nil {
		c.SetInvalidParam("transaction")
		return
	}
	/*if len(transaction.Description) == 0 {
		c.SetInvalidParam("description")
		return
	}*/
	c.RequireAppId()
	if c.Err != nil {
		return
	}
	appId := c.Params.AppId
	/*if len(appId) == 0 {
		if user, _ := c.App.GetUser(c.App.Session.UserId); user != nil {
			appId = user.AppId
		} else {
			appId = c.App.Session.AppId
		}
	}

	if len(appId) == 0 {
		c.SetInvalidParam("app_id")
		return
	}*/
	c.App.Srv.Go(func() {
		if users, err := c.App.GetUsers(&model.UserGetOptions{
			AppId:   appId,
			Page:    0,
			PerPage: 100000,
			Role:    model.CHANNEL_USER_ROLE_ID,
		}); err != nil {
			c.Err = err
			return
		} else {
			for _, user := range users {
				var ts model.Transaction
				ts.Value = transaction.Value
				ts.Description = "Начисление администратором"
				ts.UserId = user.Id

				if len(ts.UserId) != 26 {
					continue
				}

				if _, err := c.App.AccrualTransaction(&ts); err != nil {
					continue
				}

				var channel *model.Channel
				if channel, _ = c.App.FindOpennedChannel(user.Id); channel != nil {
					c.App.AddChannelMemberIfNeeded(user.Id, channel)
				} else {
					if channel, _ = c.App.CreateUnresolvedChannel(user.Id); channel != nil {
						<-c.App.Srv.Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis())
					}
				}

				if user.NotifyProps[model.PUSH_NOTIFY_PROP] == model.USER_NOTIFY_ALL && channel != nil {
					c.App.SendCustomNotifications(user, channel,
						"Вам начислены дополнительные баллы! Количество начисленных баллов: "+
							fmt.Sprintf("%.0f", ts.Value))
				}
			}
		}
	})

	ReturnStatusOK(w)
}

func discardTransactionUser(c *Context, w http.ResponseWriter, r *http.Request) {
	transaction := model.TransactionFromJson(r.Body)

	if len(transaction.Code) == 0 || len(transaction.Token) == 0 {
		c.SetInvalidParam("code or token")
		return
	}

	if transaction == nil {
		c.SetInvalidParam("transaction")
		return
	}

	if len(transaction.UserId) != 26 {
		c.SetInvalidParam("user_id")
		return
	}

	user, err := c.App.GetUser(transaction.UserId)
	if err != nil {
		c.Err = err
		return
	}

	if user.Balance < transaction.Value {
		c.SetInvalidParam("value")
		return
	}

	ruser, err := c.App.VerifyFromStageToken(transaction.Token, transaction.Code)
	if err != nil {
		c.Err = err
		return
	}

	if ruser.Id != user.Id {
		c.SetInvalidParam("user_id")
		return
	}

	transaction.Description = "Списание вручную"

	_, err = c.App.DeductionTransaction(transaction)
	if err != nil {
		c.Err = err
		return
	}
	//w.Write([]byte(result.ToJson()))
	ReturnStatusOK(w)
}

func chargeTransactionUser(c *Context, w http.ResponseWriter, r *http.Request) {
	transaction := model.TransactionFromJson(r.Body)

	/*user, err := c.App.GetUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}*/

	if transaction == nil {
		c.SetInvalidParam("transaction")
		return
	}

	if len(transaction.UserId) != 26 {
		c.SetInvalidParam("user_id")
		return
	}

	transaction.Description = "Начисление вручную"

	result, err := c.App.AccrualTransaction(transaction)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func getAllTransactions(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterTransaction := r.URL.Query().Get("after")
	beforeTransaction := r.URL.Query().Get("before")
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

	var list *model.TransactionList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllTransactionsSince(since)
	} else if len(afterTransaction) > 0 {

		list, err = c.App.GetAllTransactionsAfterTransaction(afterTransaction, c.Params.Page, c.Params.PerPage)
	} else if len(beforeTransaction) > 0 {

		list, err = c.App.GetAllTransactionsBeforeTransaction(beforeTransaction, c.Params.Page, c.Params.PerPage)
	} else {
		list, err = c.App.GetAllTransactionsPage(c.Params.Page, c.Params.PerPage)
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

func getTransaction(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTransactionId()
	if c.Err != nil {
		return
	}

	transaction, err := c.App.GetTransaction(c.Params.TransactionId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(transaction.ToJson()))

}

func updateTransaction(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTransactionId()
	if c.Err != nil {
		return
	}

	transaction := model.TransactionFromJson(r.Body)

	if transaction == nil {
		c.SetInvalidParam("transaction")
		return
	}

	// The transaction being updated in the payload must be the same one as indicated in the URL.
	if transaction.Id != c.Params.TransactionId {
		c.SetInvalidParam("id")
		return
	}

	transaction.Id = c.Params.TransactionId

	rtransaction, err := c.App.UpdateTransaction(transaction, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rtransaction.ToJson()))
}

func createTransaction(c *Context, w http.ResponseWriter, r *http.Request) {

	transaction := model.TransactionFromJson(r.Body)

	if transaction == nil {
		c.SetInvalidParam("transaction")
		return
	}

	if len(transaction.UserId) != 26 {
		c.SetInvalidParam("user_id")
	}

	result, err := c.App.CreateTransaction(transaction)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func deleteTransaction(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTransactionId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetTransaction(c.Params.TransactionId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == transaction.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, transaction.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, transaction.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteTransaction(c.Params.TransactionId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func getUserTransactions(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	var list *model.TransactionList
	var err *model.AppError
	//etag := ""
	sort := r.URL.Query().Get("sort")
	transactionGetOptions := &model.TransactionGetOptions{
		Sort:    sort,
		Page:    c.Params.Page,
		PerPage: c.Params.PerPage,
		AppId:   c.Params.AppId,
	}

	list, err = c.App.GetUserTransactions(transactionGetOptions)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}
