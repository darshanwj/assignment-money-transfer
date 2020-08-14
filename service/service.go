package service

import (
	"errors"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/rs/xid"
)

type Service interface {
	GetAccounts() []Account
	CreateAccount(acc Account)
	Transfer(trn Transfer) (*Account, error)
	GetAccountById(id uint) *Account
	GetLedgerEntries() []Ledger
	GetTransactions() []Transaction
}

type service struct {
	repo InMemoryDb
}

func New(repo InMemoryDb) Service {
	return &service{repo: repo}
}

func (s *service) GetAccounts() []Account {
	return s.repo.getAccounts()
}

func (s *service) CreateAccount(acc Account) {
	s.repo.createAccount(acc)
	// @TODO: Create ledger entries
}

func (s *service) Transfer(trn Transfer) (*Account, error) {
	// Validate first
	sender, receiver, err := s.validateTransfer(trn)
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

func (s *service) validateTransfer(trn Transfer) (sender *Account, receiver *Account, err error) {
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

func (s *service) GetAccountById(id uint) *Account {
	return s.repo.getAccountById(id)
}

func (s *service) GetLedgerEntries() []Ledger {
	return s.repo.getLedger()
}

func (s *service) GetTransactions() []Transaction {
	return s.repo.getTransactions()
}

// createTransaction should be part of an atomic operation
func (s *service) createTransaction(txnType string) Transaction {
	txn := Transaction{Id: xid.New().String(), Ts: time.Now(), Type: txnType}
	s.repo.createTransaction(txn)
	return txn
}

// createLedgerEntries creates Credit Debit pair, should be part of an atomic operation
func (s *service) createLedgerEntries(cAcc Account, dAcc Account, txn Transaction, amount money.Money) {
	s.repo.createLedgerEntries(Ledger{
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
