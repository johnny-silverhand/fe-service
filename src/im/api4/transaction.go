package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitTransaction() {

	api.BaseRoutes.Transactions.Handle("", api.ApiHandler(getAllTransactions)).Methods("GET")
	api.BaseRoutes.Transactions.Handle("", api.ApiHandler(createTransaction)).Methods("POST")

	api.BaseRoutes.Transaction.Handle("", api.ApiHandler(getTransaction)).Methods("GET")
	api.BaseRoutes.Transaction.Handle("", api.ApiHandler(updateTransaction)).Methods("PUT")
	api.BaseRoutes.Transaction.Handle("", api.ApiHandler(deleteTransaction)).Methods("DELETE")
	api.BaseRoutes.User.Handle("/transactions", api.ApiSessionRequired(getUserTransactions)).Methods("GET")
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
	if c.Err != nil {
		return
	}

	var list *model.TransactionList
	var err *model.AppError
	//etag := ""

	list, err = c.App.GetUserTransactions(c.Params.UserId, c.Params.Page, c.Params.PerPage, "")

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}
