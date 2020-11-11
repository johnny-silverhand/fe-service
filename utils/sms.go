package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Status struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AuthResponse struct {
	Status
	Data interface{} `json:"data"`
}

type SendResponse struct {
	Status
	Data []SmsMessage `json:"data"`
}

type SmsMessage struct {
	ID           int64   `json:"id"`
	From         string  `json:"from"`
	Number       string  `json:"number"`
	Text         string  `json:"text"`
	Status       int64   `json:"status"`
	ExtendStatus string  `json:"extendStatus"`
	Channel      string  `json:"channel"`
	Cost         float64 `json:"cost"`
	DateCreate   int64   `json:"dateCreate"`
	DateSend     int64   `json:"dateSend"`
}

type MessageRequest struct {
	Numbers     []string   // numbers array Обязательно (на выбор)	Номера телефонов
	Sign        string     // sign	string	Обязательно	Подпись отправителя
	Text        string     // text	string	Обязательно	Текст сообщения
	Channel     string     // channel	string	Обязательно	Канал отправки
	DateSend    *time.Time // dateSend	integer	Не обязательно	Дата для отложенной отправки сообщения (в формате unixtime)
	CallbackUrl string     // callbackUrl	string	Не обязательно	url для отправки статуса сообщения в формате http://your.site или https://your.site (в ответ система ждет ста
}

const DefaultAddress = "gate.smsaero.ru"
const DefaultVersion = "v2"
const DefaultMethod = "GET"
const AcceptJson = "application/json"

const (
	ChannelInfo          = "INFO"          // Инфоподпись для всех операторов
	ChannelDigital       = "DIGITAL"       // Цифровой канал отправки (допускается только транзакционный трафик)
	ChannelInternational = "INTERNATIONAL" // Международная доставка (Операторы РФ, Казахстана, Украины и Белоруссии)
	ChannelDirect        = "DIRECT"        // Рекламный канал отправки сообщений с бесплатной буквенной подписью.
	ChannelService       = "SERVICE"       // Сервисный канал для отправки сервисных SMS по утвержденному шаблону с платной подписью отправителя.
)

type Client struct {
	Login  string
	ApiKey string

	address string
	version string
	method  string

	client   *http.Client
	testMode bool
}

func NewClient(login, apiKey, address, version string) *Client {

	if len(address) == 0 {
		address = DefaultAddress
	}

	if len(version) == 0 {
		version = DefaultVersion
	}

	return &Client{
		Login:  login,
		ApiKey: apiKey,

		address: address,
		version: version,
		method:  DefaultMethod,
		client:  http.DefaultClient,
	}
}

func (c *Client) TestMode(mode bool) *Client {
	c.testMode = mode
	return c
}

func (c *Client) Auth() (AuthResponse, error) {

	aResp := AuthResponse{}

	params, err := c.getFullUrl("auth")
	if err != nil {
		return aResp, err
	}

	req, err := c.createRequest(params)
	if err != nil {
		return aResp, err
	}

	err = c.callApi(req, &aResp)
	if err != nil {
		return aResp, err
	}

	return aResp, nil
}

func (c *Client) Send(msg MessageRequest) (SendResponse, error) {

	sResp := SendResponse{}

	method := "sms/send"
	if c.testMode {
		method = "sms/testsend"
	}

	params, err := c.getFullUrl(method)
	if err != nil {
		return sResp, err
	}

	if len(msg.Numbers) == 0 {
		return sResp, errors.New("required parameter MessageRequest.Numbers is empty")
	}

	qParams := url.Values{}
	for _, phone := range msg.Numbers {
		qParams.Add("numbers[]", phone)
	}

	qParams.Add("sign", msg.Sign)
	qParams.Add("text", msg.Text)
	qParams.Add("channel", msg.Channel)

	if msg.DateSend != nil {
		qParams.Add("dateSend", fmt.Sprintf("%d", msg.DateSend.Unix()))
	}

	if len(msg.CallbackUrl) > 0 {
		qParams.Add("callbackUrl", msg.CallbackUrl)
	}

	params.RawQuery = qParams.Encode()

	req, err := c.createRequest(params)
	if err != nil {
		return sResp, err
	}

	err = c.callApi(req, &sResp)
	if err != nil {
		return sResp, err
	}

	return sResp, nil
}

func (c *Client) createRequest(params *url.URL) (*http.Request, error) {

	req, err := http.NewRequest(c.method, params.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", AcceptJson)
	return req, nil
}

func (c *Client) callApi(req *http.Request, response interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getFullUrl(path string) (*url.URL, error) {
	rawUrl := fmt.Sprintf("https://%s:%s@%s/%s/%s", c.Login, c.ApiKey, c.address, c.version, path)
	return url.Parse(rawUrl)
}
