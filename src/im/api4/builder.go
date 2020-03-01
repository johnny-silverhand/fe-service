package api4

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (api *API) InitBuilder() {
	api.BaseRoutes.Builder.Handle("/status", api.ApiHandler(getBuilderStatus)).Methods("GET")
	api.BaseRoutes.Builder.Handle("/download", api.ApiHandler(downloadBuild)).Methods("GET")
	api.BaseRoutes.Builder.Handle("/create", api.ApiHandler(createBuild)).Methods("POST")
}

func getBuilderStatus(c *Context, w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://foodexpress.nbr9.com/index.php?key=feapitest&action=check-build-status")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	var json string
	for true {
		bs := make([]byte, 1014)
		n, err := resp.Body.Read(bs)
		json = string(bs[:n])
		if n == 0 || err != nil{
			break
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(json))
}

func downloadBuild(c *Context, w http.ResponseWriter, r *http.Request) {


	url := "https://foodexpress.nbr9.com/index.php?key=feapitest&action=download-build&build_id=" + c.Params.BuildId + "&type=" + c.Params.BuildType
	fmt.Printf("HTTPDownload From: %s.\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	d, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("ReadFile: Size of download: %d\n Content-Length: %d\n", len(d), resp.ContentLength)

	//stream the body to the client without fully loading it into memory
	io.Copy(w, resp.Body)
	w.Write(d)
}

type Person struct {
	Name string
	Age  int
}

func createBuild(c *Context, w http.ResponseWriter, r *http.Request) {

	uri := "https://foodexpress.nbr9.com/index.php?key=feapitest&action=new-build"

	var json string
	bs := make([]byte, 1014)
	for true {
		n, err := r.Body.Read(bs)
		json = string(bs[:n])
		if n == 0 || err != nil{
			break
		}
	}
	fmt.Printf("HTTPDownload From: %s.\n", json)

	formData := url.Values{"settings": {json}}
	resp, err := http.PostForm(uri, formData)

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	w.Write(d)
}
