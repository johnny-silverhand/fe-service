package api4

import (
	"net/http"
)

func (api *API) InitSection() {
//	api.BaseRoutes.Sections.Handle("", api.ApiHandler(testSection)).Methods("GET")
}

func testSection(c *Context, w http.ResponseWriter, r *http.Request) {

	result, _ := c.App.GenerateTree()
	w.Write([]byte(result.ToJson()))
}
