package types

import (
	context "context"
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/regen-network/cosmos-proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgMakeEthBridgeClaim struct {
	EthBridgeClaim  *types.Any `protobuf:"bytes,1,opt,name=ethBridgeClaim,proto3" json:"eth_bridge_claim"`
}

type MsgMakeEthBridgeClaimResponse struct {
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// SubmitEvidence submits an arbitrary Evidence of misbehavior such as equivocation or
	// counterfactual signing.
	
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) MakeEthBridgeClaim(ctx context.Context, in *MsgMakeEthBridgeClaim, opts ...grpc.CallOption) (*MsgMakeEthBridgeClaimResponse, error) {
	out := new(MsgMakeEthBridgeClaimResponse)
	err := c.cc.Invoke(ctx, "/cosmos-ethereum-bridge.oracle.v1beta1.Msg/SubmitEvidence", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}


// MsgServer is the server API for Msg service.
type MsgServer interface {
	// Send defines a method for sending coins from one account to another account.
	MakeEthBridgeClaim(context.Context, *MsgMakeEthBridgeClaim) (*MsgMakeEthBridgeClaimResponse, error)
}
