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
	decaySteepness           = 3
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
		decaySteepness,
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
}
