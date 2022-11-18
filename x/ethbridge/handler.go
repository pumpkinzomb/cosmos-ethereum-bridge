package ethbridge

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/common"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/types"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/oracle"
)

// NewHandler returns a handler for "ethbridge" type messages.
func NewHandler(oracleKeeper oracle.Keeper, bankKeeper bankkeeper.Keeper, cdc *codec.LegacyAmino, codespace string) sdk.Handler {

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		
		switch msg := msg.(type) {
		case MsgMakeEthBridgeClaim:
			return handleMsgMakeEthBridgeClaim(ctx, cdc, oracleKeeper, bankKeeper, msg, codespace)
		default:
			errMsg := fmt.Sprintf("Unrecognized ethbridge message type: %v", msg.Type())
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// Handle a message to make a bridge claim
func handleMsgMakeEthBridgeClaim(ctx sdk.Context, cdc *codec.LegacyAmino, oracleKeeper oracle.Keeper, bankKeeper bankkeeper.Keeper, msg MsgMakeEthBridgeClaim, codespace string) sdk.Result {
	if msg.CosmosReceiver.Empty() {
		return sdk.ErrInvalidAddress(msg.CosmosReceiver.String()).Result()
	}
	if msg.Nonce < 0 {
		return types.ErrInvalidEthNonce.Result()
	}
	if !common.IsValidEthAddress(msg.EthereumSender) {
		return types.ErrInvalidEthAddress.Result()
	}
	oracleId, validator, claimString := types.CreateOracleClaimFromEthClaim(cdc, msg.EthBridgeClaim)
	status, err := oracleKeeper.ProcessClaim(ctx, oracleId, validator, claimString)
	if err != nil {
		return nil, err
	}
	if status.StatusText == oracle.SuccessStatus {
		err = processSuccessfulClaim(ctx, bankKeeper, status.FinalClaim)
		if err != nil {
			return nil, err
		}
	}
	return sdk.Result{Log: status.StatusText}
}

func processSuccessfulClaim(ctx sdk.Context, bankKeeper bankkeeper.Keeper, claim string) error {
	oracleClaim, err := types.CreateOracleClaimFromOracleString(claim)
	if err != nil {
		return err
	}
	receiverAddress := oracleClaim.CosmosReceiver
	err = bankKeeper.AddCoins(ctx, receiverAddress, oracleClaim.Amount)
	if err != nil {
		return err
	}
	return nil
}
