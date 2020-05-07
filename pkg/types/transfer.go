package types

import (
	"time"

	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
)

type Transfer struct {
	Hash      string
	Zone      string
	Sender    string
	Recipient string
	Quantity  int64
	Token     string
	Timestamp Timestamp
	Type      string
	Matched   bool
}

// transfer types
const (
	Send    = "send"
	Receive = "receive"
)

func FromMsgTransfer(msg transfer.MsgTransfer, hash, zone string, t time.Time) Transfer {
	return Transfer{
		Sender:    msg.Sender.String(),
		Recipient: msg.Receiver,
		Quantity:  msg.Amount[0].Amount.Int64(),
		Token:     msg.Amount[0].Denom,
		Type:      Send,
		Hash:      hash,
		Zone:      zone,
		Timestamp: ToTimestamp(t),
	}
}

func FromMsgPacket(data transfer.FungibleTokenPacketData, hash, zone string, t time.Time) Transfer {
	return Transfer{
		Hash:      hash,
		Zone:      zone,
		Timestamp: ToTimestamp(t),
		Sender:    data.Sender,
		Recipient: data.Receiver,
		Quantity:  data.Amount[0].Amount.Int64(),
		Token:     data.Amount[0].Denom,
		Type:      Receive,
	}
}

func OppositeType(t string) string {
	switch t {
	case "send":
		return "receive"
	case "receive":
		return "send"
	default:
		panic("invalid transfer message type")
	}
}
