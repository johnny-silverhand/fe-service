package endpoints

// API endpoints
const (
	Register               string = "/ab/rest/register.do"
	RegisterPreAuth        string = "/ab/rest/registerPreAuth.do"
	Deposit                string = "/ab/rest/deposit.do"
	Reverse                string = "/ab/rest/reverse.do"
	Refund                 string = "/ab/rest/refund.do"
	GetOrderStatusExtended string = "/ab/rest/getOrderStatusExtended.do"
	GetReceiptStatus       string = "/ab/rest/getReceiptStatus.do"
	UnBindCard             string = "/ab/rest/unBindCard.do"
	BindCard               string = "/ab/rest/bindCard.do"
	GetBindings            string = "/ab/rest/getBindings.do"
	ExtendBinding          string = "/ab/rest/extendBinding.do"
	ApplePay               string = "/ab/applepay/payment.do"
	SamsungPay             string = "/ab/samsung/payment.do"
	GooglePay              string = "/ab/google/payment.do"
	VerifyEnrollment       string = "/ab/rest/verifyEnrollment.do"
	UpdateSSLCardList      string = "/ab/rest/updateSSLCardList.do"
)
