package intervention

import (
	"log"
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
	maxSellLoopCount         uint64
	buySemaphore             chan struct{}

	tokenCache sync.Map
}

func NewBot(
	bobpad *bobpad.BobPad,
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
	buyDeadline := buyTime.Add(b.maxAgeAfterBuy)

	var followerDeadline time.Time
	hasFollower := false

	for {
		time.Sleep(b.sellInterval)

		// Check absolute buy age
		if time.Now().After(buyDeadline) {
			log.Printf("Max age after buy reached for token %d, selling", tokenID)
			b.sell(tokenID, amountBought)
			return
		}

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

		// tokenData, err := b.launchpad.GetTokenData(tokenID)
		// if err != nil {
		// 	log.Printf("Failed to get token data of token with ID %d: %s", tokenID, err)
		// 	continue
		// }

		// if i == 0 && !tokenData.FirstOrderFrom.Equal(b.launchpad.agent.Sender()) {
		// 	log.Printf("We are not the first for token with ID %d, we are gonna sell", tokenID)
		// 	icpAmountOut, err := b.launchpad.Sell(tokenID, botBalance)
		// 	if err != nil {
		// 		log.Printf("Failed to sell token with ID %d", tokenID)
		// 		continue
		// 	} else {
		// 		operationClosedEvent := NewEventPositionClosed(
		// 			tokenID,
		// 			fmt.Sprintf("https://launch.bob.fun/coin/?id=%d", tokenID),
		// 			icpAmountOut,
		// 			amountBought,
		// 			int64(icpAmountOut)-int64(b.commitment),
		// 			"not-first",
		// 		)

		// 		b.eventPublisher.PublishEvent(operationClosedEvent)
		// 		return
		// 	}

		// }

		// Check reserve threshold for early sell
		reserve, err := b.bobpad.TokenReserve(tokenID)
		if err != nil {
			log.Println(err)
			continue
		}
		minAmountOut := b.minAmountOut
		if token.IsPremium() {
			minAmountOut = uint64(float64(minAmountOut) * 1.5)
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
