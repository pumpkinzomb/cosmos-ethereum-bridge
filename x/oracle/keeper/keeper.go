package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	bnkkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stkkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	oracleerrors "github.com/pumpkinzomb/cosmos-ethereum-bridge/x/oracle/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper  bnkkeeper.Keeper
	stakeKeeper stkkeeper.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.BinaryMarshaler // The wire codec for binary encoding/decoding.

	paramSubspace paramtypes.Subspace

	consensusNeeded float64
}

// NewKeeper creates new instances of the oracle Keeper
func NewKeeper(stakeKeeper stkkeeper.Keeper, storeKey sdk.StoreKey, cdc *codec.BinaryMarshaler, paramSubspace paramtypes.Subspace, consensusNeeded float64) (Keeper, error) {
	if consensusNeeded <= 0 || consensusNeeded > 1 {
		return Keeper{}, oracleerrors.ErrMinimumConsensusNeededInvalid
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
func (k Keeper) GetProphecy(ctx sdk.Context, id string) (types.Prophecy, error) {
	if id == "" {
		return types.NewEmptyProphecy(), oracleerrors.ErrInvalidIdentifier
	}
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(id)) {
		return types.NewEmptyProphecy(), oracleerrors.ErrProphecyNotFound
	}
	bz := store.Get([]byte(id))
	var dbProphecy types.DBProphecy
	k.cdc.MustUnmarshalBinaryBare(bz, &dbProphecy)

	deSerializedProphecy, err := dbProphecy.DeserializeFromDB()
	if err != nil {
		return types.NewEmptyProphecy(), sdkerrors.Wrap(oracleerrors.ErrInternalDB, err)
	}
	return deSerializedProphecy, nil
}

// saveProphecy saves a prophecy with an initial claim
func (k Keeper) saveProphecy(ctx sdk.Context, prophecy types.Prophecy) error {
	if prophecy.ID == "" {
		return oracleerrors.ErrInvalidIdentifier
	}
	if len(prophecy.ClaimValidators) <= 0 {
		return oracleerrors.ErrNoClaims
	}
	store := ctx.KVStore(k.storeKey)
	serializedProphecy, err := prophecy.SerializeForDB()
	if err != nil {
		return sdkerrors.Wrap(oracleerrors.ErrInternalDB, err)
	}
	store.Set([]byte(prophecy.ID), k.cdc.MustMarshalBinaryBare(serializedProphecy))
	return nil
}

func (k Keeper) ProcessClaim(ctx sdk.Context, id string, validator sdk.ValAddress, claim string) (types.Status, error) {
	activeValidator := k.checkActiveValidator(ctx, validator)
	if !activeValidator {
		return types.Status{}, oracleerrors.ErrInvalidValidator
	}
	if claim == "" {
		return types.Status{}, oracleerrors.ErrInvalidClaim
	}
	prophecy, err := k.GetProphecy(ctx, id)
	if err == nil {
		if prophecy.Status.StatusText == types.SuccessStatusText || prophecy.Status.StatusText == types.FailedStatusText {
			return types.Status{}, oracleerrors.ErrProphecyFinalized
		}
		if prophecy.ValidatorClaims[validator.String()] != "" {
			return types.Status{}, oracleerrors.ErrDuplicateMessage
		}
		prophecy.AddClaim(validator, claim)
	} else {
		if err.Code() != types.CodeProphecyNotFound {
			return types.Status{}, err
		}
		prophecy = types.NewProphecy(id)
		prophecy.AddClaim(validator, claim)
	}
	prophecy = k.processCompletion(ctx, prophecy)
	err = k.saveProphecy(ctx, prophecy)
	if err != nil {
		return types.Status{}, err
	}
	return prophecy.Status, nil
}

func (k Keeper) checkActiveValidator(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	validator, found := k.stakeKeeper.GetValidator(ctx, validatorAddress)
	if !found {
		return false
	}
	bondStatus := validator.GetStatus()
	if bondStatus != sdk.Bonded {
		return false
	}
	return true
}

func (k Keeper) processCompletion(ctx sdk.Context, prophecy types.Prophecy) types.Prophecy {
	highestClaim, highestClaimPower, totalClaimsPower := prophecy.FindHighestClaim(ctx, k.stakeKeeper)
	totalPower := k.stakeKeeper.GetLastTotalPower(ctx)
	highestConsensusRatio := float64(highestClaimPower) / float64(totalPower.Int64())
	remainingPossibleClaimPower := totalPower.Int64() - totalClaimsPower
	highestPossibleClaimPower := highestClaimPower + remainingPossibleClaimPower
	highestPossibleConsensusRatio := float64(highestPossibleClaimPower) / float64(totalPower.Int64())
	if highestConsensusRatio >= k.consensusNeeded {
		prophecy.Status.StatusText = types.SuccessStatusText
		prophecy.Status.FinalClaim = highestClaim
	} else if highestPossibleConsensusRatio <= k.consensusNeeded {
		prophecy.Status.StatusText = types.FailedStatusText
	}
	return prophecy
}
