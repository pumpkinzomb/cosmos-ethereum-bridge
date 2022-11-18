package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultCodespace = "ethbridge"
)

var (
	ErrInternal          = sdkerrors.Register(ModuleName, 0, "internal")
	ErrInvalidEthNonce   = sdkerrors.Register(ModuleName, 1, "invalid ethereum nonce provided, must be >= 0")
	ErrInvalidEthAddress = sdkerrors.Register(ModuleName, 2, "invalid ethereum address provided, must be a valid hex-encoded Ethereum address")
)
