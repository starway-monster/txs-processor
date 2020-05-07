// this is not complete API of the graphql endpoint, here there are only methods which are
// currently needed by the application logic, so unused methods might be deleted as well as new ones added
package graphql

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/machinebox/graphql"
	types "github.com/mapofzones/txs-processor/pkg/types"
)

var C *graphql.Client

func init() {
	C = graphql.NewClient(os.Getenv("GRAPHQL"))

}

func LastProcessedBlock(ctx context.Context, chainID string) (int64, error) {
	var X struct {
		Data []map[string]int64 `json:"blocks_log_hub"`
	}

	err := C.Run(ctx, toRequest(fmt.Sprintf(lastProcessedBlock, chainID)), &X)
	if err != nil {
		return -1, err
	}

	if len(X.Data) == 0 {
		return 0, nil
	}

	return X.Data[0]["last_processed_block"], nil

}

func IbcStatsExist(ctx context.Context, sourceChainID, destinationChainID string, t time.Time) (bool, error) {
	var X struct {
		Data []map[string]int64 `json:"ibc_tx_hourly_stats_hub"`
	}

	err := C.Run(ctx, toRequest(fmt.Sprintf(ibcStatsExist, sourceChainID, destinationChainID, t.Format(types.Format))), &X)

	return len(X.Data) > 0, err
}

func ConnectionIDFromChannelID(ctx context.Context, source, channelID string) (string, error) {
	var X struct {
		Data []map[string]string `json:"channels"`
	}

	err := C.Run(ctx, toRequest(fmt.Sprintf(connectionIDFromChannelID, source, channelID)), &X)

	if err != nil || len(X.Data) == 0 {
		return "", err
	}

	return X.Data[0]["connection_id"], nil
}

func ClientIDFromConnectionID(ctx context.Context, source, connectionID string) (string, error) {
	var X struct {
		Data []map[string]string `json:"connections"`
	}

	err := C.Run(ctx, toRequest(fmt.Sprintf(clientIDFromConnectionID, connectionID, source)), &X)

	if err != nil || len(X.Data) == 0 {
		return "", err
	}

	return X.Data[0]["client_id"], nil
}

func ChainIDFromClientID(ctx context.Context, source, clientID string) (string, error) {
	var X struct {
		Data []map[string]string `json:"clients"`
	}

	err := C.Run(ctx, toRequest(fmt.Sprintf(chainIDFromClientID, source, clientID)), &X)

	if err != nil || len(X.Data) == 0 {
		return "", err
	}

	return X.Data[0]["chain_id"], nil
}

func ChainIDFromConnectionID(ctx context.Context, source, connectionID string) (string, error) {
	clientID, err := ClientIDFromConnectionID(ctx, source, connectionID)
	if err != nil {
		return "", err
	}
	return ChainIDFromClientID(ctx, source, clientID)
}

func ChainIDFromChannelID(ctx context.Context, source, channelID string) (string, error) {
	connectionID, err := ConnectionIDFromChannelID(ctx, source, channelID)
	if err != nil {
		return "", err
	}
	return ChainIDFromConnectionID(ctx, source, connectionID)
}
