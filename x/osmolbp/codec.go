package osmolbp

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&api.MsgCreateLBP{},
		&api.MsgExitLBP{},
		&api.MsgFinalizeLBP{},
		&api.MsgSubscribe{},
		&api.MsgWithdraw{},
	)

	// registry.RegisterInterface()

	msgservice.RegisterMsgServiceDesc(registry, api.MsgServiceDesc())
}
