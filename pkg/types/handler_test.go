package processor

import (
    watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestMessageMetadata_AddTxMetadata(t *testing.T) {
    type args struct {
        tx watcher.Transaction
    }
    accepted1 := true
    hash1 := "this_is_hash"
    accepted2 := false
    hash2 := "this_is_hash2"
    tests := []struct {
        name        string
        args        args
        expected    *TxMetadata
    }{
        {"empty_data", args{}, &TxMetadata{}},
        {"first_transform", args{watcher.Transaction{Accepted: accepted1, Hash: hash1}}, &TxMetadata{accepted1, hash1}},
        {"second_transform", args{watcher.Transaction{Accepted: accepted2, Hash: hash2}}, &TxMetadata{accepted2, hash2}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := &MessageMetadata{}
            m.AddTxMetadata(tt.args.tx)
            assert.Equal(t, tt.expected, m.TxMetadata)
        })
    }
}
