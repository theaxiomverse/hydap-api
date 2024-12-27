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
	Txns map[string]*Transaction
	mu   sync.RWMutex
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		Txns: make(map[string]*Transaction),
	}
}

func (tm *TransactionManager) Begin(module string, op string) *Transaction {
	tx := &Transaction{
		ID:        uuid.NewString(),
		Module:    module,
		Operation: op,
		Status:    "pending",
	}
	tm.mu.Lock()
	tm.Txns[tx.ID] = tx
	tm.mu.Unlock()
	return tx
}

// GetTransaction retrieves a transaction by ID
func (tm *TransactionManager) GetTransaction(id string) (*Transaction, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tx, exists := tm.Txns[id]
	return tx, exists
}

// UpdateStatus updates the status of a transaction
func (tm *TransactionManager) UpdateStatus(id string, status string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tx, exists := tm.Txns[id]; exists {
		tx.Status = status
		return true
	}
	return false
}
