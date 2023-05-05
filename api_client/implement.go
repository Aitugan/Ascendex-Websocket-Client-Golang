package api_client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type AscendExClient struct {
	Conn        *websocket.Conn
	Subscribed  bool
	Symbol      string
	StopChannel chan bool
}

func NewAscendExClient() *AscendExClient {
	return &AscendExClient{}
}

func (c *AscendExClient) Connection() error {
	u := url.URL{Scheme: "wss", Host: "ascendex.com", Path: "/api/pro/v1/stream"}
	header := http.Header{"Content-Type": []string{"application/json"}}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %s", err)
	}
	_, welcome, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	fmt.Println("Received welcome message:", string(welcome))

	c.Conn = conn
	return nil
}

func (c *AscendExClient) Disconnect() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	if c.StopChannel != nil {
		c.StopChannel <- true
	}
}

func (c *AscendExClient) SubscribeToChannel(symbol string) error {
	if c.Subscribed {
		return fmt.Errorf("already subscribed to a channel")
	}

	c.Symbol = symbol
	subscribeMessage := fmt.Sprintf(`{"op":"sub","ch": "bbo:%s","args":["spot/orderBook:%s"]}`, symbol, symbol)

	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(subscribeMessage))
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel: %s", err)
	}

	c.Subscribed = true
	return nil
}

func (c *AscendExClient) ReadMessagesFromChannel(ch chan<- BestOrderBook) {
	if c.Conn == nil {
		return
	}
	defer close(ch)

	for {
		select {
		case <-c.StopChannel:
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				fmt.Printf("error reading message from websocket: %s\n", err)
				c.Disconnect()
				return
			}
			var data map[string]interface{}
			err = json.Unmarshal(message, &data)
			if err != nil {
				fmt.Printf("error unmarshalling message from websocket: %s\n", err)
				continue
			}

			if data["m"] != "bbo" {
				continue
			}
			var dataResp struct {
				Symbol string `json:"symbol"`
				Data   struct {
					Ts  int64    `json:"ts"`
					Bid []string `json:"bid"`
					Ask []string `json:"ask"`
				} `json:"data"`
			}
			if err := json.Unmarshal(message, &dataResp); err != nil {
				fmt.Printf("error while unmarshalling book: %v", err)
				return
			}

			bidPrice, err := strconv.ParseFloat(dataResp.Data.Bid[0], 64)
			if err != nil {
				log.Println("error converting to float")
				continue
			}
			bidAmount, err := strconv.ParseFloat(dataResp.Data.Bid[1], 64)
			if err != nil {
				log.Println("error converting to float")
				continue
			}

			askPrice, err := strconv.ParseFloat(dataResp.Data.Ask[0], 64)
			if err != nil {
				log.Println("error converting to float")
				continue
			}
			askAmount, err := strconv.ParseFloat(dataResp.Data.Ask[1], 64)
			if err != nil {
				log.Println("error converting to float")
				continue
			}

			book := BestOrderBook{
				Bid: Order{
					Price:  bidPrice,
					Amount: bidAmount,
				},

				Ask: Order{
					Price:  askPrice,
					Amount: askAmount,
				},
			}
			select {
			case ch <- book:
			default:
				fmt.Println("channel full")
			}
		}
	}
}

func (c *AscendExClient) WriteMessagesToChannel() {
	for {
		fmt.Println("Wrote to channel")
		err := c.Conn.WriteMessage(websocket.PingMessage, []byte{})
		if err != nil {
			return
		}

		time.Sleep(5 * time.Second)
	}
}
