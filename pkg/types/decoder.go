package types

import (
	"encoding/base64"
	"errors"

	simapp "github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
)

var Codec = setCodec()

func toStdTx(tx sdk.Tx) (auth.StdTx, error) {
	stdTx, ok := tx.(auth.StdTx)
	if !ok {
		return auth.StdTx{}, errors.New("tx is not of type StdTx")
	}
	return stdTx, nil
}

// Decode accept tx bytes and transforms them to cosmos std tx
func Decode(tx []byte) (auth.StdTx, error) {
	txInterface, err := auth.DefaultTxDecoder(Codec)(tx)
	if err != nil {
		return auth.StdTx{}, errors.New("could not decode tx: " + base64.StdEncoding.EncodeToString(tx))
	}
	return toStdTx(txInterface)
}

func setCodec() *amino.Codec {
	_, codec := simapp.MakeCodecs()

	return codec
}
