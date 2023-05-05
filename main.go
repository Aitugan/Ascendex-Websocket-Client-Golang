package main

import (
	"fmt"
	"strings"

	client "github.com/Aitugan/ascendex-client/api_client"
)

func main() {
	ascendex := client.NewAscendExClient()
	err := ascendex.Connection()
	if err != nil {
		fmt.Printf("error connecting to AscendEX WebSocket: %s\n", err)
		return
	}
	symbol := strings.Replace("BTC_USDT", "_", "/", 1)
	err = ascendex.SubscribeToChannel(symbol)
	if err != nil {
		fmt.Printf("error subscribing to BTC/USDT channel: %s\n", err)
		return
	}
	bboCh := make(chan client.BestOrderBook, 10)
	go ascendex.ReadMessagesFromChannel(bboCh)

	go func() {
		for bbo := range bboCh {
			fmt.Printf("Bid: %f %f | Ask: %f %f\n", bbo.Bid.Price, bbo.Bid.Amount, bbo.Ask.Price, bbo.Ask.Amount)
		}
	}()

	ascendex.WriteMessagesToChannel()

	ascendex.Disconnect()
}
