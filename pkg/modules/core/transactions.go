package core

import (
	"github.com/google/uuid"
	"sync"
)

type Transaction struct {
	ID        string
	Module    string
	Operation string
	Data      []byte
	Status    string
}

type TransactionManager struct {
	txns map[string]*Transaction
	mu   sync.RWMutex
}

func (tm *TransactionManager) Begin(module string, op string) *Transaction {
	tx := &Transaction{
		ID:        uuid.New().String(),
		Module:    module,
		Operation: op,
		Status:    "pending",
	}
	tm.mu.Lock()
	tm.txns[tx.ID] = tx
	tm.mu.Unlock()
	return tx
}
