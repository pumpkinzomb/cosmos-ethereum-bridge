package ethbridge

import (
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/querier"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/types"
)

var (
	NewMsgMakeEthBridgeClaim = types.NewMsgMakeEthBridgeClaim
	NewEthBridgeClaim        = types.NewEthBridgeClaim

	NewQueryEthProphecyParams = types.NewQueryEthProphecyParams

	ErrInvalidEthNonce = types.ErrInvalidEthNonce

	RegisterCodec = types.RegisterCodec

	NewQuerier = querier.NewQuerier
)

const (
	StoreKey         = types.StoreKey
	QuerierRoute     = types.QuerierRoute
	RouterKey        = types.RouterKey
	DefaultCodespace = types.DefaultCodespace

	QueryEthProphecy = querier.QueryEthProphecy
)
