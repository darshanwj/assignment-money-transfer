package service

import (
	"time"

	"github.com/Rhymond/go-money"
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

type InMemoryDb struct {
	accounts     []Account
	transactions []Transaction
	ledger       []Ledger
}

func (db *InMemoryDb) getAccounts() []Account {
	return db.accounts
}

func (db *InMemoryDb) createAccount(acc Account) {
	db.accounts = append(db.accounts, acc)
}

func (db *InMemoryDb) getAccountById(id uint) *Account {
	for i := range db.accounts {
		if db.accounts[i].Id == id {
			return &db.accounts[i]
		}
	}

	return nil
}

func (db *InMemoryDb) getLedger() []Ledger {
	return db.ledger
}

func (db *InMemoryDb) getTransactions() []Transaction {
	return db.transactions
}

func (db *InMemoryDb) createTransaction(txn Transaction) {
	db.transactions = append(db.transactions, txn)
}

func (db *InMemoryDb) createLedgerEntries(cEntry Ledger, dEntry Ledger) {
	db.ledger = append(db.ledger, cEntry, dEntry)
}
