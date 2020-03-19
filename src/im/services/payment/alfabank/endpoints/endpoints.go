package endpoints

// API endpoints
const (
	Register               string = "/register.do"
	RegisterPreAuth        string = "/registerPreAuth.do"
	Deposit                string = "/deposit.do"
	Reverse                string = "/reverse.do"
	Refund                 string = "/refund.do"
	GetOrderStatusExtended string = "/getOrderStatusExtended.do"
	GetReceiptStatus       string = "/getReceiptStatus.do"
	UnBindCard             string = "/unBindCard.do"
	BindCard               string = "/bindCard.do"
	GetBindings            string = "/getBindings.do"
	ExtendBinding          string = "/extendBinding.do"
	ApplePay               string = "/applepay/payment.do"
	SamsungPay             string = "/samsung/payment.do"
	GooglePay              string = "/google/payment.do"
	VerifyEnrollment       string = "/verifyEnrollment.do"
	UpdateSSLCardList      string = "/updateSSLCardList.do"
)
