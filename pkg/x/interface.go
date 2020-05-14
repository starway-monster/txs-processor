package processor

import (
	"context"

	"github.com/mapofzones/txs-processor/pkg/types"
	postgres "github.com/mapofzones/txs-processor/pkg/x/postgres_impl"
)

// compile-time interface check
var _ Processor = &postgres.PostgresProcessor{}

type Querier interface {
	// method that lets us know what block we need to get, since must come in valid order
	LastProcessedBlock(string) (int64, error)

	// methods that lets us know what chain_id given some sort kind of ibc data
	ChainIDFromClientID(string) (string, error)
	ChainIDFromConnectionID(string) (string, error)
	ChainIDFromChannelID(string) (string, error)
}

type Processor interface {
	// must be able to fetch data specified in the querier to work
	Querier

	// creates entry about given blockchain
	// should be called first
	AddZone(string)

	// increment block count
	MarkBlock()

	// add data about transaction stats
	AddTxStats(types.TxStats)

	// ibc protocol stuff
	AddClient(string, string)
	AddConnection(string, string)
	AddChannel(string, string)

	// functions that change the state of the channel
	MarkChannelOpened(string)
	MarkChannelClosed(string)

	// add ibc transfers statistics
	AddIbcStats(types.IbcStats)

	// apply changes previously appended by all of the above functions
	Commit(context.Context) error

	// Reset should be called before calling any methods on Processor
	// to clear all data which might have been previously stored in it
	// since new version of processor might not be created when we create new block
	// because they have to hold file / db handles
	Reset()
}
