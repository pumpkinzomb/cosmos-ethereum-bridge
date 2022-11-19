package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/oracle/types"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/common"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the oracle MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// SubmitEvidence implements the MsgServer.SubmitEvidence method.
func (ms msgServer) MakeEthBridgeClaim(oracleCtx context.Context, msg *types.MsgMakeEthBridgeClaim) (*types.MsgMakeEthBridgeClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(oracleCtx)

	if msg.CosmosReceiver.Empty() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.CosmosReceiver.String())
	}
	if msg.Nonce < 0 {
		return nil, ErrInvalidEthNonce
	}
	if !common.IsValidEthAddress(msg.EthereumSender) {
		return nil, sdkerrors.Wrap(ErrInvalidEthAddress)
	}

	oracleId, validator, claimString := CreateOracleClaimFromEthClaim(cdc, msg.EthBridgeClaim)
	status, err := oracleKeeper.ProcessClaim(ctx, oracleId, validator, claimString)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSubmitter().String()),
		),
	)

	return &types.MsgMakeEthBridgeClaimResponse{
		Hash: evidence.Hash(),
	}, nil
}