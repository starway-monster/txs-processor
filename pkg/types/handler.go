package processor

import (
	"context"
	"time"

	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
)

// Handler is a function which processor interface must implement
// this function is expected to handle all types of messages (or do nothing upon receiving them)
type Handler interface {
	Handler(watcher.Message) func(context.Context, MessageMetadata, watcher.Message) error
}

// MessageMetada is info which might be needed inside handler function
type MessageMetadata struct {
	ChainID   string
	BlockTime time.Time
	// if this pointer is not nil, then message has happened inside tx
	*TxMetadata
}

type TxMetadata struct {
	Accepted bool
	Hash     string
}

func (m *MessageMetadata) AddTxMetadata(tx watcher.Transaction) {
	m.TxMetadata = &TxMetadata{
		Accepted: tx.Accepted,
		Hash:     tx.Hash,
	}
}
