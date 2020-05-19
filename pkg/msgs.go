package processor

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	types "github.com/mapofzones/txs-processor/pkg/types"
	processor "github.com/mapofzones/txs-processor/pkg/x"
)

func processMsgs(ctx context.Context, block types.Block, p processor.Processor, msgs []sdk.Msg) error {
	// client-id -> chain-id map
	chainIDs := map[string]string{}
	// connections-id -> client-id map
	clientIDs := map[string]string{}
	// channel-id -> connection-id map
	connectionIDs := map[string]string{}
	// transfers cache, if two transfers happen in the same tx

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case tendermint.MsgCreateClient:
			// store data to function cache in order there is a connection opening in the same block
			chainIDs[msg.ClientID] = msg.Header.ChainID
			p.AddClient(msg.ClientID, msg.Header.ChainID)

		// if somebody from other chain is trying to establish connection with us
		case connection.MsgConnectionOpenTry:
			// store data locally in order if channel is created in the same block
			clientIDs[msg.ConnectionID] = msg.ConnectionID
			p.AddConnection(msg.ConnectionID, msg.ClientID)

		// if we are trying to establish connection with someone from other chain
		case connection.MsgConnectionOpenInit:
			// store data locally in order if channel is created in the same block
			clientIDs[msg.ConnectionID] = msg.ClientID
			p.AddConnection(msg.ConnectionID, msg.ClientID)

		case channel.MsgChannelOpenTry:
			// store data locally if there are transfers in the same block
			connectionIDs[msg.ChannelID] = msg.Channel.ConnectionHops[0]
			p.AddChannel(msg.ChannelID, msg.Channel.ConnectionHops[0])

		case channel.MsgChannelOpenInit:
			// store data locally if there are transfers in the same block
			connectionIDs[msg.ChannelID] = msg.Channel.ConnectionHops[0]
			p.AddChannel(msg.ChannelID, msg.Channel.ConnectionHops[0])

		// this chain is sending tokens to another chain
		case transfer.MsgTransfer:
			chainID, err := getChainID(ctx, chainIDs, clientIDs,
				connectionIDs, block.ChainID, msg.SourceChannel, p)
			if err != nil {
				return fmt.Errorf("%w: %v", ConnectionError, err)
			}

			p.AddIbcStats(types.IbcStats{
				Source:      block.ChainID,
				Destination: chainID,
				Count:       1,
				Hour:        block.T.Truncate(time.Hour),
			})

		// this chain receives tokens
		case channel.MsgPacket:
			chainID, err := getChainID(ctx, chainIDs, clientIDs,
				connectionIDs, block.ChainID, msg.Packet.DestinationChannel, p)
			if err != nil {
				return fmt.Errorf("%w: %v", ConnectionError, err)
			}

			p.AddIbcStats(types.IbcStats{
				Source:      chainID,
				Destination: block.ChainID,
				Count:       1,
				Hour:        block.T.Truncate(time.Hour),
			})

		// messages which confirm that channel is opened
		case channel.MsgChannelOpenConfirm:
			p.MarkChannelOpened(msg.ChannelID)
		case channel.MsgChannelOpenAck:
			p.MarkChannelOpened(msg.ChannelID)

		// messages which confirm that channel is closed
		case channel.MsgChannelCloseConfirm:
			p.MarkChannelClosed(msg.ChannelID)
		case channel.MsgChannelCloseInit:
			p.MarkChannelClosed(msg.ChannelID)
		}
	}

	return nil
}

// local block-origin+channel-id -> chain-id cache
var chainIDCache = map[string]string{}

func getChainID(ctx context.Context, chainIds, clientIDs,
	connectionIDs map[string]string,
	origin, channel string, p processor.Processor) (string, error) {
	// check local cache
	if id, ok := chainIDCache[origin+channel]; ok {
		return id, nil
	}

	// check if there actually is chain id in our map
	if chainID, ok := chainIds[clientIDs[connectionIDs[channel]]]; ok {
		return chainID, nil
	}

	// if there is locally cached connectionID
	if connectionID, ok := connectionIDs[channel]; ok {
		return p.ChainIDFromConnectionID(connectionID)

	}

	// nothing in cache, query db
	id, err := p.ChainIDFromChannelID(channel)
	if err != nil {
		return "", err
	}
	// update local cache
	chainIDCache[origin+channel] = id
	return id, nil
}
