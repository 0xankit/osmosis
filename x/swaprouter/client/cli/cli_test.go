package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app"
	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/cli"
	swaprouterqueryproto "github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/queryproto"
	swaproutertestutil "github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/testutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = app.DefaultConfig()
	s.cfg.GenesisState = swaproutertestutil.UpdateTxFeeDenom(s.cfg.Codec, s.cfg.BondDenom)

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new pool
	_, err = swaproutertestutil.MsgCreatePool(s.T(), val.ClientCtx, val.Address, "5stake,5node0token", "100stake,100node0token", "0.01", "0.01", "")
	s.Require().NoError(err)

	s.T().Log("test")

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {

	// TODO: re-enable this once swaprouter is fully merged.
	t.SkipNow()

	suite.Run(t, new(IntegrationTestSuite))
}

func (s IntegrationTestSuite) TestNewSwapExactAmountOutCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountOut",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name string
		args []string

		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"swap exact amount out", // osmosisd tx swaprouter swap-exact-amount-out 10stake 20 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "20",
				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSwapExactAmountOutCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

// func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountIn() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// 			[]string{
// 				"1",
// 				"cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq",
// 				"10.0stake",
// 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdEstimateSwapExactAmountIn()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QuerySwapExactAmountInResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestGetCmdEstimateSwapExactAmountOut() {
// 	val := s.network.Validators[0]

// 	testCases := []struct {
// 		name      string
// 		args      []string
// 		expectErr bool
// 	}{
// 		{
// 			"query pool estimate swap exact amount in", // osmosisd query gamm estimate-swap-exact-amount-in 1 cosmos1n8skk06h3kyh550ad9qketlfhc2l5dsdevd3hq 10.0stake --swap-route-pool-ids=1 --swap-route-denoms=node0token
// 			[]string{
// 				"1",
// 				val.Address.String(),
// 				"10.0stake",
// 				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
// 				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
// 				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			cmd := cli.GetCmdEstimateSwapExactAmountOut()
// 			clientCtx := val.ClientCtx

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
// 			if tc.expectErr {
// 				s.Require().Error(err)
// 			} else {
// 				resp := types.QuerySwapExactAmountOutResponse{}
// 				s.Require().NoError(err, out.String())
// 				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }

func (s IntegrationTestSuite) TestNewSwapExactAmountInCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewSwapExactAmountIn",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 20000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name string
		args []string

		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"swap exact amount in", // osmosisd tx swaprouter swap-exact-amount-in 10stake 3 --swap-route-pool-ids=1 --swap-route-denoms=node0token --from=validator --keyring-backend=test --chain-id=testing --yes
			[]string{
				"10stake", "3",
				fmt.Sprintf("--%s=%d", cli.FlagSwapRoutePoolIds, 1),
				fmt.Sprintf("--%s=%s", cli.FlagSwapRouteDenoms, "node0token"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSwapExactAmountInCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdNumPools() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query num-pools",
			[]string{
				fmt.Sprintf("--%s=%s", tmcli.OutputFlag, "json"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdNumPools() // osmosisd query swaprouter num-pools
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				resp := swaprouterqueryproto.NumPoolsResponse{}
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())

				s.Require().Greater(resp.NumPools, uint64(0), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewCreatePoolCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewCreatePoolAddr",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 200000000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		osmoutils.DefaultFeeString(s.cfg),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		json         string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"one token pair pool",
			fmt.Sprintf(`
			{
			  "%s": "1node0token",
			  "%s": "100node0token",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 4,
		},
		{
			"two tokens pair pool",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"change order of json fields",
			fmt.Sprintf(`
			{
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "1node0token,3stake",
			  "%s": "0.001"
			}
			`, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileWeights, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 0,
		},
		{ // --record-tokens=100.0stake2 --record-tokens=100.0stake --record-tokens-weight=5 --record-tokens-weight=5 --swap-fee=0.01 --exit-fee=0.01 --from=validator --keyring-backend=test --chain-id=testing --yes
			"three tokens pair pool - insufficient balance check",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,1stake,2btc",
			  "%s": "100node0token,100stake,100btc",
			  "%s": "0.001",
			  "%s": "0.001"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			false, &sdk.TxResponse{}, 5,
		},
		{
			"future governor address",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": "0.001",
			  "%s": "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
			false, &sdk.TxResponse{}, 0,
		},
		// Due to CI time concerns, we leave these CLI tests commented out, and instead guaranteed via
		// the logic tests.
		// {
		// 	"future governor time",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "2h"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	false, &sdk.TxResponse{}, 0,
		// },
		// {
		// 	"future governor token + time",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "token,1000h"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	false, &sdk.TxResponse{}, 0,
		// },
		// {
		// 	"invalid future governor",
		// 	fmt.Sprintf(`
		// 	{
		// 	  "%s": "1node0token,3stake",
		// 	  "%s": "100node0token,100stake",
		// 	  "%s": "0.001",
		// 	  "%s": "0.001",
		// 	  "%s": "validdenom,invalidtime"
		// 	}
		// 	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee, cli.PoolFileFutureGovernor),
		// 	true, &sdk.TxResponse{}, 7,
		// },
		{
			"not valid json",
			"bad json",
			true, &sdk.TxResponse{}, 0,
		},
		{
			"bad pool json - missing quotes around exit fee",
			fmt.Sprintf(`
			{
			  "%s": "1node0token,3stake",
			  "%s": "100node0token,100stake",
			  "%s": "0.001",
			  "%s": 0.001
			}
	`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"empty pool json",
			"", true, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params",
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "864h",
						"%s": "2node0token,1stake",
						"%s": "2006-01-02T15:04:05Z"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileDuration, cli.PoolFileTargetPoolWeights, cli.PoolFileStartTime,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params - no start time",
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "864h",
						"%s": "2node0token,1stake"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileDuration, cli.PoolFileTargetPoolWeights,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"empty smooth change params",
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params wrong type",
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": "invalid string"
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams,
			),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"smooth change params missing duration",
			fmt.Sprintf(`
				{
					"%s": "1node0token,3stake",
					"%s": "100node0token,100stake",
					"%s": "0.001",
					"%s": "0.001",
					"%s": {
						"%s": "2node0token,1stake"
					}
				}
				`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee,
				cli.PoolFileSmoothWeightChangeParams, cli.PoolFileTargetPoolWeights,
			),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"unknown fields in json",
			fmt.Sprintf(`
			{
			  "%s": "1node0token",
			  "%s": "100node0token",
			  "%s": "0.001",
			  "%s": "0.001"
			  "unknown": true,
			}
			`, cli.PoolFileWeights, cli.PoolFileInitialDeposit, cli.PoolFileSwapFee, cli.PoolFileExitFee),
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreatePoolCmd()
			clientCtx := val.ClientCtx

			jsonFile := testutil.WriteToNewTempFile(s.T(), tc.json)

			args := []string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolFile, jsonFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				osmoutils.DefaultFeeString(s.cfg),
				fmt.Sprintf("--%s=%s", flags.FlagGas, fmt.Sprint(400000)),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				s.Require().NoError(err, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}
