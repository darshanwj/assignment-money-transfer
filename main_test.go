package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darshanwj/assignment-money-transfer/service"

	"github.com/Rhymond/go-money"
	"github.com/stretchr/testify/assert"
)

var h = &handler{service.New()}

// Some test accounts
var tt = []struct {
	id         uint
	customerId uint
	currency   string
	balance    int64
}{
	{1, 1, "USD", 100},
	{2, 1, "AED", 200},
	{3, 2, "USD", 300},
}

func TestCreateAccountHandler(t *testing.T) {
	for _, tc := range tt {
		jsonBody := fmt.Sprintf(`{"id":%d,"customer_id":%d,"balance":{"amount":%d,"currency":"%s"}}`, tc.id, tc.customerId, tc.balance, tc.currency)
		req, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.CreateAccountHandler)

		handler.ServeHTTP(rr, req)
		assert.EqualValues(t, http.StatusOK, rr.Code)
		t.Log(rr.Body)

		var acc service.Account
		err = json.Unmarshal(rr.Body.Bytes(), &acc)
		assert.Nil(t, err)
		assert.NotNil(t, acc)
		assert.EqualValues(t, tc.id, acc.Id)
		assert.EqualValues(t, tc.customerId, acc.CustomerId)
		eq, err := money.New(tc.balance, tc.currency).Equals(&acc.Balance)
		assert.Nil(t, err)
		assert.True(t, eq)
	}
	// @TODO: Test some invalid requests
}

func TestGetAccountsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetAccountsHandler)

	handler.ServeHTTP(rr, req)
	assert.EqualValues(t, http.StatusOK, rr.Code)
	t.Log(rr.Body)

	var accs []service.Account
	err = json.Unmarshal(rr.Body.Bytes(), &accs)
	assert.Nil(t, err)
	assert.NotNil(t, accs)
	assert.NotEmpty(t, accs)

	var i = 0
	for _, tc := range tt {
		acc := accs[i]
		assert.NotNil(t, acc)
		assert.EqualValues(t, tc.id, acc.Id)
		assert.EqualValues(t, tc.customerId, acc.CustomerId)
		eq, err := money.New(tc.balance, tc.currency).Equals(&acc.Balance)
		assert.Nil(t, err)
		assert.True(t, eq)
		i++
	}
}

func TestTransferHandlerInvalid(t *testing.T) {
	tt := []struct {
		sender   uint
		receiver uint
		amount   uint
		currency string
		message  string
	}{
		{800, 1, 200, "USD", "Sender account not found\n"},
		{1, 600, 200, "USD", "Receiver account not found\n"},
		{1, 3, 200, "GBP", "Currency conversion not supported\n"},
		{1, 2, 100, "USD", "Receiver account currency not same as sender\n"},
		{1, 3, 111, "USD", "Not enough balance to perform this transfer\n"},
		{1, 3, 101, "USD", "Not enough balance to perform this transfer\n"},
	}

	for _, tc := range tt {
		jsonBody := fmt.Sprintf(`{"sender_account_id":%d,"receiver_account_id":%d,"amount":{"amount":%d,"currency":"%s"}}`, tc.sender, tc.receiver, tc.amount, tc.currency)
		req, err := http.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.TransferHandler)

		handler.ServeHTTP(rr, req)
		assert.EqualValues(t, http.StatusBadRequest, rr.Code)
		t.Log(rr.Body)

		apiErr := rr.Body.String()
		assert.NotNil(t, apiErr)
		assert.EqualValues(t, tc.message, apiErr)
	}
}

func TestTransferHandler(t *testing.T) {
	amount := 70
	currency := "USD"
	jsonBody := fmt.Sprintf(`{"sender_account_id":%d,"receiver_account_id":%d,"amount":{"amount":%d,"currency":"%s"}}`, tt[0].id, tt[2].id, amount, currency)
	req, err := http.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.TransferHandler)

	handler.ServeHTTP(rr, req)
	assert.EqualValues(t, http.StatusOK, rr.Code)
	t.Log(rr.Body)

	var acc service.Account
	err = json.Unmarshal(rr.Body.Bytes(), &acc)
	assert.Nil(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, tt[0].id, acc.Id)
	assert.Equal(t, tt[0].customerId, acc.CustomerId)

	// Test sender account balance updated
	trn := money.New(int64(amount), currency)
	sender := h.GetAccountById(tt[0].id)
	before := money.New(tt[0].balance, tt[0].currency)
	ex, _ := sender.Balance.Add(trn)
	eq, _ := before.Equals(ex)
	assert.True(t, eq)

	// Test receiver account balance updated
	receiver := h.GetAccountById(tt[2].id)
	before = money.New(tt[2].balance, tt[2].currency)
	ex, _ = receiver.Balance.Subtract(trn)
	eq, _ = before.Equals(ex)
	assert.True(t, eq)

	// Test transaction
	txns := h.GetTransactions()
	t.Log(txns)
	assert.NotEmpty(t, txns)
	txn := txns[0]
	assert.Equal(t, service.TxnTypeTransfer, txn.Type)

	// Test ledger entries
	ledger := h.GetLedgerEntries()
	t.Log(ledger)
	assert.NotEmpty(t, ledger)
	lc := ledger[0]
	assert.Equal(t, tt[2].id, lc.AccountId)
	assert.Equal(t, txn.Id, lc.TransactionId)
	assert.Equal(t, service.EntryTypeCredit, lc.Type)
	eq, _ = trn.Equals(&lc.Amount)
	assert.True(t, eq)
	ld := ledger[1]
	assert.Equal(t, tt[0].id, ld.AccountId)
	assert.Equal(t, txn.Id, ld.TransactionId)
	assert.Equal(t, service.EntryTypeDebit, ld.Type)
	eq, _ = trn.Equals(&ld.Amount)
	assert.True(t, eq)
}
