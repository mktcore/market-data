package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxResponseBodyBytes = 64 * 1024

// Client fetches Binance market data using a caller-owned HTTP client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// TickerPrice preserves the price's decimal representation without rounding.
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// HTTPError describes a non-2xx response. Code is nil when unavailable.
type HTTPError struct {
	StatusCode int
	Code       *int
	Message    string
}

func (e *HTTPError) Error() string {
	if e.Code != nil {
		return fmt.Sprintf("binance: HTTP %d, code %d: %s", e.StatusCode, *e.Code, e.Message)
	}
	return fmt.Sprintf("binance: HTTP %d: %s", e.StatusCode, e.Message)
}

// NewClient retains httpClient without changing its configuration.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("binance: HTTP client is required")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("binance: parse base URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, fmt.Errorf("binance: base URL must have an HTTP(S) scheme and host")
	}
	return &Client{baseURL: u, httpClient: httpClient}, nil
}

// TickerPrice fetches the latest price for one symbol using the caller's context.
func (c *Client) TickerPrice(ctx context.Context, symbol string) (TickerPrice, error) {

	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return TickerPrice{}, fmt.Errorf("binance: symbol is required")
	}

	u := c.baseURL.ResolveReference(&url.URL{Path: "/api/v3/ticker/price"})
	u.RawQuery = url.Values{"symbol": {symbol}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("binance: create ticker price request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("binance: request ticker price: %w", err)
	}
	defer resp.Body.Close()

	// The extra byte distinguishes a body at the limit from an oversized body.
	// Oversized bodies are closed without an unbounded drain.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes+1))
	if err != nil {
		return TickerPrice{}, fmt.Errorf("binance: read ticker price response (HTTP %d): %w", resp.StatusCode, err)
	}
	if len(body) > maxResponseBodyBytes {
		return TickerPrice{}, fmt.Errorf("binance: ticker price response (HTTP %d) exceeds %d bytes", resp.StatusCode, maxResponseBodyBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httpErr := &HTTPError{StatusCode: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
		var payload struct {
			Code    *int   `json:"code"`
			Message string `json:"msg"`
		}
		if json.Unmarshal(body, &payload) == nil {
			httpErr.Code = payload.Code
			if payload.Message != "" {
				httpErr.Message = payload.Message
			}
		}
		return TickerPrice{}, httpErr
	}

	var ticker TickerPrice
	if err := json.Unmarshal(body, &ticker); err != nil {
		return TickerPrice{}, fmt.Errorf("binance: decode ticker price response: %w", err)
	}
	if ticker.Symbol != symbol {
		return TickerPrice{}, fmt.Errorf("binance: response symbol %q does not match requested symbol %q", ticker.Symbol, symbol)
	}
	if strings.TrimSpace(ticker.Price) == "" {
		return TickerPrice{}, fmt.Errorf("binance: response price is required")
	}
	return ticker, nil
}
