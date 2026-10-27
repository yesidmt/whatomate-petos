package waha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendTextRejectsWrongSenderAndExcludedNumber(t *testing.T) {
	var sends int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sendText" {
			sends++
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "WORKING", "me": map[string]string{"id": ExcludedPhone + "@c.us"},
		})
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, APIKey: "test", Session: "temporary", ExpectedPhone: "573001112233", HTTP: server.Client()}
	if _, err := c.SendText(context.Background(), "573009998877", "hello", ""); err == nil {
		t.Fatal("expected sender mismatch to block send")
	}
	if _, err := c.SendText(context.Background(), ExcludedPhone, "hello", ""); err == nil {
		t.Fatal("expected excluded recipient to block send")
	}
	if sends != 0 {
		t.Fatalf("sent %d messages with a blocked phone", sends)
	}
}

func TestConfigurationRejectsExcludedSender(t *testing.T) {
	t.Setenv("PETOS_WAHA_URL", "http://127.0.0.1:3000")
	t.Setenv("PETOS_WAHA_API_KEY", "test-key")
	t.Setenv("PETOS_WAHA_SESSION", "temporary")
	t.Setenv("PETOS_WAHA_PHONE", "+57 310 455 8226")
	if _, err := NewFromEnv(); err == nil {
		t.Fatal("excluded sender was accepted")
	}
}

func TestSendTextUsesApprovedSessionAndAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Error("API key missing")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/sessions/temporary":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "WORKING", "me": map[string]string{"id": "573001112233@c.us"},
			})
		case "/api/sendText":
			var body struct {
				Session string `json:"session"`
				ChatID  string `json:"chatId"`
				Text    string `json:"text"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Session != "temporary" || body.ChatID != "573009998877@c.us" || body.Text != "hello" {
				t.Errorf("unexpected send payload: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "true_573009998877@c.us_ABC"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, APIKey: "test-key", Session: "temporary", ExpectedPhone: "573001112233", HTTP: server.Client()}
	id, err := c.SendText(context.Background(), "+57 300 999 8877", "hello", "")
	if err != nil || id != "true_573009998877@c.us_ABC" {
		t.Fatalf("unexpected result: id=%s err=%v", id, err)
	}
}
