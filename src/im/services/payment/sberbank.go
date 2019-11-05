package payment

import (
	"fmt"
	"im/model"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/currency"
	"net/http"
)
type SberBankBackend struct {

}
func (b *SberBankBackend) sbNew() (*sberbank.Client, error) {
	cfg := sberbank.ClientConfig{
		UserName:           "test-api", // Replace with your own
		Currency:           currency.RUB,
		Password:           "test", // Replace with your own
		Language:           "ru",
		SessionTimeoutSecs: 1200,
		SandboxMode:        true,
	}

	client, err := sberbank.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
	/*order := acquiring.Order{
		OrderNumber: "test",
		Amount:      100,
		Description: "My Order for Client",
	}
	result, _, err := client.RegisterOrder(context.Background(), order)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.ErrorCode)
	fmt.Println(result.ErrorMessage)
	fmt.Println(result.FormUrl)
	fmt.Println(result.OrderId)*/
}

func (b *SberBankBackend) TestConnection() *model.AppError {


	return nil
}


func (b *SberBankBackend) RegisterOrder() (url string, err *model.AppError) {

	sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	order := Order{
		OrderNumber: "test",
		Amount:      100,
		Description: "My Order for Client",
	}
	result, _, err := client.RegisterOrder(context.Background(), order)
	if err != nil {
		panic(err)
	}
	/*fmt.Println(result.ErrorCode)
	fmt.Println(result.ErrorMessage)
	fmt.Println(result.FormUrl)
	fmt.Println(result.OrderId)*/

	return nil
}
