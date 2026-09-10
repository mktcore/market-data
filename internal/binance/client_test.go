package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTickerPriceSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v3/ticker/price" {
			t.Errorf("path = %q, want /api/v3/ticker/price", r.URL.Path)
		}
		if r.URL.RawQuery != "symbol=BTCUSDT" {
			t.Errorf("query = %q, want symbol=BTCUSDT", r.URL.RawQuery)
		}
		fmt.Fprint(w, `{"symbol":"BTCUSDT","price":"12345.67000000","extra":true}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.TickerPrice(context.Background(), "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	want := TickerPrice{Symbol: "BTCUSDT", Price: "12345.67000000"}
	if got != want {
		t.Fatalf("ticker = %+v, want %+v", got, want)
	}
}

func TestTickerPriceEmptySymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("empty symbol must not send an HTTP request")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	for _, symbol := range []string{"", " \t"} {
		if _, err := client.TickerPrice(context.Background(), symbol); err == nil {
			t.Errorf("symbol %q: expected an error", symbol)
		}
	}
}

func TestTickerPriceHTTPError(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		code    int
		hasCode bool
		message string
	}{
		{"Binance JSON", 400, `{"code":-1121,"msg":"Invalid symbol."}`, -1121, true, "Invalid symbol."},
		{"non-JSON", 502, `<html>Bad Gateway</html>`, 0, false, "Bad Gateway"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			t.Cleanup(server.Close)
			client, err := NewClient(server.URL, server.Client())
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.TickerPrice(context.Background(), "BTCUSDT")
			var httpErr *HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("error = %v, want *HTTPError", err)
			}
			if httpErr.StatusCode != tt.status || httpErr.Message != tt.message {
				t.Errorf("HTTP error = %+v, want status %d and message %q", httpErr, tt.status, tt.message)
			}
			if tt.hasCode {
				if httpErr.Code == nil || *httpErr.Code != tt.code {
					t.Errorf("code = %v, want %d", httpErr.Code, tt.code)
				}
			} else if httpErr.Code != nil {
				t.Errorf("code = %d, want nil", *httpErr.Code)
			}
		})
	}
}

func TestTickerPriceInvalidResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		message string
	}{
		{"malformed JSON", `{"symbol":!}`, "decode ticker price response"},
		{"missing symbol", `{"price":"1.00"}`, "response symbol"},
		{"mismatched symbol", `{"symbol":"ETHUSDT","price":"1.00"}`, "response symbol"},
		{"missing price", `{"symbol":"BTCUSDT"}`, "response price is required"},
		{"empty price", `{"symbol":"BTCUSDT","price":""}`, "response price is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tt.body)
			}))
			t.Cleanup(server.Close)
			client, err := NewClient(server.URL, server.Client())
			if err != nil {
				t.Fatal(err)
			}
			got, err := client.TickerPrice(context.Background(), "BTCUSDT")
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("error = %v, want message containing %q", err, tt.message)
			}
			if got != (TickerPrice{}) {
				t.Errorf("ticker = %+v, want zero value on failure", got)
			}
			if tt.name == "malformed JSON" {
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Errorf("error = %v, want wrapped *json.SyntaxError", err)
				}
			}
		})
	}
}

func TestTickerPriceAlreadyCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("already-cancelled context must not send an HTTP request")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.TickerPrice(ctx, "BTCUSDT")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
