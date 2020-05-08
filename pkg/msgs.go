package processor

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/mapofzones/txs-processor/pkg/builder"
	"github.com/mapofzones/txs-processor/pkg/graphql"
	types "github.com/mapofzones/txs-processor/pkg/types"
)

func processMsgs(ctx context.Context, block types.Block, builder *builder.MutationBuilder, msgs []sdk.Msg) error {
	// client-id -> chain-id map
	chainIDs := map[string]string{}
	// connections-id -> client-id map
	clientIDs := map[string]string{}
	// channel-id -> connection-id map
	connectionIDs := map[string]string{}
	// transfers cache, if two transfers happen in the same tx
	transfers := map[string]bool{}

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case tendermint.MsgCreateClient:
			// store data to function cache in order there is a connection opening in the same block
			chainIDs[msg.ClientID] = msg.Header.ChainID
			builder.InsertClient(block.ChainID, msg.ClientID, msg.Header.ChainID)

		// if somebody from other chain is trying to establish connection with us
		case connection.MsgConnectionOpenTry:
			// store data locally in order if channel is created in the same block
			clientIDs[msg.ConnectionID] = msg.ConnectionID
			builder.InsertConnection(block.ChainID, msg.ConnectionID, msg.ClientID)

		// if we are trying to establish connection with someone from other chain
		case connection.MsgConnectionOpenInit:
			// store data locally in order if channel is created in the same block
			clientIDs[msg.ConnectionID] = msg.ClientID
			builder.InsertConnection(block.ChainID, msg.ConnectionID, msg.ClientID)

		case channel.MsgChannelOpenTry:
			// store data locally if there are transfers in the same block
			connectionIDs[msg.ChannelID] = msg.Channel.ConnectionHops[0]
			builder.InsertChannel(block.ChainID, msg.ChannelID, msg.Channel.ConnectionHops[0])

		case channel.MsgChannelOpenInit:
			// store data locally if there are transfers in the same block
			connectionIDs[msg.ChannelID] = msg.Channel.ConnectionHops[0]
			builder.InsertChannel(block.ChainID, msg.ChannelID, msg.Channel.ConnectionHops[0])

		// this chain is sending tokens to another chain
		case transfer.MsgTransfer:
			chainID, err := getChainID(ctx, chainIDs, clientIDs, connectionIDs, block.ChainID, msg.SourceChannel)
			if err != nil {
				return err
			}

			err = upsertIbcStats(ctx, builder, transfers, block.ChainID, chainID, block.T)
			if err != nil {
				return err
			}

		// this chain receives tokens
		case channel.MsgPacket:
			chainID, err := getChainID(ctx, chainIDs, clientIDs, connectionIDs, block.ChainID, msg.DestinationChannel)
			if err != nil {
				return err
			}
			err = upsertIbcStats(ctx, builder, transfers, chainID, block.ChainID, block.T)
			if err != nil {
				return err
			}

		// messages which confirm that channel is opened
		case channel.MsgChannelOpenConfirm:
			builder.MarkChannelOpened(block.ChainID, msg.ChannelID)
		case channel.MsgChannelOpenAck:
			builder.MarkChannelOpened(block.ChainID, msg.ChannelID)

		// messages which confirm that channel is closed
		case channel.MsgChannelCloseConfirm:
			builder.MarkChannelClosed(block.ChainID, msg.ChannelID)
		case channel.MsgChannelCloseInit:
			builder.MarkChannelClosed(block.ChainID, msg.ChannelID)
		}
	}

	return nil
}

// local channel-id -> chain-id cache
var chainIDCache = map[string]string{}

func getChainID(ctx context.Context, chainIds, clientIDs, connectionIDs map[string]string, source, channel string) (string, error) {
	// check local cache
	if id, ok := chainIDCache[channel]; ok {
		return id, nil
	}

	// check if there actually is chain id in our map, probably impossible
	if chainID, ok := chainIds[clientIDs[connectionIDs[channel]]]; ok {
		return chainID, nil
	}

	// if there is locally cached connectionID
	if connectionID, ok := connectionIDs[channel]; ok {
		return graphql.ClientIDFromConnectionID(ctx, source, connectionID)

	}

	// nothing in cache, query db
	id, err := graphql.ChainIDFromChannelID(ctx, source, channel)
	if err != nil {
		return "", err
	}
	// update local cache
	chainIDCache[channel] = id
	return id, nil
}

func upsertIbcStats(ctx context.Context, b *builder.MutationBuilder, transfers map[string]bool, source, chainID string, t time.Time) error {
	b.AddZone(source)
	b.AddZone(chainID)

	// check and update local cache
	exists := transfers[source+chainID]
	transfers[source+chainID] = true

	// check db
	if !exists {
		dbExists, err := graphql.IbcStatsExist(ctx, source, chainID, t.Truncate(time.Hour))
		if err != nil {
			return err
		}
		exists = exists || dbExists
	}

	// upsert
	if exists {
		b.UpdateIbcStats(types.IbcStats{
			Source:      source,
			Destination: chainID,
			Count:       1,
			Hour:        t.Truncate(time.Hour),
		})
	} else {
		b.CreateIbcStats(types.IbcStats{
			Source:      source,
			Destination: chainID,
			Count:       1,
			Hour:        t.Truncate(time.Hour),
		})
	}

	return nil
}
