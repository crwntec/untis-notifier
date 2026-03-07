package notifier

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessage(t *testing.T) {
	var received *http.Request
	var receivedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL})
	err := client.SendMessage(context.Background(), "test-topic", Message{
		Title:    "test",
		Text:     "test",
		Priority: 4,
		Tags:     []string{"test", "test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if received.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", received.Method)
	}
	if received.URL.Path != "/test-topic" {
		t.Errorf("path = %q, want %q", received.URL.Path, "/test-topic")
	}
	if got := received.Header.Get("Title"); got != "test" {
		t.Errorf("Title = %q, want %q", got, "test")
	}
	if got := received.Header.Get("Priority"); got != "4" {
		t.Errorf("Priority = %q, want %q", got, "4")
	}
	if got := received.Header.Get("Tags"); got != "test,test" {
		t.Errorf("Tags = %q, want %q", got, "test,test")
	}
	if got := string(receivedBody); got != "test" {
		t.Errorf("body = %q, want %q", got, "test")
	}
}
