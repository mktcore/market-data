package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mktcore/market-data/internal/binance"
)

func main() {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	client, err := binance.NewClient("https://data-api.binance.vision", httpClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ticker, err := client.TickerPrice(context.Background(), "BTCUSDT")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s %s\n", ticker.Symbol, ticker.Price)
}
