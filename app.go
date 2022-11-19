package app

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	tmjson "github.com/tendermint/tendermint/libs/json"
	authtype "github.com/cosmos/cosmos-sdk/x/auth/types"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstype "github.com/cosmos/cosmos-sdk/x/params/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtype "github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/ethbridge"
	"github.com/pumpkinzomb/cosmos-ethereum-bridge/x/oracle"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"
)

const (
	appName = "ethereum-bridge"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		params.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtype.FeeCollectorName:     nil,
		stakingtype.BondedPoolName:    {authtype.Burner, authtype.Staking},
		stakingtype.NotBondedPoolName: {authtype.Burner, authtype.Staking},
	}
)

type ethereumBridgeApp struct {
	*bam.BaseApp

	appCodec codec.Marshaler
	legacyAmino *codec.LegacyAmino
	cdc codec.Marshaler
	
	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey

	accountKeeper       authkeeper.AccountKeeper
	bankKeeper          bankkeeper.Keeper
	stakingKeeper       stakingkeeper.Keeper

	paramsKeeper paramskeeper.Keeper
	oracleKeeper oracle.Keeper

	// the module manager
	mm *module.Manager
}

// NewEthereumBridgeApp is a constructor function for ethereumBridgeApp
func NewEthereumBridgeApp(logger log.Logger, db dbm.DB) *ethereumBridgeApp {

	// First define the top level codec that will be shared by the different modules
	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder())

	keys := sdk.NewKVStoreKeys(
		authtype.StoreKey, 
		banktype.StoreKey, 
		stakingtype.StoreKey,
		paramstype.StoreKey, 
		oracle.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(paramstype.TStoreKey)

	// Here you initialize your application with the store keys it requires
	var app = &ethereumBridgeApp{
		BaseApp: 		bApp,
		appCodec: 		appCodec,
		legacyAmino:    legacyAmino,
		cdc:     		appCodec,
		keys: 			keys,
		tkeys: 			tkeys,
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstype.StoreKey], tkeys[paramstype.TStoreKey])

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtype.StoreKey],
		app.GetSubspace(authtype.ModuleName),
		authtype.ProtoBaseAccount,
		maccPerms,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktype.StoreKey],
		app.accountKeeper,
		app.GetSubspace(banktype.ModuleName),
		app.ModuleAccountAddrs(),
	)

	app.stakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtype.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
	    app.GetSubspace(stakingtype.ModuleName),
	)

	// The OracleKeeper is the Keeper from the oracle module
	// It handles interactions with the oracle store
	oracleKeeper, oracleErr := oracle.NewKeeper(
		app.stakingKeeper,
		keys[oracle.StoreKey],
		legacyAmino,
		app.GetSubspace(oracle.ModuleName),
		oracle.DefaultConsensusNeeded,
	)
	if oracleErr != nil {
		tmos.Exit(oracleErr.Error())
	}
	app.oracleKeeper = oracleKeeper

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.accountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper),
		staking.NewAppModule(appCodec, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		params.NewAppModule(app.paramsKeeper),
	)

	// The AnteHandler handles signature verification and transaction pre-processing
	

	// The app.Router is the main transaction router where each module registers its routes
	// Register the bank route here

	app.Router().
		AddRoute(sdk.NewRoute(banktype.RouterKey, bank.NewHandler(app.bankKeeper))).
		AddRoute(sdk.NewRoute(stakingtype.RouterKey, staking.NewHandler(app.stakingKeeper))).
		AddRoute(sdk.NewRoute(ethbridge.RouterKey, ethbridge.NewHandler(app.oracleKeeper, app.bankKeeper, app.cdc, ethbridge.DefaultCodespace)))

	// The app.QueryRouter is the main query router where each module registers its routes
	app.QueryRouter().
		AddRoute(authtype.QuerierRoute, authkeeper.NewQuerier(app.accountKeeper, app.legacyAmino)).
		AddRoute(stakingtype.QuerierRoute, stakingkeeper.NewQuerier(app.stakingKeeper, app.legacyAmino)).
		AddRoute(ethbridge.QuerierRoute, ethbridge.NewQuerier(app.oracleKeeper, app.legacyAmino, ethbridge.DefaultCodespace))

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(
		ante.NewAnteHandler(
			app.accountKeeper, 
			app.bankKeeper, 
			ante.DefaultSigVerificationGasConsumer,
			encodingConfig.TxConfig.SignModeHandler(),
		),
	)
	app.SetEndBlocker(app.EndBlocker)

	if err := app.LoadLatestVersion(); err != nil {
		tmos.Exit(err.Error())
	}

	return app
}


// MakeCodecs constructs the *std.Codec and *codec.LegacyAmino instances used by
// simapp. It is useful for tests and clients who do not want to construct the
// full simapp
func MakeCodecs() (codec.Marshaler, *codec.LegacyAmino) {
	config := simapp.MakeTestEncodingConfig()
	return config.Marshaler, config.Amino
}


// BeginBlocker application updates every begin block
func (app *ethereumBridgeApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *ethereumBridgeApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *ethereumBridgeApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// load a particular height
func (app *ethereumBridgeApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *ethereumBridgeApp) GetSubspace(moduleName string) paramstype.Subspace {
	subspace, _ := app.paramsKeeper.GetSubspace(moduleName)
	return subspace
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *ethereumBridgeApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtype.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryMarshaler, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtype.ModuleName)
	paramsKeeper.Subspace(banktype.ModuleName)
	paramsKeeper.Subspace(stakingtype.ModuleName)

	return paramsKeeper
}

