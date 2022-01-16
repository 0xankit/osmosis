package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

func newLBP(treasury string, id uint64, tokenIn, tokenOut string, start, end time.Time, totalOut sdk.Int) api.LBP {
	zero := sdk.ZeroInt()
	return api.LBP{
		Treasury:  treasury,
		Id:        id,
		TokenOut:  tokenOut,
		TokenIn:   tokenIn,
		StartTime: start,
		EndTime:   end,

		OutRemaining: totalOut,
		OutSold:      zero,
		OutPerShare:  zero,

		Staked: zero,
		Income: zero,
		Shares: zero,

		Round:    0,
		EndRound: currentRound(start, end, end),
	}
}

func newUserPosition() api.UserPosition {
	zero := sdk.ZeroInt()
	return api.UserPosition{
		Shares:      zero,
		Staked:      zero,
		OutPerShare: zero,
		Spent:       zero,
		Purchased:   zero,
	}
}
