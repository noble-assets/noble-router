package cli_test

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/strangelove-ventures/noble/testutil/network"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/router/client/cli"
	"github.com/strangelove-ventures/noble/x/router/types"
)

func networkWithMintObjects(t *testing.T, n int) (*network.Network, []types.Mint) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		Mints := types.Mint{
			SourceDomainSender: strconv.Itoa(i),
			Nonce:              uint64(i),
		}
		nullify.Fill(&Mints)
		state.Mints = append(state.Mints, Mints)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.Mints
}

func TestShowMint(t *testing.T) {
	net, objs := networkWithMintObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc               string
		sourceDomainSender string
		nonce              string

		args []string
		err  error
		obj  types.Mint
	}{
		{
			desc:               "found",
			sourceDomainSender: objs[0].SourceDomainSender,
			nonce:              strconv.Itoa(int(objs[0].Nonce)),
			args:               common,
			obj:                objs[0],
		},
		{
			desc:               "not found",
			sourceDomainSender: "123",
			nonce:              "456",
			args:               common,
			err:                status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.sourceDomainSender,
				tc.nonce,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowMint(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetMintResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Mint)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Mint),
				)
			}
		})
	}
}