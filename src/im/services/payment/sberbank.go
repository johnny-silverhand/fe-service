package payment

import (
	"context"
	"fmt"
	"im/model"
	"im/services/payment/sberbank"
	"im/services/payment/sberbank/currency"
	"net/http"
	"strconv"
)

type SberBankBackend struct {
}

func (b *SberBankBackend) sbNew() (*sberbank.Client, error) {
	cfg := sberbank.ClientConfig{
		UserName:           "foodexp-api", // Replace with your own
		Currency:           currency.RUB,
		Password:           "foodexp", // Replace with your own
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

	/*sbClnt, err := b.sbNew()
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
	}*/

	var client *sberbank.Client

	if c, err := b.sbNew(); err != nil {
		return "", model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		client = c
	}

	order := sberbank.Order{
		OrderNumber: strconv.FormatInt(model.GetMillis(), 10),
		Amount:      100,
		Description: "My Order for Client",
		ReturnURL:   "https://yandex.ru",
	}

	/*if result, _, err := client.RegisterOrderPreAuth(context.Background(), order); err != nil {
		return "", model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Println(result)
	}*/

	if result, _, err := client.RegisterOrder(context.Background(), order); err != nil {
		return "", model.NewAppError("", "", nil, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Println(result)
	}
	/*fmt.Println(result.ErrorCode)
	fmt.Println(result.ErrorMessage)
	fmt.Println(result.FormUrl)
	fmt.Println(result.OrderId)*/

	return "", nil
}
