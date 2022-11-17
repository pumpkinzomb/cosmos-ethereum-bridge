package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetCmdMakeEthBridgeClaim is the CLI command for making a claim on an ethereum prophecy
func GetCmdMakeEthBridgeClaim(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "make-claim nonce ethereum-sender-address cosmos-receiver-address validator-address amount",
		Short: "make a claim on an ethereum prophecy",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			nonce, stringError := strconv.Atoi(args[0])
			if stringError != nil {
				return stringError
			}

			ethereumSender := args[1]
			cosmosReceiver, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			validator, err := sdk.AccAddressFromBech32(args[3])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(args[4])
			if err != nil {
				return err
			}

			ethBridgeClaim := types.NewEthBridgeClaim(nonce, ethereumSender, cosmosReceiver, validator, amount)
			msg := types.NewMsgMakeEthBridgeClaim(ethBridgeClaim)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
}
