package aquiring

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	username string
	password string
	baseURL  url.URL
}

func NewAlfaClient(username, password string) *Client {
	return &Client{
		username: username,
		password: password,
		baseURL: url.URL{
			Scheme: "https",
			Host:   "web.rbsuat.com",
			Path:   "/ab/rest",
		},
	}
}

func NewSberClient(username, password string) *Client {
	return &Client{
		username: username,
		password: password,
		baseURL: url.URL{
			Scheme: "https",
			Host:   "3dsec.sberbank.ru:443",
			Path:   "payment/rest",
		},
	}
}

func (c *Client) buildLink(action string, request RequestInterface) error {
	c.baseURL.Path += action
	var query []string
	if rmap, err := ToMap(request, "json"); err != nil {
		return err
	} else {
		query = append(query, "userName="+c.username)
		query = append(query, "password="+c.password)
		for i, value := range rmap {
			s := fmt.Sprintf("%v", value)            // interface to string
			if s != "<nil>" && s != "" && s != "0" { // костыль на то что nil != nil interface
				query = append(query, i+"="+s)
			}
		}
	}
	c.baseURL.RawQuery = strings.Join(query, "&")
	return nil
}

func (c *Client) PostRequest(action string, request RequestInterface) (*http.Response, error) {
	if err := c.buildLink(action, request); err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL.String(), "text/plain", nil)

	//fmt.Println(c.baseURL.String()) // debug

	return resp, err
}
