package bobpad

import (
	"errors"
	"fmt"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func (b *BobPad) DepositBob(amount8s uint64) error {
	var depositResp ApiVariantResponse[idl.Nat]

	if err := b.agent.Call(
		b.canisterID,
		"deposit_bob",
		[]any{amount8s},
		[]any{&depositResp},
	); err != nil {
		return fmt.Errorf("failed to deposit: %w", err)
	}

	if depositResp.Err != nil {
		return errors.New(*depositResp.Err)
	}

	return nil
}
