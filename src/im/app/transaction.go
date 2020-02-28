package app

import (
	"im/model"
	"im/store"
	"math"
	"net/http"
)

func (a *App) GetTransaction(transactionId string) (*model.Transaction, *model.AppError) {

	result := <-a.Srv.Store.Transaction().Get(transactionId)
	if result.Err != nil {
		return nil, result.Err
	}

	rtransaction := result.Data.(*model.Transaction)

	rtransaction = a.PrepareTransactionForClient(rtransaction, false)

	return rtransaction, nil
}

func (a *App) GetTransactionsPage(page int, perPage int, sort string) (*model.TransactionList, *model.AppError) {
	return a.GetTransactions(page*perPage, perPage, sort)
}

func (a *App) GetTransactions(offset int, limit int, sort string) (*model.TransactionList, *model.AppError) {

	result := <-a.Srv.Store.Transaction().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.TransactionList), nil
}

func (a *App) CreateTransaction(transaction *model.Transaction) (*model.Transaction, *model.AppError) {

	result := <-a.Srv.Store.Transaction().Save(transaction)
	if result.Err != nil {
		return nil, result.Err
	}

	rtransaction := result.Data.(*model.Transaction)

	return rtransaction, nil
}

func (a *App) UpdateTransaction(transaction *model.Transaction, safeUpdate bool) (*model.Transaction, *model.AppError) {
	//transaction.SanitizeProps()

	result := <-a.Srv.Store.Transaction().Get(transaction.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldTransaction := result.Data.(*model.Transaction)

	if oldTransaction == nil {
		err := model.NewAppError("UpdateTransaction", "api.transaction.update_transaction.find.app_error", nil, "id="+transaction.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldTransaction.DeleteAt != 0 {
		err := model.NewAppError("UpdateTransaction", "api.transaction.update_transaction.permissions_details.app_error", map[string]interface{}{"TransactionId": transaction.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newTransaction := &model.Transaction{}
	*newTransaction = *oldTransaction

	newTransaction.Description = transaction.Description

	result = <-a.Srv.Store.Transaction().Update(newTransaction)
	if result.Err != nil {
		return nil, result.Err
	}

	rtransaction := result.Data.(*model.Transaction)
	rtransaction = a.PrepareTransactionForClient(rtransaction, false)

	//a.InvalidateCacheForChannelTransactions(rtransaction.ChannelId)

	return rtransaction, nil
}

func (a *App) PrepareTransactionForClient(originalTransaction *model.Transaction, isNewTransaction bool) *model.Transaction {
	transaction := originalTransaction.Clone()

	//transaction.Metadata.Images = a.getCategoryForTransaction(transaction)

	return transaction
}

func (a *App) PrepareTransactionListForClient(originalList *model.TransactionList) *model.TransactionList {
	list := &model.TransactionList{
		Transactions: make(map[string]*model.Transaction, len(originalList.Transactions)),
		Order:        originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalTransaction := range originalList.Transactions {
		transaction := a.PrepareTransactionForClient(originalTransaction, false)

		list.Transactions[id] = transaction
	}

	return list
}

func (a *App) DeleteTransaction(transactionId, deleteByID string) (*model.Transaction, *model.AppError) {
	result := <-a.Srv.Store.Transaction().Get(transactionId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	transaction := result.Data.(*model.Transaction)

	if result := <-a.Srv.Store.Transaction().Delete(transactionId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return transaction, nil
}

func (a *App) GetAllTransactionsBeforeTransaction(transactionId string, page, perPage int) (*model.TransactionList, *model.AppError) {

	if result := <-a.Srv.Store.Transaction().GetAllTransactionsBefore(transactionId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.TransactionList), nil
	}
}

func (a *App) GetAllTransactionsAfterTransaction(transactionId string, page, perPage int) (*model.TransactionList, *model.AppError) {

	if result := <-a.Srv.Store.Transaction().GetAllTransactionsAfter(transactionId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.TransactionList), nil
	}
}

func (a *App) GetAllTransactionsAroundTransaction(transactionId string, offset, limit int, before bool) (*model.TransactionList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Transaction().GetAllTransactionsBefore(transactionId, limit, offset)
	} else {
		pchan = a.Srv.Store.Transaction().GetAllTransactionsAfter(transactionId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.TransactionList), nil
	}
}

func (a *App) GetAllTransactionsSince(time int64) (*model.TransactionList, *model.AppError) {
	if result := <-a.Srv.Store.Transaction().GetAllTransactionsSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.TransactionList), nil
	}
}

func (a *App) GetAllTransactionsPage(page int, perPage int) (*model.TransactionList, *model.AppError) {
	if result := <-a.Srv.Store.Transaction().GetAllTransactions(page*perPage, perPage, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.TransactionList), nil
	}
}

func (a *App) AccrualTransaction(transaction *model.Transaction) (*model.Transaction, *model.AppError) {
	transaction.Value = math.Abs(transaction.Value)
	result := <-a.Srv.Store.Transaction().Save(transaction)
	if result.Err != nil {
		return nil, result.Err
	}

	a.AccrualBalance(transaction.UserId, transaction.Value)

	rtransaction := result.Data.(*model.Transaction)

	return rtransaction, nil
}

func (a *App) DeductionTransaction(transaction *model.Transaction) (*model.Transaction, *model.AppError) {
	transaction.Value = 0 - math.Abs(transaction.Value)
	result := <-a.Srv.Store.Transaction().Save(transaction)
	if result.Err != nil {
		return nil, result.Err
	}

	a.DeductionBalance(transaction.UserId, transaction.Value)

	rtransaction := result.Data.(*model.Transaction)

	return rtransaction, nil
}

func (a *App) GetUserTransactions(options *model.TransactionGetOptions) (*model.TransactionList, *model.AppError) {
	//if result := <-a.Srv.Store.Transaction().GetByUserId(userId, page*perPage, perPage, model.GetOrder(sort)); result.Err != nil {
	if result := <-a.Srv.Store.Transaction().GetByUserId(*options); result.Err != nil {
		return nil, result.Err
	} else {
		transactionList := result.Data.(*model.TransactionList)
		return a.PrepareTransactionListForClient(transactionList), nil
	}
}

func (a *App) GetBonusTransactionsForUser(orderUserId string, userId string) (*model.TransactionList, *model.AppError) {

	if result := <-a.Srv.Store.Transaction().GetBonusTransactionsForUser(orderUserId, userId); result.Err != nil {
		return nil, result.Err
	} else {
		transactionList := result.Data.(*model.TransactionList)
		return a.PrepareTransactionListForClient(transactionList), nil
	}

}
