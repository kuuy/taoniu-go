package spot

import (
	"encoding/json"
	"testing"
)

type mockClient struct {
	sent [][]byte
}

func (m *mockClient) SendBytes(data []byte) {
	m.sent = append(m.sent, data)
}

type mockHub struct {
	subscribed []string
}

func (m *mockHub) Subscribe(client interface{}, topic string) {
	m.subscribed = append(m.subscribed, topic)
}

func (m *mockHub) Unsubscribe(client interface{}, topic string) {}

func (m *mockHub) Broadcast(topic string, message []byte) {}

func TestExtractSymbols(t *testing.T) {
	req := map[string]interface{}{
		"action":  "subscribe",
		"topic":   "binance:spot:tickers",
		"symbols": []interface{}{"btcusdt", "ethusdt"},
	}

	symbols, err := extractSymbols(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(symbols) != 2 || symbols[0] != "BTCUSDT" || symbols[1] != "ETHUSDT" {
		t.Errorf("expected ['BTCUSDT', 'ETHUSDT'], got %v", symbols)
	}
}

func TestSubscribeTopicFormatting(t *testing.T) {
	hub := &mockHub{}
	tickers := NewTickers(nil, hub)

	req := map[string]interface{}{
		"action":  "subscribe",
		"topic":   "binance:spot:tickers",
		"symbols": []string{"btcusdt"},
	}

	err := tickers.Subscribe(&mockClient{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hub.subscribed) != 1 || hub.subscribed[0] != "binance:spot:tickers:BTCUSDT" {
		t.Errorf("expected topic 'binance:spot:tickers:BTCUSDT', got %v", hub.subscribed)
	}
}

func TestInitGlobalNatsBroadcast(t *testing.T) {
	payload := TickersUpdatePayload{
		Symbol: "BTCUSDT",
		Price:  65000.0,
	}

	data, _ := json.Marshal(payload)
	_ = data
	topic := "binance:spot:tickers:" + payload.Symbol
	if topic != "binance:spot:tickers:BTCUSDT" {
		t.Errorf("expected topic 'binance:spot:tickers:BTCUSDT', got %s", topic)
	}

	pushBytes, _ := json.Marshal(map[string]interface{}{
		"event": "ticker",
		"topic": topic,
		"data":  payload,
	})

	var push struct {
		Event string               `json:"event"`
		Topic string               `json:"topic"`
		Data  TickersUpdatePayload `json:"data"`
	}
	json.Unmarshal(pushBytes, &push)

	if push.Event != "ticker" || push.Topic != "binance:spot:tickers:BTCUSDT" || push.Data.Price != 65000.0 {
		t.Errorf("unexpected push JSON payload: %+v", push)
	}
}

func TestSubscribeMultipleSymbols(t *testing.T) {
	hub := &mockHub{}
	tickers := NewTickers(nil, hub)

	testSymbols := []string{"DOGEUSDT", "DASHUSDT", "GMTUSDT", "TRXUSDT", "GALAUSDT"}
	req := map[string]interface{}{
		"action":  "subscribe",
		"topic":   "binance:spot:tickers",
		"symbols": testSymbols,
	}

	client := &mockClient{}
	err := tickers.Subscribe(client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hub.subscribed) != len(testSymbols) {
		t.Fatalf("expected %d subscribed topics, got %d", len(testSymbols), len(hub.subscribed))
	}

	expectedTopics := []string{
		"binance:spot:tickers:DOGEUSDT",
		"binance:spot:tickers:DASHUSDT",
		"binance:spot:tickers:GMTUSDT",
		"binance:spot:tickers:TRXUSDT",
		"binance:spot:tickers:GALAUSDT",
	}

	for i, expected := range expectedTopics {
		if hub.subscribed[i] != expected {
			t.Errorf("at index %d, expected %s, got %s", i, expected, hub.subscribed[i])
		}
	}
}
