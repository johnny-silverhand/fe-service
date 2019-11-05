package payment

import (
	"io"
	"net/http"

	"im/model"
)

type PaymentBackend interface {
	TestConnection() *model.AppError

	Reader(path string) (io.ReadCloser, *model.AppError)
	ReadFile(path string) ([]byte, *model.AppError)
	FileExists(path string) (bool, *model.AppError)
	CopyFile(oldPath, newPath string) *model.AppError
	MoveFile(oldPath, newPath string) *model.AppError
	WriteFile(fr io.Reader, path string) (int64, *model.AppError)
	RemoveFile(path string) *model.AppError

	ListDirectory(path string) (*[]string, *model.AppError)
	RemoveDirectory(path string) *model.AppError
}

func NewPaymentBackend(settings *model.PaymentBackendSettings) (PaymentBackend, *model.AppError) {
	switch *settings.Backend {
	case model.PAYMENT_PROXY_TYPE_SBERBANK:
		return &SberBankBackend{
			merchantId:  *settings.MerchantId,
			password: *settings.Password,
			username: *settings.UserName,
			currency: *settings.Currency,
			language: *settings.Language,
		}, nil
	
	}
	return nil, model.NewAppError("NewPaymentBackend", "api.payment.no_driver.app_error", nil, "", http.StatusInternalServerError)
}
