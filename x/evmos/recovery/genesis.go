package recovery

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/evmos/recovery/keeper"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/evmos/recovery/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
