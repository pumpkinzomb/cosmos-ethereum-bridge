package app

import (
	"encoding/json"
	"log"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtype "github.com/cosmos/cosmos-sdk/x/auth/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtype "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// export the state of ebd for a genesis file
func (app *ethereumBridgeApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string) (
	appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx, jailWhiteList)
	}

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	appendAccount := func(acc authtype.AccountI) (stop bool) {
		account := NewGenesisAccountI(acc)
		accounts = append(accounts, account)
		return false
	}
	app.accountKeeper.IterateAccounts(ctx, appendAccount)

	genState := NewGenesisState(
		accounts,
		auth.ExportGenesis(ctx, app.accountKeeper),
		app.bankKeeper.ExportGenesis(ctx),
		staking.ExportGenesis(ctx, app.stakingKeeper),
	)
	appState, err = codec.MarshalJSONIndent(app.legacyAmino, genState)
	if err != nil {
		return nil, nil, err
	}
	validators, err = staking.WriteValidators(ctx, app.stakingKeeper)
	if err != nil {
		return nil, nil, err
	}
	return appState, validators, nil
}

// prepare for fresh start at zero height
func (app *ethereumBridgeApp) prepForZeroHeightGenesis(ctx sdk.Context, jailWhiteList []string) {
	applyWhiteList := false

	//Check if there is a whitelist
	if len(jailWhiteList) > 0 {
		applyWhiteList = true
	}

	whiteListMap := make(map[string]bool)

	for _, addr := range jailWhiteList {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			log.Fatal(err)
		}
		whiteListMap[addr] = true
	}

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	app.stakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtype.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetRedelegation(ctx, red)
		return false
	})

	// iterate through unbonding delegations, reset creation height
	app.stakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtype.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	store := ctx.KVStore(app.keys[stakingtype.StoreKey])
	iter := sdk.KVStoreReversePrefixIterator(store, stakingtype.ValidatorsKey)
	counter := int16(0)

	var valConsAddrs []sdk.ConsAddress
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Key()[1:])
		validator, found := app.stakingKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		
		consAddr, err := validator.GetConsAddr()
		if err != nil {
			panic(err)
		}
		valConsAddrs = append(valConsAddrs, consAddr)
		if applyWhiteList && !whiteListMap[addr.String()] {
			validator.Jailed = true
		}

		app.stakingKeeper.SetValidator(ctx, validator)
		counter++
	}

	iter.Close()

	_, err := app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}
}
