package service

import (
	"errors"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/rs/xid"
)

const (
	TxnTypeOpeningBalance = "opening_balance"
	TxnTypeTransfer       = "transfer"
	EntryTypeCredit       = "C"
	EntryTypeDebit        = "D"
)

type Account struct {
	Id         uint        `json:"id"`
	CustomerId uint        `json:"customer_id"`
	Balance    money.Money `json:"balance"`
}

type Transaction struct {
	Id   string    `json:"id"`
	Type string    `json:"type"`
	Ts   time.Time `json:"timestamp"`
}

type Ledger struct {
	Id            string      `json:"id"`
	AccountId     uint        `json:"account_id"`
	TransactionId string      `json:"transaction_id"`
	Type          string      `json:"type"`
	Amount        money.Money `json:"amount"`
}

type Transfer struct {
	Sender   uint        `json:"sender_account_id"`
	Receiver uint        `json:"receiver_account_id"`
	Amount   money.Money `json:"amount"`
}

var accounts []Account
var transactions []Transaction
var ledger []Ledger

type Service interface {
	GetAccounts() []Account
	CreateAccount(acc Account)
	ValidateTransfer(trn Transfer) (sender *Account, receiver *Account, err error)
	Transfer(trn Transfer) (*Account, error)
	GetAccountById(id uint) *Account
	GetLedgerEntries() []Ledger
	GetTransactions() []Transaction
	createTransaction(txnType string) Transaction
	createLedgerEntries(cAcc Account, dAcc Account, txn Transaction, amount money.Money)
	updateAccountBalances(sender *Account, receiver *Account, amount money.Money) error
}

type service struct{}

func New() Service {
	return &service{}
}

// Get all accounts
func (s *service) GetAccounts() []Account {
	return accounts
}

// Saves a single account
func (s *service) CreateAccount(acc Account) {
	accounts = append(accounts, acc)
	// @TODO: Create ledger entries
}

// Validates and saves a transfer
func (s *service) Transfer(trn Transfer) (*Account, error) {

	sender, receiver, err := s.ValidateTransfer(trn)
	if err != nil {
		return nil, err
	}

	// START TRANSACTION if this was a db

	txn := s.createTransaction(TxnTypeTransfer)

	s.createLedgerEntries(*receiver, *sender, txn, trn.Amount)

	if err = s.updateAccountBalances(sender, receiver, trn.Amount); err != nil {
		return nil, err
	}

	// COMMIT if this was a db

	return sender, nil
}

// Validates a transfer
func (s *service) ValidateTransfer(trn Transfer) (sender *Account, receiver *Account, err error) {
	// Sender account should exist
	sender = s.GetAccountById(trn.Sender)
	if sender == nil {
		return nil, nil, errors.New("Sender account not found")
	}
	// Receiver account should exist
	receiver = s.GetAccountById(trn.Receiver)
	if receiver == nil {
		return nil, nil, errors.New("Receiver account not found")
	}
	// Both account currencies should match
	if !sender.Balance.SameCurrency(&receiver.Balance) {
		return nil, nil, errors.New("Receiver account currency not same as sender")
	}
	// Transfer currency should be same as sender account
	poor, err := sender.Balance.LessThan(&trn.Amount)
	if err != nil {
		return nil, nil, errors.New("Currency conversion not supported")
	}
	// Sender should have enough balance
	if poor {
		return nil, nil, errors.New("Not enough balance to perform this transfer")
	}

	return sender, receiver, nil
}

// Gets single account by id
func (s *service) GetAccountById(id uint) *Account {
	for i := range accounts {
		if accounts[i].Id == id {
			return &accounts[i]
		}
	}

	return nil
}

// Gets all ledger entries
func (s *service) GetLedgerEntries() []Ledger {
	return ledger
}

// Gets all transactions
func (s *service) GetTransactions() []Transaction {
	return transactions
}

// Create transaction, should be part of an atomic operation
func (s *service) createTransaction(txnType string) Transaction {
	txn := Transaction{Id: xid.New().String(), Ts: time.Now(), Type: txnType}
	transactions = append(transactions, txn)
	return txn
}

// Credit Debit pair, should be part of an atomic operation
func (s *service) createLedgerEntries(cAcc Account, dAcc Account, txn Transaction, amount money.Money) {
	ledger = append(ledger, Ledger{
		Id:            xid.New().String(),
		AccountId:     cAcc.Id,
		Amount:        amount,
		TransactionId: txn.Id,
		Type:          EntryTypeCredit,
	}, Ledger{
		Id:            xid.New().String(),
		AccountId:     dAcc.Id,
		Amount:        amount,
		TransactionId: txn.Id,
		Type:          EntryTypeDebit,
	})
}

// Credit one account, debit the other, should be part of an atomic operation
func (s *service) updateAccountBalances(sender *Account, receiver *Account, amount money.Money) error {
	sb, err := sender.Balance.Subtract(&amount)
	if err != nil {
		return err
	}
	sender.Balance = *sb

	rb, err := receiver.Balance.Add(&amount)
	if err != nil {
		return err
	}
	receiver.Balance = *rb

	return nil
}
