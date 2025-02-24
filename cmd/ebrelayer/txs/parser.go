package txs

// --------------------------------------------------------
//      Parser
//
//      Parses structs containing event information into
//      unsigned transactions for validators to sign, then
//      relays the data packets as transactions on the
//      Cosmos Bridge.
// --------------------------------------------------------

import (
  "strings"
  "strconv"
  "fmt"

  sdk "github.com/cosmos/cosmos-sdk/types"
  "github.com/pumpkinzomb/cosmos-ethereum-bridge/cmd/ebrelayer/events"
  "github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge/types"
)

func ParsePayload(validator sdk.AccAddress, event *events.LockEvent) (types.EthBridgeClaim, error) {
  
  witnessClaim := types.EthBridgeClaim{}

  // Nonce type casting (*big.Int -> int)
  nonce, nonceErr := strconv.Atoi(event.Nonce.String())
  if nonceErr != nil {
    fmt.Errorf("%s", nonceErr)
  }
  witnessClaim.Nonce = nonce

  // EthereumSender type casting (address.common -> string)
  witnessClaim.EthereumSender = event.From.Hex()

  // CosmosReceiver type casting (bytes[] -> sdk.AccAddress)
  recipient, recipientErr := sdk.AccAddressFromBech32(string(event.To[:]))
  if recipientErr != nil {
    fmt.Errorf("%s", recipientErr)
  }
  witnessClaim.CosmosReceiver = recipient

  // Validator is already the correct type (sdk.AccAddress)
  witnessClaim.Validator = validator

  // Amount type casting (*big.Int -> sdk.Coins)
  ethereumCoin := []string {event.Value.String(),"ethereum"}
  weiAmount, coinErr := sdk.ParseCoins(strings.Join(ethereumCoin, ""))
  if coinErr != nil {
    fmt.Errorf("%s", coinErr)
  }
  witnessClaim.Amount = weiAmount

  return witnessClaim, nil
}
