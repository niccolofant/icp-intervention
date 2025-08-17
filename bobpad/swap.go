package bobpad

import (
	"errors"
	"fmt"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func (b *BobPad) Swap(tokenId uint64, side string, amount8s uint64) (uint64, error) {
	var resp ApiVariantResponse[swapRespOk]

	params := swapParams{
		Side: idl.Variant{
			Name: side,
		},
		TokenID:  tokenId,
		Amount8s: amount8s,
	}

	if err := b.agent.Call(
		b.canisterID,
		"swap",
		[]any{params},
		[]any{&resp},
	); err != nil {
		return 0, fmt.Errorf("failed to swap: %w", err)
	}

	if resp.Err != nil {
		return 0, errors.New(*resp.Err)
	}

	return resp.Ok.AmountOut.BigInt().Uint64(), nil
}

type swapRespOk struct {
	Fee       idl.Nat `ic:"fee"`
	AmountOut idl.Nat `ic:"amount_out"`
	AmountIn  idl.Nat `ic:"amount_in"`
}

type swapParams struct {
	Side     idl.Variant `ic:"side"`
	TokenID  uint64      `ic:"token_id"`
	Amount8s uint64      `ic:"amount_e8s"`
}
