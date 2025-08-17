package bobpad

import (
	"fmt"
	"time"

	"github.com/aviate-labs/agent-go/principal"
)

func (b *BobPad) GetOrders(tokenId uint64, latest *uint64) ([]*Order, error) {
	var resp []getOrdersResp

	args := []any{
		struct {
			TokenID uint64  `ic:"token_id"`
			Latest  *uint64 `ic:"latest,omitempty"`
		}{
			TokenID: tokenId,
			Latest:  latest,
		},
	}

	if err := b.agent.Query(
		b.canisterID,
		"get_orders",
		args,
		[]any{&resp},
	); err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]*Order, len(resp))
	for i, o := range resp {
		orders[i] = &Order{
			Op:        o.Op,
			CreatedAt: time.Unix(0, int64(o.CreatedAt)*int64(time.Nanosecond)),
		}
	}

	return orders, nil
}

type getOrdersResp struct {
	Op        OperationType `ic:"op"`
	CreatedAt uint64        `ic:"created_at"`
}

// OperationType can be one of several variants
type OperationType struct {
	Swap              *SwapOrder              `ic:"Swap,variant"`
	ProvideLiquidity  *ProvideLiquidityOrder  `ic:"ProvideLiquidity,variant"`
	WithdrawLiquidity *WithdrawLiquidityOrder `ic:"WithdrawLiquidity,variant"`
	Transfer          *TransferOrder          `ic:"Transfer,variant"`
}

type orderSide struct {
	Buy  *struct{} `ic:"Buy,variant"`
	Sell *struct{} `ic:"Sell,variant"`
}

// Each possible order type
type SwapOrder struct {
	Fee      uint64              `ic:"fee"`
	TokenID  uint64              `ic:"token_id"`
	TokenIn  uint64              `ic:"token_in"`
	From     principal.Principal `ic:"from"`
	Side     orderSide           `ic:"side"`
	TokenOut uint64              `ic:"token_out"`
}

type ProvideLiquidityOrder struct {
	TokenAmount uint64              `ic:"token_amount"`
	TokenID     uint64              `ic:"token_id"`
	BobAmount   uint64              `ic:"bob_amount"`
	From        principal.Principal `ic:"from"`
}

type WithdrawLiquidityOrder struct {
	TokenBOut uint64              `ic:"token_b_out"`
	Shares    uint64              `ic:"shares"`
	TokenID   uint64              `ic:"token_id"`
	From      principal.Principal `ic:"from"`
	TokenAOut uint64              `ic:"token_a_out"`
}

type TransferOrder struct {
	To      principal.Principal `ic:"to"`
	TokenID uint64              `ic:"token_id"`
	From    principal.Principal `ic:"from"`
	Amount  uint64              `ic:"amount"`
}

type Order struct {
	Op        OperationType
	CreatedAt time.Time
}
