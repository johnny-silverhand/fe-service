package api4

import (
	"encoding/json"
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitMetric() {

	api.BaseRoutes.Metrics.Handle("/clients", api.ApiSessionRequired(metricsForClients)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/orders", api.ApiSessionRequired(metricsForOrders)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/ratings", api.ApiSessionRequired(metricsForClientsRating)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/bonuses", api.ApiSessionRequired(metricsForBonuses)).Methods("GET")
	api.BaseRoutes.Metrics.Handle("/spy", api.ApiSessionRequired(metricsForSpy)).Methods("GET")

}

func metricsForSpy(c *Context, w http.ResponseWriter, r *http.Request) {
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

	options := &model.UserGetOptions{
		AppId:   c.Params.AppId,
		Page:    c.Params.Page,
		PerPage: c.Params.PerPage,
	}

	if metrics, err := c.App.GetMetricsForSpy(*options, beginAt, expireAt); err != nil {
		c.Err = err
		return
	} else {
		//w.Write([]byte(metrics.ToJson()))
		if metrics == nil {
			w.Write([]byte("[]"))
			return
		}
		copy := metrics
		b, err := json.Marshal(&copy)
		if err != nil {
			w.Write([]byte("[]"))
		} else {
			w.Write([]byte(string(b)))
		}
	}
}

func metricsForBonuses(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	if metrics, err := c.App.GetMetricsForBonuses(c.Params.AppId); err != nil {
		c.Err = err
		return
	} else {
		if metrics == nil {
			w.Write([]byte("[]"))
			return
		}
		copy := metrics
		b, err := json.Marshal(&copy)
		if err != nil {
			w.Write([]byte("[]"))
		} else {
			w.Write([]byte(string(b)))
		}
	}
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

	if metrics, err := c.App.GetMetricsForRating(*options); err != nil {
		c.Err = err
		return
	} else {
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

	if metrics, err := c.App.GetMetricsForRegister(c.Params.AppId, beginAt, expireAt); err != nil {
		c.Err = err
		return
	} else {
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

	if metrics, err := c.App.GetMetricsForOrders(c.Params.AppId, beginAt, expireAt); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(metrics.ToJson()))
	}
}
