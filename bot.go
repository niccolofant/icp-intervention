package intervention

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/niccolofant/intervention/bobpad"
)

type Bot struct {
	bobpad                   *bobpad.BobPad
	amountIn                 uint64
	minAmountOut             uint64
	buyInterval              time.Duration
	sellInterval             time.Duration
	maxAgeAfterBuy           time.Duration
	maxAgeAfterFirstFollower time.Duration
	decaySteepness           float64
	maxSellLoopCount         uint64
	buySemaphore             chan struct{}

	tokenCache sync.Map
}

func NewBot(
	bobpad *bobpad.BobPad,
	decaySteepness float64,
	commitment uint64,
	reserveSellThreshold uint64,
	buyInterval time.Duration,
	sellInterval time.Duration,
	maxAgeAfterBuy time.Duration,
	maxAgeAfterFirstFollower time.Duration,
	maxSellLoopCount uint64,
	buySemaphore chan struct{},
) *Bot {
	return &Bot{
		bobpad:                   bobpad,
		decaySteepness:           decaySteepness,
		amountIn:                 commitment,
		minAmountOut:             reserveSellThreshold,
		buyInterval:              buyInterval,
		sellInterval:             sellInterval,
		maxAgeAfterBuy:           maxAgeAfterBuy,
		maxAgeAfterFirstFollower: maxAgeAfterFirstFollower,
		maxSellLoopCount:         maxSellLoopCount,
		buySemaphore:             buySemaphore,
	}
}

func (b *Bot) Start() {
	ticker := time.NewTicker(b.buyInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.buySemaphore <- struct{}{}
		go func() {
			defer func() { <-b.buySemaphore }()
			b.tryToBuyNextToken()
		}()
	}
}

func (b *Bot) tryToBuyNextToken() {
	latestToken, err := b.bobpad.GetLatestToken()
	if err != nil {
		log.Printf("Failed to retrieve latest token: %s", err)
		return
	}

	if !latestToken.IsWorthy(3 * time.Second) {
		return
	}

	// Check cache: skip if already bought
	if _, loaded := b.tokenCache.LoadOrStore(latestToken.TokenInfo.TokenID, struct{}{}); loaded {
		// token already bought, skip
		return
	}

	amountBought, err := b.bobpad.Swap(latestToken.TokenInfo.TokenID, "Buy", b.amountIn)
	if err != nil {
		b.tokenCache.Delete(latestToken.TokenInfo.TokenID)
		return
	}

	// eventOperationOpened := NewEventPositionOpened(
	// 	nextTokenID,
	// 	fmt.Sprintf("https://launch.bob.fun/coin/?id=%d", nextTokenID),
	// 	amountBought,
	// 	b.commitment,
	// )

	//b.eventPublisher.PublishEvent(eventOperationOpened)
	log.Printf("Bought %d", latestToken.TokenInfo.TokenID)

	b.tryToSellToken(latestToken, amountBought)
}

func (b *Bot) tryToSellToken(token bobpad.Token, amountBought uint64) {
	tokenID := token.TokenInfo.TokenID

	buyTime := time.Now()

	var followerDeadline time.Time
	hasFollower := false

	for {
		time.Sleep(b.sellInterval)

		// Get balance
		botBalance, err := b.bobpad.TokenBalance(tokenID)
		if err != nil {
			log.Printf("Failed to get bot balance of token with ID %d: %s", tokenID, err)
			continue
		}
		if botBalance == 0 {
			log.Printf("Bot has balance 0 for token %d", tokenID)
			b.tokenCache.Delete(tokenID)
			return
		}

		minAmountOut := b.minAmountOut
		if token.IsPremium() {
			minAmountOut = uint64(float64(minAmountOut) * 1.5)
		}

		// Check reserve threshold for early sell
		reserve, err := b.bobpad.TokenReserve(tokenID)
		if err != nil {
			log.Println(err)
			continue
		}

		dynamicDeadline := buyTime.Add(b.dynamicMaxAge(reserve, minAmountOut))

		// Check absolute buy age
		if time.Now().After(dynamicDeadline) {
			log.Printf("Dynamic max age reached for token %d, selling", tokenID)
			b.sell(tokenID, amountBought)
			return
		}

		if !hasFollower {
			latestOrders := uint64(1)
			orders, err := b.bobpad.GetOrders(tokenID, &latestOrders)
			if err != nil {
				log.Printf("Failed to get orders for token %d: %s", tokenID, err)
				continue
			}

			if len(orders) > 0 {
				// Find our first buy
				order := orders[0]
				if order.Op.Swap != nil && order.Op.Swap.Side.Buy != nil && order.Op.Swap.From.String() != "7veju-2fuqc-osq44-pc2kj-hsr5j-xshba-szn5o-mgfnu-cwvgx-r54be-jqe" {
					followerDeadline = time.Now().Add(b.maxAgeAfterFirstFollower)
					hasFollower = true
					log.Printf("Follower detected for token %d, will sell after %s", tokenID, b.maxAgeAfterFirstFollower)

				}
			}
		} else {
			// If follower countdown is active, check deadline
			if time.Now().After(followerDeadline) {
				log.Printf("Follower deadline reached for token %d, selling", tokenID)
				b.sell(tokenID, botBalance)
				return
			}
		}

		if reserve >= minAmountOut {
			log.Printf("Sell threshold reached for token %d", tokenID)
			b.sell(tokenID, botBalance)
			return
		}
	}
}

func (b *Bot) sell(tokenID uint64, amount uint64) {
	icpAmountOut, err := b.bobpad.Swap(tokenID, "Sell", amount)
	if err != nil {
		log.Printf("Failed to sell token %d", tokenID)
		return
	}
	log.Printf("Sold token %d for %d ICP", tokenID, icpAmountOut)
	b.tokenCache.Delete(tokenID)
}

// dynamicMaxAge adjusts the max age after buy depending on reserve progress.
// - If reserve is far from threshold -> maxAgeAfterBuy is long.
// - If reserve is close to threshold -> maxAgeAfterBuy shrinks.
func (b *Bot) dynamicMaxAge(reserve, minAmountOut uint64) time.Duration {
	if minAmountOut == 0 {
		return b.maxAgeAfterBuy // avoid division by zero
	}

	progress := float64(reserve) / float64(minAmountOut)
	if progress > 1 {
		progress = 1
	}

	// Parameters
	minDeadline := 15 * time.Minute
	maxDeadline := b.maxAgeAfterBuy
	k := b.decaySteepness

	// Exponential decay: 1 - e^(-k*progress)
	decay := 1 - math.Exp(-k*progress)

	// Interpolate between maxDeadline and minDeadline
	adjusted := time.Duration(float64(maxDeadline)*(1-decay) + float64(minDeadline)*decay)

	return adjusted
}
