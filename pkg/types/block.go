package types

import (
	"bytes"
	"encoding/hex"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth"
	watcher "github.com/mapofzones/cosmos-watcher/types"
	"github.com/tendermint/tendermint/types"
)

// Block is a unit of data being sent over in order to be processed
type Block watcher.Block

type TxWithHash struct {
	Hash string
	Tx   auth.StdTx
}

// GetValidStdTxs unmarshalls all txs into cosmos stdTx objects, returns error if this is not possible
// only txs with result code 0 are included
func (b Block) GetValidStdTxs() ([]TxWithHash, error) {
	txs := make([]TxWithHash, 0)
	for _, txBytes := range b.Txs {
		if b.valid(txBytes) {
			stdTx, err := Decode(txBytes)
			if err != nil {
				return nil, err
			}
			txs = append(txs, TxWithHash{Hash: hex.EncodeToString(txBytes.Hash()), Tx: stdTx})
		}
	}

	return txs, nil
}

// check tx error code, panic if txStatus slice doesn't have this tx
func (b Block) valid(tx types.Tx) bool {
	hash := tx.Hash()

	for _, status := range b.Results {
		if bytes.Equal(hash, status.Hash) {
			return status.ResultCode == 0
		}
	}

	panic("tx status is not stored inside the block")
}

func (b Block) ToTxStats(zone string, validTxCount int) TxStats {
	return TxStats{
		ChainID: zone,
		Count:   validTxCount,
		Hour:    b.T.Truncate(time.Hour),
	}
}
