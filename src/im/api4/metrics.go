package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitMetric() {

	api.BaseRoutes.Metrics.Handle("/clients", api.ApiSessionRequired(metricsForClients)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/orders", api.ApiSessionRequired(metricsForOrders)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/ratings", api.ApiSessionRequired(metricsForClientsRating)).Methods("GET")

}

func metricsForClientsRating(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	options := &model.UserGetOptions{
		AppId:   c.Params.AppId,
		Page:    c.Params.Page,
		PerPage: c.Params.PerPage,
	}

	result := <-c.App.Srv.Store.User().GetMetricsForRating(*options)
	if result.Err != nil {
		metrics := new(model.UserMetricsForRatingList)
		w.Write([]byte(metrics.ToJson()))
	} else {
		metrics := result.Data.(*model.UserMetricsForRatingList)
		w.Write([]byte(metrics.ToJson()))
	}
}

func metricsForClients(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	beginAtString := r.URL.Query().Get("begin_at")
	expireAtString := r.URL.Query().Get("expire_at")

	var beginAt int64
	var expireAt int64
	var parseError error

	if len(beginAtString) > 0 {
		beginAt, parseError = strconv.ParseInt(beginAtString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("begin_at")
			return
		}
	}
	if len(expireAtString) > 0 {
		expireAt, parseError = strconv.ParseInt(expireAtString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("expire_at")
			return
		}
	}

	result := <-c.App.Srv.Store.User().GetMetricsForRegister(c.Params.AppId, beginAt, expireAt)
	if result.Err != nil {
		metrics := new(model.MetricsForRegister)
		w.Write([]byte(metrics.ToJson()))
	} else {
		metrics := result.Data.(*model.MetricsForRegister)
		w.Write([]byte(metrics.ToJson()))
	}
}

func metricsForOrders(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	beginAtString := r.URL.Query().Get("begin_at")
	expireAtString := r.URL.Query().Get("expire_at")

	var beginAt int64
	var expireAt int64
	var parseError error

	if len(beginAtString) > 0 {
		beginAt, parseError = strconv.ParseInt(beginAtString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("begin_at")
			return
		}
	}
	if len(expireAtString) > 0 {
		expireAt, parseError = strconv.ParseInt(expireAtString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("expire_at")
			return
		}
	}

	result := <-c.App.Srv.Store.Order().GetMetricsForOrders(c.Params.AppId, beginAt, expireAt)
	if result.Err != nil {
		metrics := new(model.MetricsForOrders)
		w.Write([]byte(metrics.ToJson()))
	} else {
		metrics := result.Data.(*model.MetricsForOrders)
		w.Write([]byte(metrics.ToJson()))
	}
}
