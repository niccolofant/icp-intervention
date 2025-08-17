package bobpad

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

var (
	blacklistedCreators = []string{
		"bdedb-mxdar-r6dzs-twggw-htds6-vts3b-m3s7z-gx7dc-mtapy-lizfx-zae",
		"gaxjt-o3xdb-4d5fu-px5tp-inpdn-dzpdm-hdfjf-6iu4c-gg6bj-7cpa7-fae",
	}
)

func (b *BobPad) GetLatestToken() (Token, error) {
	var resp []Token

	params := getTokensParams{
		SortBy: idl.Variant{
			Name: "CreatedAt", // or "LastTraded", "Bonded", "MarketCap"
		},
		Pagination: pagination{
			Skip:   0,
			Length: 100000,
		},
	}

	// Call the canister
	if err := b.agent.Query(
		b.canisterID,
		"get_tokens",
		[]any{params},
		[]any{&resp},
	); err != nil {
		return Token{}, fmt.Errorf("failed to get tokens: %w", err)
	}

	return resp[0], nil
}

type sortBy struct {
	CreatedAt  *struct{} `ic:"CreatedAt,variant"`
	LastTraded *struct{} `ic:"LastTraded,variant"`
	Bonded     *struct{} `ic:"Bonded,variant"`
	MarketCap  *struct{} `ic:"MarketCap,variant"`
}

type getTokensParams struct {
	SortBy     idl.Variant `ic:"sort_by"`
	Pagination pagination  `ic:"pagination"`
}

type pagination struct {
	Skip   uint64 `ic:"skip"`
	Length uint64 `ic:"length"`
}

type Token struct {
	PriceInBob     float64   `ic:"price_in_bob"`
	MarketCapInBob float64   `ic:"market_cap_in_bob"`
	HolderCount    uint64    `ic:"holder_count"`
	TokenInfo      tokenInfo `ic:"token_info"`
	TotalSupply    idl.Nat   `ic:"total_supply"`
}

func (t Token) IsWorthy(xSeconds time.Duration) bool {
	now := time.Now()
	cutoff := now.Add(-xSeconds)

	createdAt := time.Unix(int64(t.TokenInfo.CreatedAt), 0)

	name := strings.ToLower(t.TokenInfo.Name)
	ticker := strings.ToLower(t.TokenInfo.Ticker)

	return t.MarketCapInBob == 2109.2855470912427 &&
		createdAt.After(cutoff) &&
		!strings.Contains(name, "test") &&
		!strings.Contains(ticker, "test") &&
		!strings.Contains(ticker, "sniper") &&
		!strings.Contains(ticker, "bot") &&
		!isBlacklistedCreator(t.TokenInfo.CreatedBy.String())
}

func (t Token) IsPremium() bool {
	name := strings.ToLower(t.TokenInfo.Name)

	return strings.Contains(name, "bob") ||
		t.TokenInfo.MaybeTwitter != nil ||
		t.TokenInfo.MaybeTelegram != nil
}

func isBlacklistedCreator(creator string) bool {
	return slices.Contains(blacklistedCreators, creator)
}

type tokenInfo struct {
	MaybeWebsite  *string             `ic:"maybe_website,omitempty"`
	Ticker        string              `ic:"ticker"`
	TokenID       uint64              `ic:"token_id"`
	MaybeOC       *string             `ic:"maybe_oc,omitempty"`
	Name          string              `ic:"name"`
	Description   string              `ic:"description"`
	CreatedAt     uint64              `ic:"created_at"`
	CreatedBy     principal.Principal `ic:"created_by"`
	ImagePath     string              `ic:"image_path"`
	MaybeTwitter  *string             `ic:"maybe_twitter,omitempty"`
	MaybeTelegram *string             `ic:"maybe_telegram,omitempty"`
}
