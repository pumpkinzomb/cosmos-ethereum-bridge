package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	bnkkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	oracletypes "github.com/pumpkinzomb/cosmos-ethereum-bridge/x/oracle/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper  bnkkeeper.Keeper
	stakeKeeper stakingkeeper.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.LegacyAmino // The wire codec for binary encoding/decoding.

	paramSubspace paramtypes.Subspace

	consensusNeeded float64
}

// NewKeeper creates new instances of the oracle Keeper
func NewKeeper(stakeKeeper stakingkeeper.Keeper, storeKey sdk.StoreKey, cdc *codec.LegacyAmino, paramSubspace paramtypes.Subspace, consensusNeeded float64) (Keeper, error) {
	if consensusNeeded <= 0 || consensusNeeded > 1 {
		return Keeper{}, oracletypes.ErrMinimumConsensusNeededInvalid
	}
	return Keeper{
		stakeKeeper:     stakeKeeper,
		storeKey:        storeKey,
		cdc:             cdc,
		paramSubspace:   paramSubspace,
		consensusNeeded: consensusNeeded,
	}, nil
}

// Codespace returns the codespace
func (k Keeper) Codespace() paramtypes.Subspace {
	return k.paramSubspace
}

// GetProphecy gets the entire prophecy data struct for a given id
func (k Keeper) GetProphecy(ctx sdk.Context, id string) (oracletypes.Prophecy, error) {
	if id == "" {
		return oracletypes.NewEmptyProphecy(), oracletypes.ErrInvalidIdentifier
	}
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(id)) {
		return oracletypes.NewEmptyProphecy(), oracletypes.ErrProphecyNotFound
	}

	var dbProphecy oracletypes.DBProphecy
	
	bz := store.Get([]byte(id))
	k.cdc.MustUnmarshalJSON(bz, &dbProphecy)

	deSerializedProphecy, err := dbProphecy.DeserializeFromDB()
	if err != nil {
		return oracletypes.NewEmptyProphecy(), sdkerrors.Wrap(oracletypes.ErrInternalDB, err.Error())
	}
	return deSerializedProphecy, nil
}

// saveProphecy saves a prophecy with an initial claim
func (k Keeper) saveProphecy(ctx sdk.Context, prophecy oracletypes.Prophecy) error {
	if prophecy.ID == "" {
		return oracletypes.ErrInvalidIdentifier
	}
	if len(prophecy.ClaimValidators) <= 0 {
		return oracletypes.ErrNoClaims
	}
	store := ctx.KVStore(k.storeKey)
	serializedProphecy, err := prophecy.SerializeForDB()
	if err != nil {
		return sdkerrors.Wrap(oracletypes.ErrInternalDB, err.Error())
	}
	store.Set([]byte(prophecy.ID), k.cdc.MustMarshalJSON(serializedProphecy))
	return nil
}

func (k Keeper) ProcessClaim(ctx sdk.Context, id string, validator sdk.ValAddress, claim string) (oracletypes.Status, error) {
	activeValidator := k.checkActiveValidator(ctx, validator)
	if !activeValidator {
		return oracletypes.Status{}, oracletypes.ErrInvalidValidator
	}
	if claim == "" {
		return oracletypes.Status{}, oracletypes.ErrInvalidClaim
	}
	prophecy, err := k.GetProphecy(ctx, id)
	if err == nil {
		if prophecy.Status.StatusText == oracletypes.SuccessStatusText || prophecy.Status.StatusText == oracletypes.FailedStatusText {
			return oracletypes.Status{}, oracletypes.ErrProphecyFinalized
		}
		if prophecy.ValidatorClaims[validator.String()] != "" {
			return oracletypes.Status{}, oracletypes.ErrDuplicateMessage
		}
		prophecy.AddClaim(validator, claim)
	} else {
		if err != oracletypes.ErrProphecyNotFound {
			return oracletypes.Status{}, err
		}
		prophecy = oracletypes.NewProphecy(id)
		prophecy.AddClaim(validator, claim)
	}
	prophecy = k.processCompletion(ctx, prophecy)
	err = k.saveProphecy(ctx, prophecy)
	if err != nil {
		return oracletypes.Status{}, err
	}
	return prophecy.Status, nil
}

func (k Keeper) checkActiveValidator(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	validator, found := k.stakeKeeper.GetValidator(ctx, validatorAddress)
	if !found {
		return false
	}
	bondStatus := validator.GetStatus()
	if bondStatus != stakingtypes.Bonded {
		return false
	}
	return true
}

func (k Keeper) processCompletion(ctx sdk.Context, prophecy oracletypes.Prophecy) oracletypes.Prophecy {
	highestClaim, highestClaimPower, totalClaimsPower := prophecy.FindHighestClaim(ctx, k.stakeKeeper)
	totalPower := k.stakeKeeper.GetLastTotalPower(ctx)
	highestConsensusRatio := float64(highestClaimPower) / float64(totalPower.Int64())
	remainingPossibleClaimPower := totalPower.Int64() - totalClaimsPower
	highestPossibleClaimPower := highestClaimPower + remainingPossibleClaimPower
	highestPossibleConsensusRatio := float64(highestPossibleClaimPower) / float64(totalPower.Int64())
	if highestConsensusRatio >= k.consensusNeeded {
		prophecy.Status.StatusText = oracletypes.SuccessStatusText
		prophecy.Status.FinalClaim = highestClaim
	} else if highestPossibleConsensusRatio <= k.consensusNeeded {
		prophecy.Status.StatusText = oracletypes.FailedStatusText
	}
	return prophecy
}
