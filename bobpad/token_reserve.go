package bobpad

import (
	"fmt"
)

func (b *BobPad) TokenReserve(tokenId uint64) (uint64, error) {
	var resp []*poolType

	if err := b.agent.Query(
		b.canisterID,
		"get_pools",
		[]any{[]uint64{tokenId}},
		[]any{&resp},
	); err != nil {
		return 0, fmt.Errorf("failed to get token reserve: %w", err)
	}

	for _, tb := range resp {
		if tb.Launchpad != nil {
			return tb.Launchpad.ReserveBase, nil
		}
	}

	return 0, fmt.Errorf("token with ID %d not found", tokenId)
}

type TokenReserve struct {
	Bob        uint64
	OtherToken uint64
}

type poolType struct {
	AMM       *AMMPool       `ic:"AMM,variant"`
	Launchpad *LaunchpadPool `ic:"Launchpad,variant"`
}

type AMMPool struct {
	TokenA         uint64 `ic:"token_a"`
	TokenB         uint64 `ic:"token_b"`
	TokenID        uint64 `ic:"token_id"`
	TotalLiquidity uint64 `ic:"total_liquidity"`
}

type LaunchpadPool struct {
	TokenID      uint64 `ic:"token_id"`
	ReserveBase  uint64 `ic:"reserve_base_token"`
	ReserveQuote uint64 `ic:"reserve_quote_token"`
	TotalSupply  uint64 `ic:"total_supply"`
}
