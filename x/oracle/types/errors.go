package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

//Exported code type numbers
var (
	ErrProphecyNotFound              = sdkerrors.Register(ModuleName, 1, "prophecy with given id not found")
	ErrMinimumConsensusNeededInvalid = sdkerrors.Register(ModuleName, 2, "minimum consensus proportion of validator staking power must be > 0 and <= 1")
	ErrNoClaims                      = sdkerrors.Register(ModuleName, 3, "cannot create prophecy without initial claim")
	ErrInvalidIdentifier             = sdkerrors.Register(ModuleName, 4, "invalid identifier provided, must be a nonempty string")
	ErrProphecyFinalized             = sdkerrors.Register(ModuleName, 5, "Prophecy already finalized")
	ErrDuplicateMessage              = sdkerrors.Register(ModuleName, 6, "Already processed message from validator for this id")
	ErrInvalidClaim                  = sdkerrors.Register(ModuleName, 7, "Claim cannot be empty string")
	ErrInvalidValidator              = sdkerrors.Register(ModuleName, 8, "Claim must be made by actively bonded validator")
	ErrInternalDB                    = sdkerrors.Register(ModuleName, 9, "Internal error serializing/deserializing prophecy: ")
)

