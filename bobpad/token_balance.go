package bobpad

import (
	"fmt"
)

func (b *BobPad) TokenBalance(tokenId uint64) (uint64, error) {
	var resp []tokenBalance

	if err := b.agent.Query(
		b.canisterID,
		"get_balances",
		[]any{b.agent.Sender()},
		[]any{&resp},
	); err != nil {
		return 0, fmt.Errorf("failed to get token balance: %w", err)
	}

	for _, tb := range resp {
		if tb.TokenID == tokenId {
			return tb.Balance, nil
		}
	}

	return 0, nil
}

type tokenBalance struct {
	TokenID uint64 `ic:"token_id"`
	Balance uint64 `ic:"balance"`
}
