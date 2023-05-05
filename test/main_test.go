package test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	apiClient "github.com/Aitugan/ascendex-client/api_client"
)

func TestConnection(t *testing.T) {
	client := apiClient.NewAscendExClient()

	err := client.Connection()
	if err != nil {
		t.Errorf("Connection failed: %v", err)
	}

	client.Disconnect()
}

func TestSubscribeToChannel(t *testing.T) {
	ascendex := apiClient.NewAscendExClient()
	err := ascendex.Connection()
	if err != nil {
		fmt.Printf("error connecting to AscendEX WebSocket: %s\n", err)
		return
	}
	symbol := strings.Replace("BTC_USDT", "_", "/", 1)
	if err := ascendex.SubscribeToChannel(symbol); err != nil {
		t.Errorf("Failed to subscribe to channel BTC/USDT: %v", err)
	}

	ascendex.Disconnect()
}

func TestReadMessagesFromChannel(t *testing.T) {
	ascendex := apiClient.NewAscendExClient()
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
	bboCh := make(chan apiClient.BestOrderBook, 10)
	go ascendex.ReadMessagesFromChannel(bboCh)

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	case book := <-bboCh:
		if reflect.DeepEqual(book, apiClient.BestOrderBook{}) {
			t.Error("expected some value, got nil")
		}
	}
}

// not fully implemented
func TestWriteMessagesToChannel(t *testing.T) {
	ascendex := apiClient.NewAscendExClient()
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

	go ascendex.WriteMessagesToChannel()

	for i := 0; i < 3; i++ {
		_, message, err := ascendex.Conn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if string(message) != "" {
			t.Errorf("expected empty message, got '%s'", message)
		}
		time.Sleep(5 * time.Second)
	}

	ascendex.StopChannel <- true
}
