package types

import (
	"errors"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	"github.com/tendermint/go-amino"
)

var Codec = setCodec()

func decodeTx(tx []byte) (sdk.Tx, error) {
	return auth.DefaultTxDecoder(Codec)(tx)
}

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
		return auth.StdTx{}, errors.New("could not decode tx")
	}
	return toStdTx(txInterface)
}

func setCodec() *amino.Codec {
	ModuleBasics := module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler, distr.ProposalHandler, upgradeclient.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
	)
	cdc := codecstd.MakeCodec(ModuleBasics)
	return cdc
}
