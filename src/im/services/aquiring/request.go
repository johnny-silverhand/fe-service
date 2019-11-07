package aquiring

type RequestInterface interface {
}

type RequestRegistration struct {
	RequestInterface

	Username    string `json:"userName"`
	Password    string `json:"password"`
	OrderNumber string `json:"orderNumber"`
	Amount      string `json:"amount"`
	ReturnUrl   string `json:"returnUrl"`

	Description string `json:"description"`

	//PageView			string `json:"pageView"`

	/*Currency			*int `json:"currency"`
	SessionTimeoutSecs	*int `json:"sessionTimeoutSecs"`
	FailUrl				*string `json:"failUrl"`
	Language 			*string `json:"language"`
	PageView			*string `json:"pageView"`
	ClientId			*string `json:"clientId"`
	MerchantLogin 		*string `json:"merchantLogin"`
	JsonParams			*string `json:"jsonParams"`
	ExpirationDate		*string `json:"expirationDate"`
	BindingId			*string `json:"bindingId"`
	Features			*string `json:"features"`*/
}

type RequestOrderStatus struct {
	RequestInterface

	Username string `json:"userName"`
	Password string `json:"password"`
	OrderId  string `json:"orderId"`

	Language string `json:"language"`
}
