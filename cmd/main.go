package main

import (
	"time"

	"github.com/niccolofant/intervention"
	"github.com/niccolofant/intervention/bobpad"
)

const (
	commitment               = uint64(300e8)   // 200 BOB
	reserveSellThreshold     = uint64(450e8)   // 300 BOB
	buyInterval              = 1 * time.Second // 1 sec
	sellInterval             = 2 * time.Second // 1 sec
	MaxAgeAfterBuy           = 5 * time.Hour
	MaxAgeAfterFirstFollower = 30 * time.Minute
	maxSellLoopCount         = 1000
	maxConcurrentBuyTasks    = 100
)

func main() {
	id, err := intervention.LoadIntentity()
	if err != nil {
		panic(err)
	}

	agent, err := intervention.GetAgent(id)
	if err != nil {
		panic(err)
	}

	bobPad := bobpad.New(agent)

	bot := intervention.NewBot(
		bobPad,
		commitment,
		reserveSellThreshold,
		buyInterval,
		sellInterval,
		MaxAgeAfterBuy,
		MaxAgeAfterFirstFollower,
		maxSellLoopCount,
		make(chan struct{}, maxConcurrentBuyTasks),
	)

	go bot.Start()

	select {}

	// err = bobPad.DepositBob(90e8)
	// if err != nil {
	// 	panic(err)
	// }

	// latestToken, err := bobPad.GetLatestToken()
	// if err != nil {
	// 	panic(err)
	// }

	// amountOut, err := bobPad.Swap(10979592588320680705, "Buy", 100000000)
	// if err != nil {
	// 	panic(err)
	// }

	// amountOut, err := bobPad.Swap(10979592588320680705, "Sell", 1285685002173)
	// if err != nil {
	// 	panic(err)
	// }

	// log.Println(amountOut)
}
