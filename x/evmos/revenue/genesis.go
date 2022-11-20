package revenue

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/evmos/revenue/keeper"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/evmos/revenue/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, revenue := range data.Revenues {
		contract := revenue.GetContractAddr()
		deployer := revenue.GetDeployerAddr()
		withdrawer := revenue.GetWithdrawerAddr()

		// Set initial contracts receiving transaction fees
		k.SetRevenue(ctx, revenue)
		k.SetDeployerMap(ctx, deployer, contract)

		if len(withdrawer) != 0 {
			k.SetWithdrawerMap(ctx, withdrawer, contract)
		}
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:   k.GetParams(ctx),
		Revenues: k.GetRevenues(ctx),
	}
}
